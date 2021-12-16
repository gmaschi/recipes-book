postgres:
	sudo docker run --name postgres-recipes -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=root -e POSTGRES_DB=recipes -d postgres:latest

createdb:
	sudo docker exec -it postgres-recipes createdb --username=root --owner=root recipes

dropdb:
	sudo docker exec -it postgres-recipes dropdb recipes

migrateup:
	migrate -path internal/services/datastore/postgresql/recipes/migration/ -database "postgresql://root:root@localhost:5432/recipes?sslmode=disable" -verbose up

migrateup1:
	migrate -path internal/services/datastore/postgresql/recipes/migration/ -database "postgresql://root:root@localhost:5432/recipes?sslmode=disable" -verbose up 1

migratedown:
	migrate -path internal/services/datastore/postgresql/recipes/migration/ -database "postgresql://root:root@localhost:5432/recipes?sslmode=disable" -verbose down

migratedown1:
	migrate -path internal/services/datastore/postgresql/recipes/migration/ -database "postgresql://root:root@localhost:5432/recipes?sslmode=disable" -verbose down 1

sqlc:
	sqlc generate

test:
	go test -v -cover ./...

server:
	go run main.go

mock:
	mockgen -package mockdb -destination db/mock/store.go github.com/gmaschi/go-recipes-book/db/sqlc Store

.PHONY: postgres createdb dropdb migrateup migrateup1 migratedown migratedown1 sqlc test server mock