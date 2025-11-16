package postgres

import (
	"github.com/shmul/avito-task/config"
	"context"
	"database/sql"
	"fmt"
	"time"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type Storage struct{
	db *sql.DB
}

func NewConnection(cfg *config.Config) (*Storage, error){
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}

func (s *Storage) DB() *sql.DB {
	return s.db
}

func (s *Storage) HealthCheck(ctx context.Context) error {
	return s.db.PingContext(ctx)
}