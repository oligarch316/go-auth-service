-- Users table
CREATE TABLE users (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    display_name TEXT,
    password_hash TEXT NOT NULL,
    admin INTEGER DEFAULT 0
);

-- Invites table
CREATE TABLE invites (
    id INTEGER PRIMARY KEY,
    owner_id INTEGER,
    FOREIGN KEY(owner_id) REFERENCES users(id)
);

-- Initial user
INSERT INTO users (name, password_hash, admin) VALUES ('testuser', '$2a$10$uISdA44MZq7ePA0a/mea5uWb292tY.LRm87u.TmwOU9/51E02pTyG', 1);
