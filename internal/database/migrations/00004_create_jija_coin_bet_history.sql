-- +goose Up
CREATE TABLE jija_coin_bet_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    amount INTEGER NOT NULL CHECK (amount > 0),
    probability REAL NOT NULL CHECK (probability >= 0.0 AND probability <= 1.0),
    won INTEGER NOT NULL CHECK (won IN (0, 1)),
    payout INTEGER NOT NULL,
    bet_period TEXT NOT NULL,
    placed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_jija_coin_bet_history_user_id ON jija_coin_bet_history (user_id, bet_period);

-- +goose Down
DROP TABLE jija_coin_bet_history;
