package pgstorage

import (
	"database/sql"

	_ "github.com/lib/pq"
)

type PGStorage struct {
	db *sql.DB
}

func New(connStr string) (*PGStorage, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	return &PGStorage{
		db: db,
	}, nil
}

func (s *PGStorage) Ping() error {
	return s.db.Ping()
}
