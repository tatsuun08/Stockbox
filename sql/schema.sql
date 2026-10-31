PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS users (
  id INTEGER PRIMARY KEY,
  username TEXT NOT NULL UNIQUE,
  passwd TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS login_sessions (
  id INTEGER PRIMARY KEY,
  session_id TEXT NOT NULL,
  user_id INTEGER NOT NULL,
  expires_at TEXT  DEFAULT (datetime('now', '+1 day')),
  created_at TEXT DEFAULT (datetime('now')),
  FOREIGN KEY (user_id) REFERENCES users(id)
);
