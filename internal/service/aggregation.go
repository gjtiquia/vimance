package service

import (
	"context"
	"fmt"
	"sort"
	"time"
)

type AggregationResult struct {
	TotalAmount int64
	IncomeSum   int64
	ExpenseSum  int64
	RecordCount int
	HasData     bool
	ByTag       []TagSum
}

type TagSum struct {
	TagName string
	Amount  int64
	Count   int
}

type PeriodAggregationResult struct {
	TotalAmount int64
	IncomeSum   int64
	ExpenseSum  int64
	RecordCount int
	HasData     bool
	Periods     []PeriodSum
}

type PeriodSum struct {
	Period     string
	Amount     int64
	IncomeSum  int64
	ExpenseSum int64
	Count      int
}

type PeriodGrouping string

const (
	PeriodByDay   PeriodGrouping = "day"
	PeriodByWeek  PeriodGrouping = "week"
	PeriodByMonth PeriodGrouping = "month"
	PeriodByYear  PeriodGrouping = "year"
)

func (s *Service) Aggregate(ctx context.Context, dateFrom, dateTo string, currencyID *int64, tagIDs []int64, fuzzy string) (*AggregationResult, error) {
	results, err := s.QueryRecords(ctx, dateFrom, dateTo, currencyID, tagIDs, fuzzy)
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return &AggregationResult{
			HasData: false,
		}, nil
	}

	result := &AggregationResult{
		HasData:     true,
		RecordCount: len(results),
	}

	tagSums := make(map[string]*TagSum)
	seenRecordsByTag := make(map[int64]bool)

	for _, r := range results {
		result.TotalAmount += r.AmountCents
		if r.AmountCents > 0 {
			result.IncomeSum += r.AmountCents
		} else {
			result.ExpenseSum += r.AmountCents
		}

		if len(r.TagNames) == 0 {
			key := "(untagged)"
			if ts, ok := tagSums[key]; ok {
				ts.Amount += r.AmountCents
				ts.Count++
			} else {
				tagSums[key] = &TagSum{
					TagName: key,
					Amount:  r.AmountCents,
					Count:   1,
				}
			}
		} else {
			for _, tagName := range r.TagNames {
				if ts, ok := tagSums[tagName]; ok {
					ts.Amount += r.AmountCents
					ts.Count++
				} else {
					tagSums[tagName] = &TagSum{
						TagName: tagName,
						Amount:  r.AmountCents,
						Count:   1,
					}
				}
			}
		}
		_ = seenRecordsByTag
	}

	for _, ts := range tagSums {
		result.ByTag = append(result.ByTag, *ts)
	}
	sort.Slice(result.ByTag, func(i, j int) bool {
		return result.ByTag[i].TagName < result.ByTag[j].TagName
	})

	return result, nil
}

func (s *Service) AggregateByPeriod(ctx context.Context, dateFrom, dateTo string, currencyID *int64, tagIDs []int64, fuzzy string, grouping PeriodGrouping) (*PeriodAggregationResult, error) {
	results, err := s.QueryRecords(ctx, dateFrom, dateTo, currencyID, tagIDs, fuzzy)
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return &PeriodAggregationResult{
			HasData: false,
		}, nil
	}

	result := &PeriodAggregationResult{
		HasData:     true,
		RecordCount: len(results),
	}

	periodMap := make(map[string]*PeriodSum)

	for _, r := range results {
		result.TotalAmount += r.AmountCents
		if r.AmountCents > 0 {
			result.IncomeSum += r.AmountCents
		} else {
			result.ExpenseSum += r.AmountCents
		}

		periodKey := periodKey(r.Date, grouping)
		if ps, ok := periodMap[periodKey]; ok {
			ps.Amount += r.AmountCents
			if r.AmountCents > 0 {
				ps.IncomeSum += r.AmountCents
			} else {
				ps.ExpenseSum += r.AmountCents
			}
			ps.Count++
		} else {
			ps := &PeriodSum{
				Period: periodKey,
				Count:  1,
			}
			ps.Amount = r.AmountCents
			if r.AmountCents > 0 {
				ps.IncomeSum = r.AmountCents
			} else {
				ps.ExpenseSum = r.AmountCents
			}
			periodMap[periodKey] = ps
		}
	}

	for _, ps := range periodMap {
		result.Periods = append(result.Periods, *ps)
	}
	sort.Slice(result.Periods, func(i, j int) bool {
		return result.Periods[i].Period < result.Periods[j].Period
	})

	return result, nil
}

func periodKey(dateStr string, grouping PeriodGrouping) string {
	switch grouping {
	case PeriodByDay:
		return dateStr
	case PeriodByWeek:
		t, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			return dateStr
		}
		year, week := t.ISOWeek()
		return formatISOWeek(year, week)
	case PeriodByMonth:
		if len(dateStr) >= 7 {
			return dateStr[:7]
		}
		return dateStr
	case PeriodByYear:
		if len(dateStr) >= 4 {
			return dateStr[:4]
		}
		return dateStr
	default:
		if len(dateStr) >= 7 {
			return dateStr[:7]
		}
		return dateStr
	}
}

func formatISOWeek(year int, week int) string {
	return fmt.Sprintf("%04d-W%02d", year, week)
}