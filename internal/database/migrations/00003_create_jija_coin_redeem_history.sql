-- +goose Up
CREATE TABLE jija_coin_redeem_history (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id       INTEGER NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    amount        INTEGER NOT NULL CHECK (amount > 0),
    redeem_period TEXT NOT NULL,
    redeemed_at   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    UNIQUE (user_id, redeem_period)
);

CREATE INDEX idx_jija_coin_redeem_history_user_id ON jija_coin_redeem_history (user_id);

-- +goose Down
DROP TABLE jija_coin_redeem_history;
