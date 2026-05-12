package service

import (
	"context"
	"time"

	"github.com/gjtiquia/vimance/internal/db"
)

type TargetWithActual struct {
	Target       db.Target
	SavedQuery   SavedQueryFilters
	ActualAmount *int64
	HasData      bool
}

type SavedQueryFilters struct {
	DateFrom   string
	DateTo     string
	CurrencyID *int64
	TagIDs     []int64
	FuzzyText  string
}

func (s *Service) CreateTarget(ctx context.Context, name string, savedQueryID int64, targetCents int64, createdBy int64) (db.Target, error) {
	now := time.Now().Unix()
	return s.queries.CreateTarget(ctx, db.CreateTargetParams{
		Name:         name,
		SavedQueryID: savedQueryID,
		TargetCents:  targetCents,
		CreatedAt:    now,
		CreatedBy:    createdBy,
		UpdatedAt:    now,
		UpdatedBy:    createdBy,
	})
}

func (s *Service) GetTarget(ctx context.Context, id int64) (db.Target, error) {
	return s.queries.GetTarget(ctx, id)
}

func (s *Service) ListTargets(ctx context.Context) ([]db.Target, error) {
	return s.queries.ListTargets(ctx)
}

func (s *Service) UpdateTarget(ctx context.Context, id int64, name string, savedQueryID int64, targetCents int64, updatedBy int64) (db.Target, error) {
	now := time.Now().Unix()
	return s.queries.UpdateTarget(ctx, db.UpdateTargetParams{
		ID:           id,
		Name:         name,
		SavedQueryID: savedQueryID,
		TargetCents:  targetCents,
		UpdatedAt:    now,
		UpdatedBy:    updatedBy,
	})
}

func (s *Service) DeleteTarget(ctx context.Context, id int64) error {
	return s.queries.DeleteTarget(ctx, id)
}

func (s *Service) GetTargetWithActual(ctx context.Context, id int64) (*TargetWithActual, error) {
	target, err := s.GetTarget(ctx, id)
	if err != nil {
		return nil, err
	}

	sq, err := s.GetSavedQuery(ctx, target.SavedQueryID)
	if err != nil {
		return nil, err
	}

	sqItem, err := s.buildSavedQueryFilters(ctx, sq)
	if err != nil {
		return nil, err
	}

	agg, err := s.Aggregate(ctx, sqItem.DateFrom, sqItem.DateTo, sqItem.CurrencyID, sqItem.TagIDs, sqItem.FuzzyText)
	if err != nil {
		return nil, err
	}

	result := &TargetWithActual{
		Target:     target,
		SavedQuery: *sqItem,
		HasData:    agg.HasData,
	}

	if agg.HasData {
		actual := agg.TotalAmount
		result.ActualAmount = &actual
	}

	return result, nil
}

func (s *Service) ListTargetsWithActuals(ctx context.Context) ([]TargetWithActual, error) {
	targets, err := s.ListTargets(ctx)
	if err != nil {
		return nil, err
	}

	results := make([]TargetWithActual, 0, len(targets))

	for _, target := range targets {
		sq, err := s.GetSavedQuery(ctx, target.SavedQueryID)
		if err != nil {
			continue
		}

		sqItem, err := s.buildSavedQueryFilters(ctx, sq)
		if err != nil {
			continue
		}

		agg, err := s.Aggregate(ctx, sqItem.DateFrom, sqItem.DateTo, sqItem.CurrencyID, sqItem.TagIDs, sqItem.FuzzyText)
		if err != nil {
			continue
		}

		twa := TargetWithActual{
			Target:     target,
			SavedQuery: *sqItem,
			HasData:    agg.HasData,
		}

		if agg.HasData {
			actual := agg.TotalAmount
			twa.ActualAmount = &actual
		}

		results = append(results, twa)
	}

	return results, nil
}

func (s *Service) buildSavedQueryFilters(ctx context.Context, sq db.SavedQuery) (*SavedQueryFilters, error) {
	item := &SavedQueryFilters{
		DateFrom:  sq.DateFrom,
		DateTo:    sq.DateTo,
		FuzzyText: sq.FuzzyText,
	}

	if sq.CurrencyID.Valid {
		v := sq.CurrencyID.Int64
		item.CurrencyID = &v
	}

	tags, err := s.GetSavedQueryTags(ctx, sq.ID)
	if err != nil {
		return nil, err
	}

	item.TagIDs = make([]int64, len(tags))
	for i, t := range tags {
		item.TagIDs[i] = t.ID
	}

	return item, nil
}