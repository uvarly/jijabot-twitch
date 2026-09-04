-- +goose Up
CREATE TABLE jija_coin_wallets (
    user_id    INTEGER PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
    balance    INTEGER NOT NULL DEFAULT 0 CHECK (balance >= 0),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- CREATE INDEX idx_jija_coin_wallets_user_id ON jija_coin_wallets (user_id);

-- +goose Down
DROP TABLE jija_coin_wallets;
