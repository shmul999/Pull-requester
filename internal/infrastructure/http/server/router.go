package server

import (
	"log/slog"
	"net/http"
	"github.com/shmul/avito-task/internal/domain/service"
	"github.com/shmul/avito-task/internal/infrastructure/http/handlers"
	"github.com/shmul/avito-task/internal/infrastructure/http/middleware"
)

type Router struct {
	teamHandler *handlers.TeamHandler
	userHandler *handlers.UserHandler
	prHandler   *handlers.PRHandler
	log         *slog.Logger
}

func NewRouter(userService *service.UserService, teamService *service.TeamService, prService *service.PRService, log *slog.Logger) *Router {
	return &Router{
		teamHandler: handlers.NewTeamHandler(teamService),
		userHandler: handlers.NewUserHandler(userService, prService),
		prHandler:   handlers.NewPRHandler(prService),
		log:         log,
	}
}

func (r *Router) SetupRoutes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/team/add", r.teamHandler.AddTeam)
	mux.HandleFunc("/team/get", r.teamHandler.GetTeam)

	mux.HandleFunc("/users/setIsActive", r.userHandler.SetUserActive)
	mux.HandleFunc("/users/getReview", r.userHandler.GetUserReview)

	mux.HandleFunc("/pullRequest/create", r.prHandler.CreatePR)
	mux.HandleFunc("/pullRequest/merge", r.prHandler.MergePR)
	mux.HandleFunc("/pullRequest/reassign", r.prHandler.ReassignReviewer)

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status": "ok"}`))
	})

	//middleware
	handler := middleware.CORS(mux)
	handler = middleware.Logging(r.log)(handler)
	handler = middleware.Recovery(r.log)(handler)

	return handler
}
