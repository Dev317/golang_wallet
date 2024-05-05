-- name: CreateAccount :one
INSERT INTO accounts (
  user_id, address, chain_id, balance
) VALUES (
  $1, $2, $3, $4
)
RETURNING *;

-- name: GetAccountByUserId :many
SELECT * FROM accounts WHERE user_id = $1;

-- name: GetAccountByAddressAndByChainId :one
SELECT * FROM accounts WHERE address = $1 AND chain_id = $2 LIMIT 1;
