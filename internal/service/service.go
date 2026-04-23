package service

import (
	"database/sql"

	"github.com/gjtiquia/vimance/internal/db"
)

type Service struct {
	database *sql.DB
	queries  *db.Queries
}

func New(database *sql.DB) *Service {
	return &Service{
		database: database,
		queries:  db.New(database),
	}
}

func (s *Service) WithTransaction(fn func(*db.Queries) error) error {
	tx, err := s.database.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	if err := fn(db.New(tx)); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *Service) WithTransactionResult(fn func(*db.Queries) (db.Record, error)) (db.Record, error) {
	var result db.Record

	err := s.WithTransaction(func(q *db.Queries) error {
		var err error
		result, err = fn(q)
		return err
	})

	return result, err
}
