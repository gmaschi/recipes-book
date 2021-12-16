package main

import (
	"database/sql"
	"github.com/gmaschi/go-recipes-book/internal/factories"
	db "github.com/gmaschi/go-recipes-book/internal/services/datastore/postgresql/recipes/sqlc"
	"log"
)

const (
	dbDriver      = "postgres"
	dbSource      = "postgresql://root:root@localhost:5432/recipes?sslmode=disable"
	serverAddress = "0.0.0.0:8080"
)

func main() {
	conn, err := sql.Open(dbDriver, dbSource)
	if err != nil {
		log.Fatalln("could not connect to database:", err)
	}
	store := db.NewStore(conn)
	server := factories.New(store)

	err = server.Start(serverAddress)
	if err != nil {
		log.Fatalln("cannot start server", err)
	}
}
