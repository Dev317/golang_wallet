-- name: CreateAccount :one
INSERT INTO accounts (
  user_id, address, chain_id
) VALUES (
  $1, $2, $3
)
RETURNING *;

-- name: GetAccountByUserId :many
SELECT * FROM accounts WHERE user_id = $1;

-- name: GetAccountByAddressAndByChainId :one
SELECT * FROM accounts WHERE address = $1 AND chain_id = $2 LIMIT 1;

-- name: GetAccountAddressById :one
SELECT address FROM accounts WHERE id = $1 LIMIT 1;
