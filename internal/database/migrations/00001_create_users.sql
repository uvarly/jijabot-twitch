-- +goose Up
CREATE TABLE users (
    id             INTEGER PRIMARY KEY AUTOINCREMENT,
    twitch_user_id TEXT NOT NULL UNIQUE,
    username       TEXT NOT NULL,
    created_at     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_username ON users (username);

-- +goose Down
DROP TABLE users;
