CREATE TABLE "rate_limits" (
  "id" SERIAL PRIMARY KEY,
  "clientid" VARCHAR(255) UNIQUE NOT NULL,
  "predefined_limit" INT NOT NULL,
  "remaining_limit" INT NOT NULL
);
