-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens(token, created_at, updated_at, user_id, expires_at)
VALUES ($1, $2, $3, $4, $5)
  RETURNING *;

-- name: RevokeToken :exec
UPDATE refresh_tokens
SET updated_at = $1, revoked_at = $2
WHERE token = $3;

-- name: GetTokenByUserID :one
SELECT refresh_tokens.token FROM refresh_tokens
WHERE refresh_tokens.user_id = (
  SELECT users.id FROM users
  INNER JOIN refresh_tokens on refresh_tokens.user_id = users.id
  WHERE users.id = $1
);

-- name: GetUserFromRefreshToken :one
SELECT * FROM users
WHERE users.id = (
  SELECT user_id FROM refresh_tokens
  INNER JOIN users on user_id = users.id
  WHERE refresh_tokens.token = $1
);
