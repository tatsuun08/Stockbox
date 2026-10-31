-- name: GetUser :one
SELECT * FROM users
WHERE users.username = :username
LIMIT 1;

-- name: CreateUser :one
INSERT INTO users (username, passwd)
VALUES (?, ?)
ON CONFLICT(username) DO NOTHING
RETURNING (SELECT changes() = 0);

-- name: CreateSession :one
INSERT INTO login_sessions (session_id, user_id)
VALUES(?, ?)
RETURNING id;

-- name: GetLoginSession :one
SELECT * FROM login_sessions
WHERE session_id = :session_id
LIMIT 1;

-- name: DeleteLoginSession :exec
DELETE FROM login_sessions
WHERE session_id = :session_id
LIMIT 1;