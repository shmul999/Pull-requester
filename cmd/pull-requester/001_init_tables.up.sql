CREATE TABLE IF NOT EXISTS teams (
    team_name VARCHAR(255) PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS users (
    user_id VARCHAR(255) PRIMARY KEY,
    username VARCHAR(255) NOT NULL,
    team_name VARCHAR(255) NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS pull_requests (
    pull_request_id VARCHAR(255) PRIMARY KEY,
    pull_request_name VARCHAR(255) NOT NULL,
    author_id VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'OPEN',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    merged_at TIMESTAMP WITH TIME ZONE,
    CHECK (status IN ('OPEN', 'MERGED'))
);

CREATE TABLE IF NOT EXISTS pr_reviewers (
    pull_request_id VARCHAR(255),
    reviewer_id VARCHAR(255),
    assigned_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (pull_request_id, reviewer_id)
);

CREATE INDEX IF NOT EXISTS idx_users_team_active ON users(team_name, is_active);
CREATE INDEX IF NOT EXISTS idx_pr_reviewers ON pr_reviewers(reviewer_id);
CREATE INDEX IF NOT EXISTS idx_pr_status ON pull_requests(status);