package service_test

import (
	"testing"
)

func TestQueryRecordsDateRange(t *testing.T) {
	s := setupTestService(t)

	user, _ := s.CreateUser(t.Context(), "testuser")
	currency, _ := s.CreateCurrency(t.Context(), "USD")

	s.CreateRecord(t.Context(), "2026-01-10", 1000, currency.ID, "", user.ID)
	s.CreateRecord(t.Context(), "2026-01-15", 2000, currency.ID, "", user.ID)
	s.CreateRecord(t.Context(), "2026-01-20", 3000, currency.ID, "", user.ID)
	s.CreateRecord(t.Context(), "2026-01-25", 4000, currency.ID, "", user.ID)

	results, err := s.QueryRecords(t.Context(), "2026-01-12", "2026-01-22", nil, nil, "")
	if err != nil {
		t.Fatalf("failed to query records: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 records in range, got %d", len(results))
	}
}

func TestQueryRecordsCurrencyFilter(t *testing.T) {
	s := setupTestService(t)

	user, _ := s.CreateUser(t.Context(), "testuser")
	usd, _ := s.CreateCurrency(t.Context(), "USD")
	php, _ := s.CreateCurrency(t.Context(), "PHP")

	s.CreateRecord(t.Context(), "2026-01-15", 1000, usd.ID, "", user.ID)
	s.CreateRecord(t.Context(), "2026-01-16", 2000, php.ID, "", user.ID)
	s.CreateRecord(t.Context(), "2026-01-17", 3000, usd.ID, "", user.ID)
	s.CreateRecord(t.Context(), "2026-01-18", 4000, php.ID, "", user.ID)

	phpID := php.ID
	results, err := s.QueryRecords(t.Context(), "2026-01-01", "2026-01-31", &phpID, nil, "")
	if err != nil {
		t.Fatalf("failed to query records: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 PHP records, got %d", len(results))
	}
	for _, r := range results {
		if r.CurrencyCode != "PHP" {
			t.Errorf("expected currency PHP, got %s", r.CurrencyCode)
		}
	}
}

func TestQueryRecordsAllTagsFilter(t *testing.T) {
	s := setupTestService(t)

	user, _ := s.CreateUser(t.Context(), "testuser")
	currency, _ := s.CreateCurrency(t.Context(), "USD")
	food, _ := s.CreateTag(t.Context(), "food", "", "", user.ID)
	drink, _ := s.CreateTag(t.Context(), "drink", "", "", user.ID)
	snack, _ := s.CreateTag(t.Context(), "snack", "", "", user.ID)

	// record with food+drink
	r1, _ := s.CreateRecord(t.Context(), "2026-01-15", 1000, currency.ID, "", user.ID)
	s.AddRecordTag(t.Context(), r1.ID, food.ID, user.ID)
	s.AddRecordTag(t.Context(), r1.ID, drink.ID, user.ID)

	// record with food only
	r2, _ := s.CreateRecord(t.Context(), "2026-01-16", 2000, currency.ID, "", user.ID)
	s.AddRecordTag(t.Context(), r2.ID, food.ID, user.ID)

	// record with food+drink+snack
	r3, _ := s.CreateRecord(t.Context(), "2026-01-17", 3000, currency.ID, "", user.ID)
	s.AddRecordTag(t.Context(), r3.ID, food.ID, user.ID)
	s.AddRecordTag(t.Context(), r3.ID, drink.ID, user.ID)
	s.AddRecordTag(t.Context(), r3.ID, snack.ID, user.ID)

	// query with food+drink filter -> should match r1 and r3 (not r2)
	results, err := s.QueryRecords(t.Context(), "2026-01-01", "2026-01-31", nil, []int64{food.ID, drink.ID}, "")
	if err != nil {
		t.Fatalf("failed to query records: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 records matching both food+drink, got %d", len(results))
	}

	hasR1, hasR3 := false, false
	for _, r := range results {
		if r.ID == r1.ID {
			hasR1 = true
		}
		if r.ID == r3.ID {
			hasR3 = true
		}
	}
	if !hasR1 {
		t.Error("expected r1 in results")
	}
	if !hasR3 {
		t.Error("expected r3 in results")
	}
}

func TestQueryRecordsFuzzyFilter(t *testing.T) {
	s := setupTestService(t)

	user, _ := s.CreateUser(t.Context(), "testuser")
	currency, _ := s.CreateCurrency(t.Context(), "USD")

	food, _ := s.CreateTag(t.Context(), "groceries", "", "", user.ID)

	s.CreateRecordWithTags(t.Context(), "2026-01-15", 1000, currency.ID, "coffee beans", user.ID, nil)
	s.CreateRecordWithTags(t.Context(), "2026-01-16", 2000, currency.ID, "rent payment", user.ID, nil)
	s.CreateRecordWithTags(t.Context(), "2026-01-17", 3000, currency.ID, "lunch out", user.ID, []int64{food.ID})

	// fuzzy search "coffee" -> notes match
	results, err := s.QueryRecords(t.Context(), "2026-01-01", "2026-01-31", nil, nil, "coffee")
	if err != nil {
		t.Fatalf("failed to query records: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 record matching 'coffee', got %d", len(results))
	}

	// fuzzy search "groceries" -> tag name match
	results, err = s.QueryRecords(t.Context(), "2026-01-01", "2026-01-31", nil, nil, "groceries")
	if err != nil {
		t.Fatalf("failed to query records: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 record matching 'groceries' tag, got %d", len(results))
	}

	// fuzzy search "nonexistent" -> no results
	results, err = s.QueryRecords(t.Context(), "2026-01-01", "2026-01-31", nil, nil, "nonexistent")
	if err != nil {
		t.Fatalf("failed to query records: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 records, got %d", len(results))
	}
}

