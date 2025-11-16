ALTER TABLE users DROP CONSTRAINT IF EXISTS fk_users_team;
ALTER TABLE pull_requests DROP CONSTRAINT IF EXISTS fk_pr_author;
ALTER TABLE pr_reviewers DROP CONSTRAINT IF EXISTS fk_pr_reviewers_pr;
ALTER TABLE pr_reviewers DROP CONSTRAINT IF EXISTS fk_pr_reviewers_user;