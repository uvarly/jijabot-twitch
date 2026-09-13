-- +goose Up
CREATE TABLE arena_duel_status_history (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    duel_id         INTEGER NOT NULL REFERENCES arena_duels (id) ON DELETE CASCADE,
    status          TEXT NOT NULL CHECK (status IN ('pending', 'completed', 'declined', 'cancelled', 'expired')),
    challenger_roll INTEGER CHECK (challenger_roll BETWEEN 0 AND 20),
    opponent_roll   INTEGER CHECK (opponent_roll BETWEEN 0 AND 20),
    result          TEXT CHECK (result IN ('challenger_won', 'opponent_won', 'draw')),
    occured_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CHECK (
        (status = 'completed' AND challenger_roll IS NOT NULL AND opponent_roll IS NOT NULL AND result IS NOT NULL) OR
        (status <> 'completed' AND challenger_roll IS NULL AND opponent_roll IS NULL AND result IS NULL)
    )
);

CREATE INDEX idx_arena_duel_status_history_duel_id ON arena_duel_status_history (duel_id);
CREATE UNIQUE INDEX idx_arena_duel_status_history_at_most_one_non_pending_status_per_duel
    ON arena_duel_status_history (duel_id) WHERE status <> 'pending';

-- +goose Down
DROP TABLE arena_duel_status_history;
