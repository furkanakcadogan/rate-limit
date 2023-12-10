CREATE TABLE "ratelimitingdb" (
  "id" SERIAL PRIMARY KEY,
  "clientid" VARCHAR(255) UNIQUE NOT NULL,
  "rate_limit" INT NOT NULL,
  "refill_interval" INT NOT NULL
);
