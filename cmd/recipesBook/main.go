package main

import (
	"database/sql"
	"github.com/gmaschi/go-recipes-book/internal/factories/book-recipe-factory"
	db "github.com/gmaschi/go-recipes-book/internal/services/datastore/postgresql/recipes/sqlc"
	"github.com/gmaschi/go-recipes-book/pkg/config/env"
	_ "github.com/lib/pq"
	"log"
)

func main() {
	config, err := env.NewConfig()
	if err != nil {
		log.Fatalln("cannot load env variables")
	}

	conn, err := sql.Open(config.DbDriver, config.DbSource)
	if err != nil {
		log.Fatalln("could not connect to database:", err)
	}

	store := db.NewStore(conn)
	server, err := bookRecipeFactory.New(config, store)
	if err != nil {
		log.Fatalln("could not start server:", err)
	}

	err = server.Start(config.ServerAddress)
	if err != nil {
		log.Fatalln("cannot start server", err)
	}
}
