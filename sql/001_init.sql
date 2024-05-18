DROP TABLE IF EXISTS users;

CREATE TABLE users (
    username TEXT PRIMARY KEY,
    password TEXT NOT NULL
);

INSERT INTO users VALUES ('alice', 'abc123');
INSERT INTO users VALUES ('bob', '123abc');
