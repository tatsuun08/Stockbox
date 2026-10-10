-- name: GetUser :one
SELECT * FROM users
WHERE ID = ? LIMIT 1;