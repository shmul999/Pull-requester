package main

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"github.com/shmul/avito-task/config"
	"github.com/shmul/avito-task/internal/domain/service"
	"github.com/shmul/avito-task/internal/infrastructure/http/server"
	"github.com/shmul/avito-task/internal/infrastructure/storage/migrations"
	"github.com/shmul/avito-task/internal/infrastructure/storage/postgres"
)

//go:embed *.sql
var migrationsFS embed.FS

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	configPath := "./config/config.yaml"
    if env := os.Getenv("ENV"); env == "docker" {
        configPath = "./config/config.docker.yaml"
    }

	cfg := config.Load(configPath)
	log := SetupLogger(cfg.Env)
	waitForDB(cfg, log)

	log.Info("starting pull-requester", slog.String("env", cfg.Env))
	log.Debug("debug messages are enabled")

	log.Info("connecting to database...")
	db, err := postgres.NewConnection(cfg)
	if err != nil {
		log.Error("failed to connect to database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer db.Close()
	log.Info("successfully connected to database")

	log.Info("running database migrations...")
	if err := migrations.Run(db.DB(), migrationsFS); err != nil {
		log.Error("failed to run migrations", slog.String("error", err.Error()))
		os.Exit(1)
	}
	log.Info("migrations completed successfully")

	userRepo := postgres.NewUserRepository(db.DB())
	teamRepo := postgres.NewTeamRepository(db.DB())
	prRepo := postgres.NewPRRepository(db.DB())

	userService := service.NewUserService(userRepo, teamRepo)
	teamService := service.NewTeamService(teamRepo, userRepo)
	prService := service.NewPRService(prRepo, userRepo, teamRepo, &service.PRServiceConfig{
		ReviewerCount: cfg.App.ReviewerCount,
		RandomSeed:    int64(cfg.App.RandomSeed),
	})

	log.Info("initializing HTTP server...")
	router := server.NewRouter(userService, teamService, prService, log)
	handler := router.SetupRoutes()

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      handler,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	log.Info("starting HTTP server", slog.Int("port", cfg.Server.Port))
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("failed to start HTTP server", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	log.Info("server is running", slog.Int("port", cfg.Server.Port))

	<-quit
	log.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error("failed to shutdown server", slog.String("error", err.Error()))
	}

	log.Info("server stopped")
}

func SetupLogger(env string) *slog.Logger {
	var log *slog.Logger
	switch env {
	case envLocal:
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envDev:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envProd:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	default:
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}
	return log
}

func waitForDB(cfg *config.Config, log *slog.Logger) {
    if os.Getenv("ENV") != "docker" {
        return
    }
    
    dsn := fmt.Sprintf(
        "host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
        cfg.Database.Host,
        cfg.Database.Port,
        cfg.Database.User,
        cfg.Database.Password,
        cfg.Database.Name,
        cfg.Database.SSLMode,
    )
    
    for i := 0; i < 30; i++ {
        db, err := sql.Open("pgx", dsn)
        if err != nil {
            log.Info("Waiting for database...")
            time.Sleep(1 * time.Second)
            continue
        }
        
        if err := db.Ping(); err == nil {
            db.Close()
            log.Info("Database is ready!")
            return
        }
        db.Close()
        time.Sleep(1 * time.Second)
    }
    log.Error("Database connection timeout")
}