package service

import (
	"context"
	"database/sql"
	"time"

	"github.com/gjtiquia/vimance/internal/db"
)

func (s *Service) CreateSavedQuery(ctx context.Context, name, dateFrom, dateTo string, currencyID *int64, fuzzy string, createdBy int64) (db.SavedQuery, error) {
	now := time.Now().Unix()
	return s.queries.CreateSavedQuery(ctx, db.CreateSavedQueryParams{
		Name:       name,
		DateFrom:   dateFrom,
		DateTo:     dateTo,
		CurrencyID: int64ToNullInt64(currencyID),
		FuzzyText:  fuzzy,
		CreatedAt:  now,
		CreatedBy:  createdBy,
		UpdatedAt:  now,
		UpdatedBy:  createdBy,
	})
}

func (s *Service) CreateSavedQueryWithTags(ctx context.Context, name, dateFrom, dateTo string, currencyID *int64, fuzzy string, createdBy int64, tagIDs []int64) (db.SavedQuery, error) {
	var result db.SavedQuery

	err := s.WithTransaction(func(q *db.Queries) error {
		now := time.Now().Unix()

		sq, err := q.CreateSavedQuery(ctx, db.CreateSavedQueryParams{
			Name:       name,
			DateFrom:   dateFrom,
			DateTo:     dateTo,
			CurrencyID: int64ToNullInt64(currencyID),
			FuzzyText:  fuzzy,
			CreatedAt:  now,
			CreatedBy:  createdBy,
			UpdatedAt:  now,
			UpdatedBy:  createdBy,
		})
		if err != nil {
			return err
		}
		result = sq

		for _, tagID := range tagIDs {
			if err := q.AddSavedQueryTag(ctx, db.AddSavedQueryTagParams{
				QueryID: sq.ID,
				TagID:   tagID,
			}); err != nil {
				return err
			}
		}

		return nil
	})

	return result, err
}

func (s *Service) GetSavedQuery(ctx context.Context, id int64) (db.SavedQuery, error) {
	return s.queries.GetSavedQuery(ctx, id)
}

func (s *Service) ListSavedQueries(ctx context.Context) ([]db.SavedQuery, error) {
	return s.queries.ListSavedQueries(ctx)
}

func (s *Service) GetSavedQueryTags(ctx context.Context, queryID int64) ([]db.Tag, error) {
	return s.queries.ListSavedQueryTags(ctx, queryID)
}

func (s *Service) UpdateSavedQuery(ctx context.Context, id int64, name, dateFrom, dateTo string, currencyID *int64, fuzzy string, updatedBy int64) (db.SavedQuery, error) {
	now := time.Now().Unix()
	return s.queries.UpdateSavedQuery(ctx, db.UpdateSavedQueryParams{
		ID:         id,
		Name:       name,
		DateFrom:   dateFrom,
		DateTo:     dateTo,
		CurrencyID: int64ToNullInt64(currencyID),
		FuzzyText:  fuzzy,
		UpdatedAt:  now,
		UpdatedBy:  updatedBy,
	})
}

func (s *Service) DeleteSavedQuery(ctx context.Context, id int64) error {
	return s.queries.DeleteSavedQuery(ctx, id)
}

func int64ToNullInt64(p *int64) sql.NullInt64 {
	if p == nil {
		return sql.NullInt64{Valid: false}
	}
	return sql.NullInt64{Int64: *p, Valid: true}
}

func nullInt64ToInt64(n sql.NullInt64) *int64 {
	if !n.Valid {
		return nil
	}
	return &n.Int64
}
