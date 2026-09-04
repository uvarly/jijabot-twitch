-- +goose Up
CREATE TABLE jija_coin_redeem_history (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id       INTEGER NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    amount        INTEGER NOT NULL CHECK (amount > 0),
    redeemed_at   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    redeemed_date TEXT GENERATED ALWAYS AS (date(redeemed_at)) STORED,

    UNIQUE (user_id, redeemed_date)
);

CREATE INDEX idx_jija_coin_redeem_history_user_id ON jija_coin_redeem_history (user_id);

-- +goose Down
DROP TABLE jija_coin_redeem_history;
