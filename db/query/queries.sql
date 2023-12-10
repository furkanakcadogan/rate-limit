-- name: CreateRateLimit :exec
CREATE TABLE "rate_limits" (
    "id" SERIAL PRIMARY KEY,
    "clientid" VARCHAR(255) UNIQUE NOT NULL,
    "rate_limit" INT NOT NULL,
    "refill_interval" INT NOT NULL
);

-- name: GetRateLimitByID :one
SELECT * FROM "rate_limits" WHERE "id" = $1;

-- name: CreateNewUser :exec
INSERT INTO "rate_limits" ("clientid", "rate_limit", "refill_interval") VALUES ($1, $2, $3);
