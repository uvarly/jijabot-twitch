-- +goose Up
CREATE TABLE arena_duel_mmr_history (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    duel_id      INTEGER NOT NULL REFERENCES arena_duels (id) ON DELETE CASCADE,
    user_id      INTEGER NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    rating_delta INTEGER NOT NULL,
    occured_at   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_arena_duel_mmr_history_duel_id ON arena_duel_mmr_history (duel_id);
CREATE INDEX idx_arena_duel_mmr_history_user_id ON arena_duel_mmr_history (user_id);

-- +goose Down
DROP TABLE arena_duel_mmr_history;
