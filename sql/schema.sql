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

CREATE TABLE IF NOT EXISTS stockbox (
  id INTEGER PRIMARY KEY,
  box_name TEXT NOT NULL,
  host_id INTEGER NOT NULL,
  FOREIGN KEY (host_id) REFERENCES users(id)
)

CREATE TABLE IF NOT EXISTS read_later (
  id INTEGER PRIMARY KEY,
  box_id INTEGER NOT NULL,
  author_id INTEGER NOT NULL,
  title TEXT NOT NULL,
  urlstr TEXT,
  content TEXT,
  isShared BOOLEAN DEFAULT 0 NOT NULL,

  FOREIGN KEY (box_id) REFERENCES stockbox(id) ON DELETE CASCADE
)
