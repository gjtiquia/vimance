package service

import (
	"context"
	"sort"
	"strings"

	"github.com/gjtiquia/vimance/internal/db"
)

type QueryResult struct {
	ID           int64
	Date         string
	AmountCents  int64
	AmountStr    string
	CurrencyCode string
	Notes        string
	TagNames     []string
}

func (s *Service) QueryRecords(ctx context.Context, dateFrom, dateTo string, currencyID *int64, tagIDs []int64, fuzzy string) ([]QueryResult, error) {
	records, err := s.queries.ListActiveRecordsByDateRange(ctx, db.ListActiveRecordsByDateRangeParams{
		FromDate: dateFrom,
		ToDate:   dateTo,
	})
	if err != nil {
		return nil, err
	}

	if len(records) == 0 {
		return []QueryResult{}, nil
	}

	// filter by currency
	if currencyID != nil {
		var filtered []db.ActiveRecord
		for _, r := range records {
			if r.CurrencyID == *currencyID {
				filtered = append(filtered, r)
			}
		}
		records = filtered
		if len(records) == 0 {
			return []QueryResult{}, nil
		}
	}

	// collect record IDs for tag batch fetch
	recordIDs := make([]int64, len(records))
	for i, r := range records {
		recordIDs[i] = r.ID
	}

	// batch fetch tag names for display
	tagNamesByRecord := make(map[int64][]string)
	tagRows, err := s.queries.GetRecordTagsByIDs(ctx, recordIDs)
	if err != nil {
		return nil, err
	}
	for _, t := range tagRows {
		tagNamesByRecord[t.RecordID] = append(tagNamesByRecord[t.RecordID], t.Name)
	}

	// batch fetch tag IDs for ALL-tags matching
	if len(tagIDs) > 0 {
		tagIDRows, err := s.queries.GetRecordTagIDsByIDs(ctx, recordIDs)
		if err != nil {
			return nil, err
		}

		tagIDsByRecord := make(map[int64]map[int64]bool)
		for _, t := range tagIDRows {
			if tagIDsByRecord[t.RecordID] == nil {
				tagIDsByRecord[t.RecordID] = make(map[int64]bool)
			}
			tagIDsByRecord[t.RecordID][t.TagID] = true
		}

		var filtered []db.ActiveRecord
		for _, r := range records {
			recordTagIDs := tagIDsByRecord[r.ID]
			matchesAll := true
			for _, tid := range tagIDs {
				if !recordTagIDs[tid] {
					matchesAll = false
					break
				}
			}
			if matchesAll {
				filtered = append(filtered, r)
			}
		}
		records = filtered
		if len(records) == 0 {
			return []QueryResult{}, nil
		}
	}

	// filter by fuzzy (search notes + tag names)
	if fuzzy != "" {
		fuzzyLower := strings.ToLower(fuzzy)
		var filtered []db.ActiveRecord
		for _, r := range records {
			if strings.Contains(strings.ToLower(r.Notes), fuzzyLower) {
				filtered = append(filtered, r)
				continue
			}
			for _, name := range tagNamesByRecord[r.ID] {
				if strings.Contains(strings.ToLower(name), fuzzyLower) {
					filtered = append(filtered, r)
					break
				}
			}
		}
		records = filtered
		if len(records) == 0 {
			return []QueryResult{}, nil
		}
	}

	// load all currencies for mapping
	currMap, err := s.loadCurrencyMap(ctx)
	if err != nil {
		return nil, err
	}

	// sort by date DESC, created_at DESC
	sort.Slice(records, func(i, j int) bool {
		if records[i].Date != records[j].Date {
			return records[i].Date > records[j].Date
		}
		return records[i].CreatedAt > records[j].CreatedAt
	})

	// build results
	results := make([]QueryResult, len(records))
	for i, r := range records {
		code := currMap[r.CurrencyID]
		if code == "" {
			code = "???"
		}
		results[i] = QueryResult{
			ID:           r.ID,
			Date:         r.Date,
			AmountCents:  r.AmountCents,
			AmountStr:    FormatCents(r.AmountCents),
			CurrencyCode: code,
			Notes:        r.Notes,
			TagNames:     tagNamesByRecord[r.ID],
		}
	}

	return results, nil
}

func (s *Service) loadCurrencyMap(ctx context.Context) (map[int64]string, error) {
	currencies, err := s.queries.ListCurrencies(ctx)
	if err != nil {
		return nil, err
	}

	m := make(map[int64]string, len(currencies))
	for _, c := range currencies {
		m[c.ID] = c.Code
	}
	return m, nil
}
