-- db/query/insert_queries.sql
-- name: InsertRateLimit :exec
INSERT INTO public.rate_limits (clientid, predefined_limit, remaining_limit) VALUES ($1, $2, $3);
