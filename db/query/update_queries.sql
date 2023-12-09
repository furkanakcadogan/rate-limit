-- db/query/update_queries.sql
-- name: UpdateRemainingLimit :exec
UPDATE rate_limits SET remaining_limit = $2 WHERE clientid = $1;
