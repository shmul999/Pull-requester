package postgres

import (
    "database/sql"
    "fmt"
    "time"
    "github.com/shmul/avito-task/internal/domain/entity"
    "github.com/shmul/avito-task/internal/domain/repo"
)

type PRRepository struct {
    db *sql.DB
}

func NewPRRepository(db *sql.DB) repo.PRRepository {
    return &PRRepository{db: db}
}

func (r *PRRepository) Create(pr *entity.PullRequest) error {
    tx, err := r.db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()

    var createdAt time.Time
    err = tx.QueryRow(`
        INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id, status)
        VALUES ($1, $2, $3, $4)
        RETURNING created_at
    `, pr.PullRequestID, pr.PullRequestName, pr.AuthorID, pr.Status).Scan(&createdAt)
    if err != nil {
        return fmt.Errorf("failed to create PR: %w", err)
    }

    pr.CreatedAt = &createdAt

    for _, reviewerID := range pr.AssignedReviewers {
        _, err = tx.Exec(`
            INSERT INTO pr_reviewers (pull_request_id, reviewer_id)
            VALUES ($1, $2)
            ON CONFLICT (pull_request_id, reviewer_id) DO NOTHING
        `, pr.PullRequestID, reviewerID)
        if err != nil {
            return fmt.Errorf("failed to assign reviewer %s: %w", reviewerID, err)
        }
    }

    return tx.Commit()
}

func (r *PRRepository) GetByID(prID string) (*entity.PullRequest, error) {
    var pr entity.PullRequest
    var mergedAt sql.NullTime
    
    err := r.db.QueryRow(`
        SELECT pull_request_id, pull_request_name, author_id, status, created_at, merged_at
        FROM pull_requests 
        WHERE pull_request_id = $1
    `, prID).Scan(
        &pr.PullRequestID,
        &pr.PullRequestName,
        &pr.AuthorID,
        &pr.Status,
        &pr.CreatedAt,
        &mergedAt,
    )
    
    if err == sql.ErrNoRows {
        return nil, fmt.Errorf("PR not found: %s", prID)
    }
    if err != nil {
        return nil, fmt.Errorf("failed to get PR: %w", err)
    }

    if mergedAt.Valid {
        pr.MergedAt = &mergedAt.Time
    }

    rows, err := r.db.Query(`
        SELECT reviewer_id 
        FROM pr_reviewers 
        WHERE pull_request_id = $1
        ORDER BY reviewer_id
    `, prID)
    if err != nil {
        return nil, fmt.Errorf("failed to get reviewers: %w", err)
    }
    defer rows.Close()

    var reviewers []string
    for rows.Next() {
        var reviewerID string
        if err := rows.Scan(&reviewerID); err != nil {
            return nil, fmt.Errorf("failed to scan reviewer: %w", err)
        }
        reviewers = append(reviewers, reviewerID)
    }

    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("error iterating reviewers: %w", err)
    }

    pr.AssignedReviewers = reviewers
    return &pr, nil
}

func (r *PRRepository) Update(pr *entity.PullRequest) error {
    tx, err := r.db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()

    var mergedAt sql.NullTime
    if pr.MergedAt != nil {
        mergedAt = sql.NullTime{Time: *pr.MergedAt, Valid: true}
    }

    _, err = tx.Exec(`
        UPDATE pull_requests 
        SET pull_request_name = $1, status = $2, merged_at = $3
        WHERE pull_request_id = $4
    `, pr.PullRequestName, pr.Status, mergedAt, pr.PullRequestID)
    if err != nil {
        return fmt.Errorf("failed to update PR: %w", err)
    }

    _, err = tx.Exec("DELETE FROM pr_reviewers WHERE pull_request_id = $1", pr.PullRequestID)
    if err != nil {
        return fmt.Errorf("failed to clear reviewers: %w", err)
    }

    for _, reviewerID := range pr.AssignedReviewers {
        _, err = tx.Exec(`
            INSERT INTO pr_reviewers (pull_request_id, reviewer_id)
            VALUES ($1, $2)
        `, pr.PullRequestID, reviewerID)
        if err != nil {
            return fmt.Errorf("failed to assign reviewer %s: %w", reviewerID, err)
        }
    }

    return tx.Commit()
}

func (r *PRRepository) GetByReviewer(userID string) ([]*entity.PullRequest, error) {
    rows, err := r.db.Query(`
        SELECT pr.pull_request_id, pr.pull_request_name, pr.author_id, pr.status, pr.created_at, pr.merged_at
        FROM pull_requests pr
        JOIN pr_reviewers prr ON pr.pull_request_id = prr.pull_request_id
        WHERE prr.reviewer_id = $1
        ORDER BY pr.created_at DESC
    `, userID)
    if err != nil {
        return nil, fmt.Errorf("failed to get PRs by reviewer: %w", err)
    }
    defer rows.Close()

    var prs []*entity.PullRequest
    for rows.Next() {
        var pr entity.PullRequest
        var mergedAt sql.NullTime
        
        err := rows.Scan(
            &pr.PullRequestID,
            &pr.PullRequestName,
            &pr.AuthorID,
            &pr.Status,
            &pr.CreatedAt,
            &mergedAt,
        )
        if err != nil {
            return nil, fmt.Errorf("failed to scan PR: %w", err)
        }

        if mergedAt.Valid {
            pr.MergedAt = &mergedAt.Time
        }

        reviewerRows, err := r.db.Query(`
            SELECT reviewer_id FROM pr_reviewers WHERE pull_request_id = $1
        `, pr.PullRequestID)
        if err != nil {
            return nil, fmt.Errorf("failed to get reviewers for PR %s: %w", pr.PullRequestID, err)
        }

        var reviewers []string
        for reviewerRows.Next() {
            var reviewerID string
            if err := reviewerRows.Scan(&reviewerID); err != nil {
                reviewerRows.Close()
                return nil, fmt.Errorf("failed to scan reviewer: %w", err)
            }
            reviewers = append(reviewers, reviewerID)
        }
        reviewerRows.Close()

        pr.AssignedReviewers = reviewers
        prs = append(prs, &pr)
    }

    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("error iterating PRs: %w", err)
    }

    return prs, nil
}

func (r *PRRepository) Exists(prID string) (bool, error) {
    query := `SELECT EXISTS(SELECT 1 FROM pull_requests WHERE pull_request_id = $1)`
    
    var exists bool
    err := r.db.QueryRow(query, prID).Scan(&exists)
    return exists, err
}