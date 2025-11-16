// internal/infrastructure/storage/postgres/team_repository.go
package postgres

import (
    "database/sql"
    "fmt"
    "github.com/shmul/avito-task/internal/domain/entity"
    "github.com/shmul/avito-task/internal/domain/repo"
)

type TeamRepository struct {
    db *sql.DB
}

func NewTeamRepository(db *sql.DB) repo.TeamRepository {
    return &TeamRepository{db: db}
}

//для тестов
// func (r *TeamRepository) GetDB() *sql.DB {
// 	return r.db
// }
//

func (r *TeamRepository) Create(team *entity.Team) error {
    tx, err := r.db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()

    _, err = tx.Exec("INSERT INTO teams (team_name) VALUES ($1) ON CONFLICT (team_name) DO NOTHING", team.TeamName)
    if err != nil {
        return fmt.Errorf("failed to create team: %w", err)
    }

    for _, member := range team.Members {
        _, err = tx.Exec(`
            INSERT INTO users (user_id, username, team_name, is_active)
            VALUES ($1, $2, $3, $4)
            ON CONFLICT (user_id) 
            DO UPDATE SET 
                username = EXCLUDED.username,
                team_name = EXCLUDED.team_name,
                is_active = EXCLUDED.is_active,
                updated_at = CURRENT_TIMESTAMP
        `, member.UserID, member.Username, team.TeamName, member.IsActive)
        if err != nil {
            return fmt.Errorf("failed to create user %s: %w", member.UserID, err)
        }
    }

    return tx.Commit()
}

func (r *TeamRepository) GetByName(teamName string) (*entity.Team, error) {
    var exists bool
    err := r.db.QueryRow("SELECT EXISTS(SELECT 1 FROM teams WHERE team_name = $1)", teamName).Scan(&exists)
    if err != nil {
        return nil, fmt.Errorf("failed to check team existence: %w", err)
    }
    if !exists {
        return nil, fmt.Errorf("team not found: %s", teamName)
    }

    rows, err := r.db.Query(`
        SELECT user_id, username, team_name, is_active
        FROM users 
        WHERE team_name = $1
        ORDER BY user_id
    `, teamName)
    if err != nil {
        return nil, fmt.Errorf("failed to get team members: %w", err)
    }
    defer rows.Close()

    var members []entity.User
    for rows.Next() {
        var user entity.User
        if err := rows.Scan(
            &user.UserID,
            &user.Username,
            &user.TeamName,
            &user.IsActive,
        ); err != nil {
            return nil, fmt.Errorf("failed to scan user: %w", err)
        }
        members = append(members, user)
    }

    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("error iterating users: %w", err)
    }

    return &entity.Team{
        TeamName: teamName,
        Members:  members,
    }, nil
}

func (r *TeamRepository) Exists(teamName string) (bool, error) {
    query := `SELECT EXISTS(SELECT 1 FROM teams WHERE team_name = $1)`
    
    var exists bool
    err := r.db.QueryRow(query, teamName).Scan(&exists)
    return exists, err
}