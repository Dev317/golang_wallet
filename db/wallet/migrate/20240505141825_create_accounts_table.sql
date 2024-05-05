-- +goose Up
CREATE TABLE accounts (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGSERIAL NOT NULL,
    address VARCHAR NOT NULL,
    chain_id INT NOT NULL,
    balance NUMERIC NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_user_id FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE SET NULL
);

CREATE INDEX accounts_user_id_index ON accounts (user_id);
CREATE INDEX accounts_address_index ON accounts (address);

-- +goose Down
DROP TABLE IF EXISTS accounts;
