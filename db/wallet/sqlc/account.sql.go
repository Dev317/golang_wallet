// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0
// source: account.sql

package db_wallet

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const createAccount = `-- name: CreateAccount :one
INSERT INTO accounts (
  user_id, address, chain_id, balance
) VALUES (
  $1, $2, $3, $4
)
RETURNING id, user_id, address, chain_id, balance, created_at, updated_at
`

type CreateAccountParams struct {
	UserID  int64          `json:"user_id"`
	Address string         `json:"address"`
	ChainID int32          `json:"chain_id"`
	Balance pgtype.Numeric `json:"balance"`
}

func (q *Queries) CreateAccount(ctx context.Context, arg CreateAccountParams) (Account, error) {
	row := q.db.QueryRow(ctx, createAccount,
		arg.UserID,
		arg.Address,
		arg.ChainID,
		arg.Balance,
	)
	var i Account
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.Address,
		&i.ChainID,
		&i.Balance,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getAccountByAddressAndByChainId = `-- name: GetAccountByAddressAndByChainId :one
SELECT id, user_id, address, chain_id, balance, created_at, updated_at FROM accounts WHERE address = $1 AND chain_id = $2 LIMIT 1
`

type GetAccountByAddressAndByChainIdParams struct {
	Address string `json:"address"`
	ChainID int32  `json:"chain_id"`
}

func (q *Queries) GetAccountByAddressAndByChainId(ctx context.Context, arg GetAccountByAddressAndByChainIdParams) (Account, error) {
	row := q.db.QueryRow(ctx, getAccountByAddressAndByChainId, arg.Address, arg.ChainID)
	var i Account
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.Address,
		&i.ChainID,
		&i.Balance,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getAccountByUserId = `-- name: GetAccountByUserId :many
SELECT id, user_id, address, chain_id, balance, created_at, updated_at FROM accounts WHERE user_id = $1
`

func (q *Queries) GetAccountByUserId(ctx context.Context, userID int64) ([]Account, error) {
	rows, err := q.db.Query(ctx, getAccountByUserId, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Account
	for rows.Next() {
		var i Account
		if err := rows.Scan(
			&i.ID,
			&i.UserID,
			&i.Address,
			&i.ChainID,
			&i.Balance,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
