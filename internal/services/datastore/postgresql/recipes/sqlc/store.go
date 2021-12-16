package db

import "database/sql"

type Store interface {
	Querier
}

type PostgresqlStore struct {
	db *sql.DB
	*Queries
}

func NewStore(db *sql.DB) PostgresqlStore {
	return PostgresqlStore{
		db:      db,
		Queries: New(db),
	}
}
