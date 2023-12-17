postgres_setup:
	docker run --name ratelimitingdb -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:latest
postgres_start:
	docker start ratelimitingdb

createdb:
	docker exec -it ratelimitingdb createdb --username=root --owner=root ratelimitingdb
dropdb:
	docker exec -it postgres12 dropdb --username=root ratelimitingdb

createdb2:
	docker exec -it rate-limit-postgres_db-1 createdb --username=root --owner=root ratelimitingdb

network:
	docker run --network host -d rate-limit-postgres_db-1

build:
	docker-compose up --build -d

migrateup:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/ratelimitingdb?sslmode=disable" -verbose up 

migratedown:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/ratelimitingdb?sslmode=disable" -verbose down
redis_clear_port:
	redis-cli shutdown
redis_setup:
	docker run --name rate_limiting -p 6379:6379 redis
redis_start:
	docker start rate_limiting


sqlc:
	sqlc generate

test:
	go test -v -cover ./...

compose:
	docker-compose up -d

.PHONY: postgres createdb dropdb migrateup migratedown sqlc test
