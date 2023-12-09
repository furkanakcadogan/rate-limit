postgres:
	docker run --name client_rate_limiting -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:latest
createdb:
	docker exec -it client_rate_limiting createdb --username=root --owner=root rate_limiting

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
