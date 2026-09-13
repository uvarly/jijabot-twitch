-- +goose Up
CREATE TABLE arena_duel_stake_history (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    duel_id     INTEGER NOT NULL REFERENCES arena_duels (id) ON DELETE CASCADE,
    user_id     INTEGER NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    amount      INTEGER NOT NULL CHECK (amount > 0),
    reason      TEXT NOT NULL CHECK (reason IN ('stake', 'refund', 'payout')),
    occured_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_arena_duel_stake_history_duel_id ON arena_duel_stake_history (duel_id);
CREATE INDEX idx_arena_duel_stake_history_user_id ON arena_duel_stake_history (user_id);

-- +goose Down
DROP TABLE arena_duel_stake_history;
