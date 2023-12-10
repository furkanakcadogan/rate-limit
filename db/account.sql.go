// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.24.0
// source: account.sql

package db

import (
	"context"
)

const createRateLimit = `-- name: CreateRateLimit :one
INSERT INTO ratelimitingdb (
  clientid,
  rate_limit,
  refill_interval
) VALUES (
  $1, $2, $3
)
RETURNING id, clientid, rate_limit, refill_interval
`

type CreateRateLimitParams struct {
	Clientid       string `json:"clientid"`
	RateLimit      int32  `json:"rate_limit"`
	RefillInterval int32  `json:"refill_interval"`
}

func (q *Queries) CreateRateLimit(ctx context.Context, arg CreateRateLimitParams) (Ratelimitingdb, error) {
	row := q.queryRow(ctx, q.createRateLimitStmt, createRateLimit, arg.Clientid, arg.RateLimit, arg.RefillInterval)
	var i Ratelimitingdb
	err := row.Scan(
		&i.ID,
		&i.Clientid,
		&i.RateLimit,
		&i.RefillInterval,
	)
	return i, err
}

const deleteRateLimit = `-- name: DeleteRateLimit :exec
DELETE FROM ratelimitingdb
WHERE clientid = $1
`

func (q *Queries) DeleteRateLimit(ctx context.Context, clientid string) error {
	_, err := q.exec(ctx, q.deleteRateLimitStmt, deleteRateLimit, clientid)
	return err
}

const getRateLimit = `-- name: GetRateLimit :one
SELECT id, clientid, rate_limit, refill_interval FROM ratelimitingdb
WHERE clientid = $1 LIMIT 1
`

func (q *Queries) GetRateLimit(ctx context.Context, clientid string) (Ratelimitingdb, error) {
	row := q.queryRow(ctx, q.getRateLimitStmt, getRateLimit, clientid)
	var i Ratelimitingdb
	err := row.Scan(
		&i.ID,
		&i.Clientid,
		&i.RateLimit,
		&i.RefillInterval,
	)
	return i, err
}

const listRateLimits = `-- name: ListRateLimits :one
SELECT id, clientid, rate_limit, refill_interval FROM ratelimitingdb
ORDER BY clientid
LIMIT $1
OFFSET $2
`

type ListRateLimitsParams struct {
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

func (q *Queries) ListRateLimits(ctx context.Context, arg ListRateLimitsParams) (Ratelimitingdb, error) {
	row := q.queryRow(ctx, q.listRateLimitsStmt, listRateLimits, arg.Limit, arg.Offset)
	var i Ratelimitingdb
	err := row.Scan(
		&i.ID,
		&i.Clientid,
		&i.RateLimit,
		&i.RefillInterval,
	)
	return i, err
}

const updateRateLimit = `-- name: UpdateRateLimit :exec
UPDATE ratelimitingdb
SET rate_limit = $2, refill_interval = $3
WHERE clientid = $1
`

type UpdateRateLimitParams struct {
	Clientid       string `json:"clientid"`
	RateLimit      int32  `json:"rate_limit"`
	RefillInterval int32  `json:"refill_interval"`
}

func (q *Queries) UpdateRateLimit(ctx context.Context, arg UpdateRateLimitParams) error {
	_, err := q.exec(ctx, q.updateRateLimitStmt, updateRateLimit, arg.Clientid, arg.RateLimit, arg.RefillInterval)
	return err
}