-- name: GetUser :one
SELECT * FROM users
WHERE users.username = sqlc.arg(username)
LIMIT 1;

-- name: CreateUser :one
INSERT INTO users (username, passwd)
VALUES (?, ?)
ON CONFLICT(username) DO NOTHING
RETURNING id;

-- name: CreateSession :one
INSERT INTO login_sessions (session_id, user_id)
VALUES(?, ?)
RETURNING id;

-- name: GetLoginSession :one
SELECT * FROM login_sessions
WHERE session_id = sqlc.arg(session_id)
LIMIT 1;

-- name: DeleteLoginSession :exec
DELETE FROM login_sessions
WHERE session_id = sqlc.arg(session_id);

-- name: GetMe :one
SELECT username, user_id FROM users
JOIN login_sessions AS ls ON users.id = ls.user_id
WHERE ls.session_id = sqlc.arg(session_id)
LIMIT 1;

-- name: ListBox :many
SELECT * FROM stockbox
WHERE host_id = sqlc.arg(user_id)
LIMIT 50;

-- name: CreateBox :one
INSERT INTO stockbox (box_name, host_id)
VALUES (?, ?)
RETURNING id;


-- name: DeleteBox :exec 
DELETE FROM stockbox 
WHERE id = sqlc.arg(box_id) AND host_id = sqlc.arg(host_id);

-- name: ListReadLater :many
SELECT * FROM read_later
WHERE box_id = sqlc.arg(box_id)
ORDER BY created_at DESC
LIMIT 50;

-- name: CreateReadLater :one
INSERT INTO read_later (box_id, author_id, title, urlstr, content, isShared)
VALUES (?, ?, ?, ?, ?, ?)
RETURNING id;

-- name: DeleteReadLater :exec
DELETE FROM read_later
WHERE id = sqlc.arg(id) AND author_id = sqlc.arg(user_id);

-- name: UpdateReadLater :one
UPDATE read_later
SET 
  box_id = sqlc.arg(box_id), 
  author_id = sqlc.arg(author_id), 
  title = sqlc.arg(title), 
  urlstr = sqlc.arg(urlstr), 
  content = sqlc.arg(content), 
  isshared = sqlc.arg(isshared)
WHERE id = sqlc.arg(id)
RETURNING id;
