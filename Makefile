postgres_setup:
	docker run --name rate_limiting_db -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:latest
postgres_start:
	docker start rate_limiting_db
createdb:
	docker exec -it rate_limiting_db createdb --username=root --owner=root rate_limiting

redis_clear_port:
	redis-cli shutdown
redis_setup:
	docker run --name rate_limiting_redis -p 6379:6379 redis
redis_start:
	docker start rate_limiting_redis
grpc:
	docker run rate-limiter-app

dropdb:
	docker exec -it postgres12 dropdb --username=root rate_limiting

migrateup:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/rate_limiting?sslmode=disable" -verbose up 

migratedown:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/rate_limiting?sslmode=disable" -verbose down

sqlc:
	sqlc generate

test:
	go test -v -cover ./...



.PHONY: postgres createdb dropdb migrateup migratedown sqlc test
