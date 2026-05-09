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

type RecordFull struct {
	Record       db.Record
	CurrencyCode string
	Tags         []db.Tag
	Parents      []LinkCandidate
}

func (s *Service) GetRecordFull(ctx context.Context, id int64) (*RecordFull, error) {
	record, err := s.queries.GetRecord(ctx, id)
	if err != nil {
		return nil, err
	}

	tags, err := s.queries.GetRecordTags(ctx, id)
	if err != nil {
		return nil, err
	}

	currency, err := s.queries.GetCurrency(ctx, record.CurrencyID)
	if err != nil {
		return nil, err
	}

	parentRecords, err := s.queries.GetRecordParents(ctx, id)
	if err != nil {
		return nil, err
	}

	parents := make([]LinkCandidate, 0, len(parentRecords))
	if len(parentRecords) > 0 {
		parentIDs := make([]int64, len(parentRecords))
		for i, p := range parentRecords {
			parentIDs[i] = p.ID
		}

		tagRows, err := s.queries.GetRecordTagsByIDs(ctx, parentIDs)
		if err != nil {
			return nil, err
		}
		parentTagMap := make(map[int64][]string)
		for _, t := range tagRows {
			parentTagMap[t.RecordID] = append(parentTagMap[t.RecordID], t.Name)
		}

		for _, p := range parentRecords {
			parents = append(parents, LinkCandidate{
				ID:          p.ID,
				Date:        p.Date,
				AmountCents: p.AmountCents,
				CurrencyID:  p.CurrencyID,
				Notes:       p.Notes,
				TagNames:    parentTagMap[p.ID],
			})
		}
	}

	return &RecordFull{
		Record:       record,
		CurrencyCode: currency.Code,
		Tags:         tags,
		Parents:      parents,
	}, nil
}

func (s *Service) UpdateRecordWithTagsAndLinks(ctx context.Context, id int64, date string, amountCents int64, currencyID int64, notes string, updatedBy int64, tagIDs []int64, parentIDs []int64) (db.Record, error) {
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

		if err := q.RemoveAllRecordTags(ctx, id); err != nil {
			return record, err
		}

		for _, tagID := range tagIDs {
			if err := q.AddRecordTag(ctx, db.AddRecordTagParams{
				RecordID:  record.ID,
				TagID:     tagID,
				CreatedAt: now,
				CreatedBy: updatedBy,
				UpdatedAt: now,
				UpdatedBy: updatedBy,
			}); err != nil {
				return record, err
			}
		}

		if err := q.RemoveAllRecordLinks(ctx, id); err != nil {
			return record, err
		}

		for _, parentID := range parentIDs {
			if err := q.AddRecordLink(ctx, db.AddRecordLinkParams{
				ParentID:  parentID,
				ChildID:   record.ID,
				CreatedAt: now,
				CreatedBy: updatedBy,
			}); err != nil {
				return record, err
			}
		}

		return record, nil
	})
}

func (s *Service) RemoveRecordTag(ctx context.Context, recordID int64, tagID int64) error {
	return s.queries.RemoveRecordTag(ctx, db.RemoveRecordTagParams{
		RecordID: recordID,
		TagID:    tagID,
	})
}
