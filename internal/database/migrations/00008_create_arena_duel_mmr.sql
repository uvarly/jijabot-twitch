-- +goose Up
CREATE TABLE arena_duel_mmr (
    user_id    INTEGER PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
    rating     INTEGER NOT NULL CHECK (rating >= 0),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- CREATE INDEX idx_arena_duel_mmr_user_id ON arena_duel_mmr (user_id);

-- +goose Down
DROP TABLE arena_duel_mmr;
