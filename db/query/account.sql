-- name: CreateRateLimit :one
INSERT INTO rate_limiting_db (
  clientid,
  rate_limit,
  refill_interval
) VALUES (
  $1, $2, $3
)
RETURNING *;
-- name: GetRateLimit :one
SELECT * FROM rate_limiting_db
WHERE id = $1 LIMIT 1;
-- name: ListRateLimits :one
SELECT * FROM rate_limiting_db
ORDER BY id
LIMIT $1
OFFSET $2;
-- name: UpdateRateLimit :exec
UPDATE rate_limiting_db
SET rate_limit = $2, refill_interval = $3
WHERE id = $1;
-- name: DeleteRateLimit :exec
DELETE FROM rate_limiting_db
WHERE id = $1;
