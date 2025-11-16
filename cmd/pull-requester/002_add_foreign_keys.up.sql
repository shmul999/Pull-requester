ALTER TABLE users ADD CONSTRAINT fk_users_team 
    FOREIGN KEY (team_name) REFERENCES teams(team_name) ON DELETE CASCADE;

ALTER TABLE pull_requests ADD CONSTRAINT fk_pr_author 
    FOREIGN KEY (author_id) REFERENCES users(user_id);

ALTER TABLE pr_reviewers ADD CONSTRAINT fk_pr_reviewers_pr 
    FOREIGN KEY (pull_request_id) REFERENCES pull_requests(pull_request_id) ON DELETE CASCADE;

ALTER TABLE pr_reviewers ADD CONSTRAINT fk_pr_reviewers_user 
    FOREIGN KEY (reviewer_id) REFERENCES users(user_id);