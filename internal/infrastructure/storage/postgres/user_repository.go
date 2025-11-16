package postgres

import (
    "database/sql"
    "fmt"
    "github.com/shmul/avito-task/internal/domain/entity"
    "github.com/shmul/avito-task/internal/domain/repo"
)

type UserRepository struct {
    db *sql.DB
}

func NewUserRepository(db *sql.DB) repo.UserRepository {
    return &UserRepository{db: db}
}

func (r *UserRepository) CreateOrUpdate(user *entity.User) error {
    query := `
        INSERT INTO users (user_id, username, team_name, is_active)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (user_id) 
        DO UPDATE SET 
            username = EXCLUDED.username,
            team_name = EXCLUDED.team_name,
            is_active = EXCLUDED.is_active,
            updated_at = CURRENT_TIMESTAMP
    `
    _, err := r.db.Exec(query, user.UserID, user.Username, user.TeamName, user.IsActive)
    return err
}

func (r *UserRepository) GetDB() *sql.DB {
	return r.db
}

func (r *UserRepository) GetByID(userID string) (*entity.User, error) {
    query := `
        SELECT user_id, username, team_name, is_active
        FROM users 
        WHERE user_id = $1
    `
    
    var user entity.User
    err := r.db.QueryRow(query, userID).Scan(
        &user.UserID,
        &user.Username, 
        &user.TeamName,
        &user.IsActive,
    )
    
    if err == sql.ErrNoRows {
        return nil, fmt.Errorf("user not found: %s", userID)
    }
    
    return &user, err
}

func (r *UserRepository) SetActive(userID string, isActive bool) (*entity.User, error) {
    query := `
        UPDATE users 
        SET is_active = $1, updated_at = CURRENT_TIMESTAMP
        WHERE user_id = $2
        RETURNING user_id, username, team_name, is_active
    `
    
    var user entity.User
    err := r.db.QueryRow(query, isActive, userID).Scan(
        &user.UserID,
        &user.Username,
        &user.TeamName, 
        &user.IsActive,
    )
    
    if err == sql.ErrNoRows {
        return nil, fmt.Errorf("user not found: %s", userID)
    }
    
    return &user, err
}

func (r *UserRepository) GetActiveUsersByTeam(teamName string) ([]*entity.User, error) {
    query := `
        SELECT user_id, username, team_name, is_active
        FROM users 
        WHERE team_name = $1 AND is_active = true
        ORDER BY user_id
    `
    
    rows, err := r.db.Query(query, teamName)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var users []*entity.User
    for rows.Next() {
        var user entity.User
        if err := rows.Scan(
            &user.UserID,
            &user.Username,
            &user.TeamName,
            &user.IsActive,
        ); err != nil {
            return nil, err
        }
        users = append(users, &user)
    }
    
    return users, rows.Err()
}

func (r *UserRepository) GetByTeam(teamName string) ([]*entity.User, error) {
    query := `
        SELECT user_id, username, team_name, is_active
        FROM users 
        WHERE team_name = $1
        ORDER BY user_id
    `
    
    rows, err := r.db.Query(query, teamName)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var users []*entity.User
    for rows.Next() {
        var user entity.User
        if err := rows.Scan(
            &user.UserID,
            &user.Username,
            &user.TeamName,
            &user.IsActive,
        ); err != nil {
            return nil, err
        }
        users = append(users, &user)
    }
    
    return users, rows.Err()
}

func (r *UserRepository) Exists(userID string) (bool, error) {
    query := `SELECT EXISTS(SELECT 1 FROM users WHERE user_id = $1)`
    
    var exists bool
    err := r.db.QueryRow(query, userID).Scan(&exists)
    return exists, err
}