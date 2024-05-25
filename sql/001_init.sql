PRAGMA foreign_keys = ON;

DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS tokens;

CREATE TABLE users (
    username TEXT PRIMARY KEY,
    salt TEXT NOT NULL,
    hash TEXT NOT NULL,
    count INTEGER NOT NULL
);

CREATE TABLE tokens (
    key TEXT PRIMARY KEY,
    owner TEXT NOT NULL,
    FOREIGN KEY(owner) REFERENCES users(username) ON DELETE CASCADE
);
