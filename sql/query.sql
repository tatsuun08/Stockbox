-- name: GetUser :one
SELECT * FROM users
WHERE users.username = :username AND users.passwd = :passwd 
LIMIT 1;

-- name: CreateUser :one
INSERT INTO users (username, passwd)
VALUES (?, ?)
ON CONFLICT(username) DO NOTHING
RETURNING (SELECT changes() = 0);
