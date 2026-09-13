-- +goose Up
CREATE TABLE jija_coin_grant_history (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    granter_id INTEGER NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    grantee_id INTEGER NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    amount     INTEGER NOT NULL CHECK (amount > 0),
    granted_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_jija_coin_grant_history_grantee_id ON jija_coin_grant_history (grantee_id);

-- +goose Down
DROP TABLE jija_coin_grant_history;
