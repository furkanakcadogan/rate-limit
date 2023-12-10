-- name: CreateRateLimit :one
INSERT INTO ratelimitingdb (
  clientid,
  rate_limit,
  refill_interval
) VALUES (
  $1, $2, $3
)
RETURNING *;
-- name: GetRateLimit :one
SELECT * FROM ratelimitingdb
WHERE id = $1 LIMIT 1;
-- name: ListRateLimits :one
SELECT * FROM ratelimitingdb
ORDER BY id
LIMIT $1
OFFSET $2;
-- name: UpdateRateLimit :exec
UPDATE ratelimitingdb
SET rate_limit = $2, refill_interval = $3
WHERE id = $1;
-- name: DeleteRateLimit :exec
DELETE FROM ratelimitingdb
WHERE id = $1;
