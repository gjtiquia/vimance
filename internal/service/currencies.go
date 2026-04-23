package service

import (
	"context"
	"time"

	"github.com/gjtiquia/vimance/internal/db"
)

func (s *Service) CreateCurrency(ctx context.Context, code string) (db.Currency, error) {
	now := time.Now().Unix()
	return s.queries.CreateCurrency(ctx, db.CreateCurrencyParams{
		Code:      code,
		CreatedAt: now,
		UpdatedAt: now,
	})
}

func (s *Service) GetCurrency(ctx context.Context, id int64) (db.Currency, error) {
	return s.queries.GetCurrency(ctx, id)
}

func (s *Service) GetCurrencyByCode(ctx context.Context, code string) (db.Currency, error) {
	return s.queries.GetCurrencyByCode(ctx, code)
}

func (s *Service) ListCurrencies(ctx context.Context) ([]db.Currency, error) {
	return s.queries.ListCurrencies(ctx)
}

func (s *Service) UpdateCurrency(ctx context.Context, id int64, code string) (db.Currency, error) {
	return s.queries.UpdateCurrency(ctx, db.UpdateCurrencyParams{
		ID:        id,
		Code:      code,
		UpdatedAt: time.Now().Unix(),
	})
}

func (s *Service) DeleteCurrency(ctx context.Context, id int64) error {
	return s.queries.DeleteCurrency(ctx, id)
}
