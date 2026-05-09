package service

import (
	"context"
	"time"

	"github.com/gjtiquia/vimance/internal/db"
)

type LinkCandidate struct {
	ID          int64
	Date        string
	AmountCents int64
	CurrencyID  int64
	Notes       string
	TagNames    []string
}

func (s *Service) LinkRecords(ctx context.Context, parentID int64, childID int64, createdBy int64) error {
	now := time.Now().Unix()
	return s.queries.AddRecordLink(ctx, db.AddRecordLinkParams{
		ParentID:  parentID,
		ChildID:   childID,
		CreatedAt: now,
		CreatedBy: createdBy,
	})
}

func (s *Service) UnlinkRecords(ctx context.Context, parentID int64, childID int64) error {
	return s.queries.RemoveRecordLink(ctx, db.RemoveRecordLinkParams{
		ParentID: parentID,
		ChildID:  childID,
	})
}

func (s *Service) GetRecordParents(ctx context.Context, recordID int64) ([]db.Record, error) {
	return s.queries.GetRecordParents(ctx, recordID)
}

func (s *Service) GetRecordChildren(ctx context.Context, recordID int64) ([]db.Record, error) {
	return s.queries.GetRecordChildren(ctx, recordID)
}

func (s *Service) SearchLinkCandidates(ctx context.Context, dateFrom string, dateTo string, currencyID int64, excludeID int64) ([]LinkCandidate, error) {
	rows, err := s.queries.SearchParentCandidates(ctx, db.SearchParentCandidatesParams{
		Date:       dateFrom,
		Date_2:     dateTo,
		CurrencyID: currencyID,
		ID:         excludeID,
	})
	if err != nil {
		return nil, err
	}

	if len(rows) == 0 {
		return []LinkCandidate{}, nil
	}

	recordIDs := make([]int64, len(rows))
	for i, r := range rows {
		recordIDs[i] = r.ID
	}

	tagRows, err := s.queries.GetRecordTagsByIDs(ctx, recordIDs)
	if err != nil {
		return nil, err
	}

	tagMap := make(map[int64][]string)
	for _, t := range tagRows {
		tagMap[t.RecordID] = append(tagMap[t.RecordID], t.Name)
	}

	candidates := make([]LinkCandidate, len(rows))
	for i, r := range rows {
		candidates[i] = LinkCandidate{
			ID:          r.ID,
			Date:        r.Date,
			AmountCents: r.AmountCents,
			CurrencyID:  r.CurrencyID,
			Notes:       r.Notes,
			TagNames:    tagMap[r.ID],
		}
	}

	return candidates, nil
}
