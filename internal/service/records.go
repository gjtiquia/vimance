package service

import (
	"context"
	"database/sql"
	"time"

	"github.com/gjtiquia/vimance/internal/db"
)

func (s *Service) CreateRecord(ctx context.Context, date string, amountCents int64, currencyID int64, notes string, createdBy int64) (db.Record, error) {
	now := time.Now().Unix()
	return s.queries.CreateRecord(ctx, db.CreateRecordParams{
		Date:         date,
		AmountCents:  amountCents,
		CurrencyID:   currencyID,
		Notes:        notes,
		CreatedAt:    now,
		CreatedBy:    createdBy,
		UpdatedAt:    now,
		UpdatedBy:    createdBy,
	})
}

func (s *Service) CreateRecordWithTags(ctx context.Context, date string, amountCents int64, currencyID int64, notes string, createdBy int64, tagIDs []int64) (db.Record, error) {
	return s.CreateRecordWithTagsAndLinks(ctx, date, amountCents, currencyID, notes, createdBy, tagIDs, nil)
}

func (s *Service) CreateRecordWithTagsAndLinks(ctx context.Context, date string, amountCents int64, currencyID int64, notes string, createdBy int64, tagIDs []int64, parentIDs []int64) (db.Record, error) {
	return s.WithTransactionResult(func(q *db.Queries) (db.Record, error) {
		now := time.Now().Unix()

		record, err := q.CreateRecord(ctx, db.CreateRecordParams{
			Date:         date,
			AmountCents:  amountCents,
			CurrencyID:   currencyID,
			Notes:        notes,
			CreatedAt:    now,
			CreatedBy:    createdBy,
			UpdatedAt:    now,
			UpdatedBy:    createdBy,
		})
		if err != nil {
			return record, err
		}

		for _, tagID := range tagIDs {
			err := q.AddRecordTag(ctx, db.AddRecordTagParams{
				RecordID:  record.ID,
				TagID:     tagID,
				CreatedAt: now,
				CreatedBy: createdBy,
				UpdatedAt: now,
				UpdatedBy: createdBy,
			})
			if err != nil {
				return record, err
			}
		}

		for _, parentID := range parentIDs {
			err := q.AddRecordLink(ctx, db.AddRecordLinkParams{
				ParentID:  parentID,
				ChildID:   record.ID,
				CreatedAt: now,
				CreatedBy: createdBy,
			})
			if err != nil {
				return record, err
			}
		}

		return record, nil
	})
}

func (s *Service) GetRecord(ctx context.Context, id int64) (db.Record, error) {
	return s.queries.GetRecord(ctx, id)
}

func (s *Service) ListRecords(ctx context.Context) ([]db.Record, error) {
	return s.queries.ListRecords(ctx)
}

func (s *Service) ListActiveRecords(ctx context.Context) ([]db.ActiveRecord, error) {
	return s.queries.ListActiveRecords(ctx)
}

func (s *Service) ListActiveRecordsByDateRange(ctx context.Context, startDate string, endDate string) ([]db.ActiveRecord, error) {
	return s.queries.ListActiveRecordsByDateRange(ctx, db.ListActiveRecordsByDateRangeParams{
		FromDate: startDate,
		ToDate:   endDate,
	})
}

func (s *Service) UpdateRecord(ctx context.Context, id int64, date string, amountCents int64, currencyID int64, notes string, updatedBy int64) (db.Record, error) {
	return s.queries.UpdateRecord(ctx, db.UpdateRecordParams{
		ID:           id,
		Date:         date,
		AmountCents:  amountCents,
		CurrencyID:   currencyID,
		Notes:        notes,
		UpdatedAt:    time.Now().Unix(),
		UpdatedBy:    updatedBy,
	})
}

func (s *Service) UpdateRecordWithTags(ctx context.Context, id int64, date string, amountCents int64, currencyID int64, notes string, updatedBy int64, tagIDs []int64) (db.Record, error) {
	return s.WithTransactionResult(func(q *db.Queries) (db.Record, error) {
		now := time.Now().Unix()

		record, err := q.UpdateRecord(ctx, db.UpdateRecordParams{
			ID:           id,
			Date:         date,
			AmountCents:  amountCents,
			CurrencyID:   currencyID,
			Notes:        notes,
			UpdatedAt:    now,
			UpdatedBy:    updatedBy,
		})
		if err != nil {
			return record, err
		}

		err = q.RemoveAllRecordTags(ctx, id)
		if err != nil {
			return record, err
		}

		for _, tagID := range tagIDs {
			err := q.AddRecordTag(ctx, db.AddRecordTagParams{
				RecordID:  record.ID,
				TagID:     tagID,
				CreatedAt: now,
				CreatedBy: updatedBy,
				UpdatedAt: now,
				UpdatedBy: updatedBy,
			})
			if err != nil {
				return record, err
			}
		}

		return record, nil
	})
}

func (s *Service) SoftDeleteRecord(ctx context.Context, id int64, deletedBy int64) error {
	return s.queries.SoftDeleteRecord(ctx, db.SoftDeleteRecordParams{
		ID:        id,
		DeletedAt: sql.NullInt64{Int64: time.Now().Unix(), Valid: true},
		DeletedBy: sql.NullInt64{Int64: deletedBy, Valid: true},
	})
}

func (s *Service) RestoreRecord(ctx context.Context, id int64) (db.Record, error) {
	return s.queries.RestoreRecord(ctx, id)
}

func (s *Service) HardDeleteRecord(ctx context.Context, id int64) error {
	return s.queries.HardDeleteRecord(ctx, id)
}

func (s *Service) GetRecordTags(ctx context.Context, recordID int64) ([]db.Tag, error) {
	return s.queries.GetRecordTags(ctx, recordID)
}

func (s *Service) GetTagRecords(ctx context.Context, tagID int64) ([]db.Record, error) {
	return s.queries.GetTagRecords(ctx, tagID)
}

func (s *Service) AddRecordTag(ctx context.Context, recordID int64, tagID int64, createdBy int64) error {
	now := time.Now().Unix()
	return s.queries.AddRecordTag(ctx, db.AddRecordTagParams{
		RecordID:  recordID,
		TagID:     tagID,
		CreatedAt: now,
		CreatedBy: createdBy,
		UpdatedAt: now,
		UpdatedBy: createdBy,
	})
}

func (s *Service) RemoveRecordTag(ctx context.Context, recordID int64, tagID int64) error {
	return s.queries.RemoveRecordTag(ctx, db.RemoveRecordTagParams{
		RecordID: recordID,
		TagID:    tagID,
	})
}
