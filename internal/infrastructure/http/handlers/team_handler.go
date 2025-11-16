package handlers

import (
    "encoding/json"
    "fmt"
    "net/http"
    "github.com/shmul/avito-task/internal/domain/entity"
    "github.com/shmul/avito-task/internal/domain/service"
    "github.com/shmul/avito-task/internal/infrastructure/http/dto"
)

type TeamHandler struct {
    teamService *service.TeamService
}

func NewTeamHandler(teamService *service.TeamService) *TeamHandler {
    return &TeamHandler{
        teamService: teamService,
    }
}

func (h *TeamHandler) AddTeam(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    var req dto.CreateTeamRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        sendError(w, "Invalid request body", "BAD_REQUEST", http.StatusBadRequest)
        return
    }

    // Convert DTO to domain entity
    team := &entity.Team{
        TeamName: req.TeamName,
        Members:  make([]entity.User, len(req.Members)),
    }

    for i, member := range req.Members {
        team.Members[i] = entity.User{
            UserID:   member.UserID,
            Username: member.Username,
            TeamName: req.TeamName,
            IsActive: member.IsActive,
        }
    }

    // Create team
    if err := h.teamService.CreateTeam(team); err != nil {
        if err.Error() == fmt.Sprintf("team already exists: %s", req.TeamName) {
            sendError(w, "team_name already exists", "TEAM_EXISTS", http.StatusBadRequest)
            return
        }
        sendError(w, err.Error(), "INTERNAL_ERROR", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(dto.TeamResponse{Team: team})
}

func (h *TeamHandler) GetTeam(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    teamName := r.URL.Query().Get("team_name")
    if teamName == "" {
        sendError(w, "team_name is required", "BAD_REQUEST", http.StatusBadRequest)
        return
    }

    team, err := h.teamService.GetTeam(teamName)
    if err != nil {
        if err.Error() == fmt.Sprintf("team not found: %s", teamName) {
            sendError(w, "resource not found", "NOT_FOUND", http.StatusNotFound)
            return
        }
        sendError(w, err.Error(), "INTERNAL_ERROR", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(team)
}