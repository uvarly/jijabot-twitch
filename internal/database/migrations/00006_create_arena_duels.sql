-- +goose Up
CREATE TABLE arena_duels (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    challenger_id   INTEGER NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    opponent_id     INTEGER NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    stake           INTEGER NOT NULL CHECK (stake > 0),
    status          TEXT NOT NULL CHECK (status IN ('pending', 'resolved', 'declined', 'cancelled', 'expired')),
    challenger_roll INTEGER CHECK (challenger_roll BETWEEN 0 AND 20),
    opponent_roll   INTEGER CHECK (opponent_roll BETWEEN 0 AND 20),
    result          TEXT CHECK (result IN ('challenger_won', 'opponent_won', 'draw')),
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at      TIMESTAMP NOT NULL,
    resolved_at     TIMESTAMP,

    CHECK (challenger_id <> opponent_id),
    CHECK (
        (status = 'resolved'  AND challenger_roll IS NOT NULL AND opponent_roll IS NOT NULL AND result IS NOT NULL) OR
        (status <> 'resolved' AND challenger_roll IS NULL     AND opponent_roll IS NULL     AND result IS NULL)
    )
);

CREATE INDEX idx_arena_duels_challenger_id     ON arena_duels (challenger_id);
CREATE INDEX idx_arena_duels_opponent_id       ON arena_duels (opponent_id);
CREATE INDEX idx_arena_duels_status_expires_at ON arena_duels (status, expires_at);

CREATE UNIQUE INDEX idx_arena_duels_one_pending_per_challenger ON arena_duels (challenger_id) WHERE status = 'pending';
CREATE UNIQUE INDEX idx_arena_duels_one_pending_per_opponent   ON arena_duels (opponent_id)   WHERE status = 'pending';

-- +goose Down
DROP TABLE arena_duels;
