-- db/query/select_queries.sql
-- name: GetRateLimitByClientID :one
SELECT * FROM rate_limits WHERE clientid = $1;
