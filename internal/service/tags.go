package service

import (
	"context"
	"database/sql"
	"time"

	"github.com/gjtiquia/vimance/internal/db"
)

func (s *Service) CreateTag(ctx context.Context, name string, description string, notes string, createdBy int64) (db.Tag, error) {
	now := time.Now().Unix()
	return s.queries.CreateTag(ctx, db.CreateTagParams{
		Name:        name,
		Description: description,
		Notes:       notes,
		CreatedAt:   now,
		CreatedBy:   createdBy,
		UpdatedAt:   now,
		UpdatedBy:   createdBy,
	})
}

func (s *Service) GetTag(ctx context.Context, id int64) (db.Tag, error) {
	return s.queries.GetTag(ctx, id)
}

func (s *Service) GetTagByName(ctx context.Context, name string) (db.Tag, error) {
	return s.queries.GetTagByName(ctx, name)
}

func (s *Service) ListTags(ctx context.Context) ([]db.Tag, error) {
	return s.queries.ListTags(ctx)
}

func (s *Service) ListActiveTags(ctx context.Context) ([]db.ActiveTag, error) {
	return s.queries.ListActiveTags(ctx)
}

func (s *Service) ListPinnedTags(ctx context.Context) ([]db.Tag, error) {
	return s.queries.ListPinnedTags(ctx)
}

func (s *Service) UpdateTag(ctx context.Context, id int64, name string, description string, notes string, updatedBy int64) (db.Tag, error) {
	return s.queries.UpdateTag(ctx, db.UpdateTagParams{
		ID:          id,
		Name:        name,
		Description: description,
		Notes:       notes,
		UpdatedAt:   time.Now().Unix(),
		UpdatedBy:   updatedBy,
	})
}

func (s *Service) SoftDeleteTag(ctx context.Context, id int64, deletedBy int64) error {
	return s.queries.SoftDeleteTag(ctx, db.SoftDeleteTagParams{
		ID:        id,
		DeletedAt: sql.NullInt64{Int64: time.Now().Unix(), Valid: true},
		DeletedBy: sql.NullInt64{Int64: deletedBy, Valid: true},
	})
}

func (s *Service) RestoreTag(ctx context.Context, id int64) (db.Tag, error) {
	return s.queries.RestoreTag(ctx, id)
}

func (s *Service) HardDeleteTag(ctx context.Context, id int64) error {
	return s.queries.HardDeleteTag(ctx, id)
}

func (s *Service) PinTag(ctx context.Context, tagID int64, createdBy int64) error {
	maxPos, err := s.queries.GetMaxPinnedPosition(ctx)
	if err != nil {
		return err
	}

	var position int64
	switch v := maxPos.(type) {
	case int64:
		position = v + 1
	case float64:
		position = int64(v) + 1
	default:
		position = 1
	}

	now := time.Now().Unix()
	return s.queries.PinTag(ctx, db.PinTagParams{
		TagID:     tagID,
		Position:  position,
		CreatedAt: now,
		CreatedBy: createdBy,
		UpdatedAt: now,
		UpdatedBy: createdBy,
	})
}

func (s *Service) UnpinTag(ctx context.Context, tagID int64) error {
	return s.queries.UnpinTag(ctx, tagID)
}
