-- name: GetUser :one
SELECT * FROM users
WHERE users.username = :username
LIMIT 1;

-- name: CreateUser :one
INSERT INTO users (username, passwd)
VALUES (?, ?)
ON CONFLICT(username) DO NOTHING
RETURNING (SELECT changes() = 0);
