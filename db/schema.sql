CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    username TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS matches (
    id INTEGER PRIMARY KEY,
    home_team TEXT NOT NULL,
    away_team TEXT NOT NULL,
    home_score INTEGER,
    away_score INTEGER,
    stage TEXT,
    starts_at DATETIME NOT NULL,
    status TEXT DEFAULT 'scheduled'
);

CREATE TABLE IF NOT EXISTS guesses (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id TEXT NOT NULL REFERENCES users(id),
    match_id INTEGER NOT NULL REFERENCES matches(id),
    home_score INTEGER NOT NULL,
    away_score INTEGER NOT NULL,
    points INTEGER DEFAULT 0,
    calculated INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, match_id)
);

CREATE TABLE IF NOT EXISTS sync_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    type TEXT NOT NULL,
    synced_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
