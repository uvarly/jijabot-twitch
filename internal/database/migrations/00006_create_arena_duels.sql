-- +goose Up
CREATE TABLE arena_duels (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    challenger_id INTEGER NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    opponent_id   INTEGER NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    stake         INTEGER NOT NULL CHECK (stake > 0),
    created_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at    TIMESTAMP NOT NULL,

    CHECK (challenger_id <> opponent_id)
);

CREATE INDEX idx_arena_duels_challenger_id ON arena_duels (challenger_id);
CREATE INDEX idx_arena_duels_opponent_id   ON arena_duels (opponent_id);
CREATE INDEX idx_arena_duels_expires_at    ON arena_duels (expires_at);

-- +goose Down
DROP TABLE arena_duels;