func TestQueryRecordsEmpty(t *testing.T) {
	s := setupTestService(t)

	results, err := s.QueryRecords(t.Context(), "2026-01-01", "2026-01-31", nil, nil, "")
	if err != nil {
		t.Fatalf("failed to query records: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestQueryRecordsCurrencyIDPopulated(t *testing.T) {
	s := setupTestService(t)

	user, _ := s.CreateUser(t.Context(), "testuser")
	usd, _ := s.CreateCurrency(t.Context(), "USD")

	s.CreateRecord(t.Context(), "2026-01-15", 1000, usd.ID, "", user.ID)

	results, err := s.QueryRecords(t.Context(), "2026-01-01", "2026-01-31", nil, nil, "")
	if err != nil {
		t.Fatalf("failed to query records: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].CurrencyID != usd.ID {
		t.Errorf("expected CurrencyID %d, got %d", usd.ID, results[0].CurrencyID)
	}
}

func TestQueryRecordsSortOrder(t *testing.T) {
	s := setupTestService(t)

	user, _ := s.CreateUser(t.Context(), "testuser")
	currency, _ := s.CreateCurrency(t.Context(), "USD")

	r1, _ := s.CreateRecord(t.Context(), "2026-01-20", 1000, currency.ID, "", user.ID)
	_, _ = s.CreateRecord(t.Context(), "2026-01-15", 2000, currency.ID, "", user.ID)
	_, _ = s.CreateRecord(t.Context(), "2026-01-15", 3000, currency.ID, "", user.ID)

	results, err := s.QueryRecords(t.Context(), "2026-01-01", "2026-01-31", nil, nil, "")
	if err != nil {
		t.Fatalf("failed to query records: %v", err)
	}

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	// sorted: date DESC, created_at DESC
	// r1: Jan 20 (first)
	// r2: Jan 15 created before r3 (second)
	// r3: Jan 15 created after r2 (third)
	if results[0].ID != r1.ID {
		t.Errorf("expected first result ID %d (later date), got %d", r1.ID, results[0].ID)
	}
}

func TestQueryRecordsAllFiltersCombined(t *testing.T) {
	s := setupTestService(t)

	user, _ := s.CreateUser(t.Context(), "testuser")
	usd, _ := s.CreateCurrency(t.Context(), "USD")
	php, _ := s.CreateCurrency(t.Context(), "PHP")
	food, _ := s.CreateTag(t.Context(), "food", "", "", user.ID)
	drink, _ := s.CreateTag(t.Context(), "drink", "", "", user.ID)

	// match target: USD + food+drink + "coffee" in notes
	r1, _ := s.CreateRecord(t.Context(), "2026-01-15", 1000, usd.ID, "coffee beans", user.ID)
	s.AddRecordTag(t.Context(), r1.ID, food.ID, user.ID)
	s.AddRecordTag(t.Context(), r1.ID, drink.ID, user.ID)

	// wrong currency (PHP)
	r2, _ := s.CreateRecord(t.Context(), "2026-01-16", 2000, php.ID, "coffee latte", user.ID)
	s.AddRecordTag(t.Context(), r2.ID, food.ID, user.ID)
	s.AddRecordTag(t.Context(), r2.ID, drink.ID, user.ID)
	_ = r2

	// missing drink tag
	r3, _ := s.CreateRecord(t.Context(), "2026-01-17", 3000, usd.ID, "coffee shop", user.ID)
	s.AddRecordTag(t.Context(), r3.ID, food.ID, user.ID)
	_ = r3

	// no fuzzy match
	r4, _ := s.CreateRecord(t.Context(), "2026-01-18", 4000, usd.ID, "rent", user.ID)
	s.AddRecordTag(t.Context(), r4.ID, food.ID, user.ID)
	s.AddRecordTag(t.Context(), r4.ID, drink.ID, user.ID)
	_ = r4

	usdID := usd.ID
	results, err := s.QueryRecords(t.Context(), "2026-01-01", "2026-01-31", &usdID, []int64{food.ID, drink.ID}, "coffee")
	if err != nil {
		t.Fatalf("failed to query records: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("expected 1 result matching all filters, got %d", len(results))
	}
	if len(results) > 0 && results[0].ID != r1.ID {
		t.Errorf("expected result ID %d, got %d", r1.ID, results[0].ID)
	}
}
