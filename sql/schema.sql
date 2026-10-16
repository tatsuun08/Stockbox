CREATE TABLE IF NOT EXISTS users (
  id INTEGER PRIMARY KEY,
  username text NOT NULL UNIQUE,
  passwd text NOT NULL
)