package service_test

import (
	"testing"
)

func TestSavedQueryCRUD(t *testing.T) {
	s := setupTestService(t)

	user, err := s.CreateUser(t.Context(), "testuser")
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	query, err := s.CreateSavedQuery(t.Context(), "monthly report", "2026-05-01", "2026-05-31", nil, "coffee", user.ID)
	if err != nil {
		t.Fatalf("failed to create saved query: %v", err)
	}

	if query.ID == 0 {
		t.Error("expected query ID to be set")
	}
	if query.Name != "monthly report" {
		t.Errorf("expected name 'monthly report', got '%s'", query.Name)
	}

	fetched, err := s.GetSavedQuery(t.Context(), query.ID)
	if err != nil {
		t.Fatalf("failed to get saved query: %v", err)
	}
	if fetched.Name != query.Name {
		t.Errorf("expected name '%s', got '%s'", query.Name, fetched.Name)
	}

	queries, err := s.ListSavedQueries(t.Context())
	if err != nil {
		t.Fatalf("failed to list saved queries: %v", err)
	}
	if len(queries) != 1 {
		t.Errorf("expected 1 saved query, got %d", len(queries))
	}

	err = s.DeleteSavedQuery(t.Context(), query.ID)
	if err != nil {
		t.Fatalf("failed to delete saved query: %v", err)
	}

	queries, err = s.ListSavedQueries(t.Context())
	if err != nil {
		t.Fatalf("failed to list saved queries: %v", err)
	}
	if len(queries) != 0 {
		t.Errorf("expected 0 saved queries after delete, got %d", len(queries))
	}
}

func TestSavedQueryWithTags(t *testing.T) {
	s := setupTestService(t)

	user, err := s.CreateUser(t.Context(), "testuser")
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	tag1, err := s.CreateTag(t.Context(), "food", "", "", user.ID)
	if err != nil {
		t.Fatalf("failed to create tag1: %v", err)
	}
	tag2, err := s.CreateTag(t.Context(), "drink", "", "", user.ID)
	if err != nil {
		t.Fatalf("failed to create tag2: %v", err)
	}

	currency, err := s.CreateCurrency(t.Context(), "USD")
	if err != nil {
		t.Fatalf("failed to create currency: %v", err)
	}

	currencyID := currency.ID
	query, err := s.CreateSavedQueryWithTags(t.Context(), "groceries", "2026-05-01", "2026-05-31", &currencyID, "", user.ID, []int64{tag1.ID, tag2.ID})
	if err != nil {
		t.Fatalf("failed to create saved query with tags: %v", err)
	}

	if query.ID == 0 {
		t.Error("expected query ID to be set")
	}

	tags, err := s.GetSavedQueryTags(t.Context(), query.ID)
	if err != nil {
		t.Fatalf("failed to get saved query tags: %v", err)
	}
	if len(tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(tags))
	}

	err = s.DeleteSavedQuery(t.Context(), query.ID)
	if err != nil {
		t.Fatalf("failed to delete saved query: %v", err)
	}

	tags, err = s.GetSavedQueryTags(t.Context(), query.ID)
	if err != nil {
		t.Fatalf("failed to get deleted query tags: %v", err)
	}
	if len(tags) != 0 {
		t.Errorf("expected 0 tags after cascade delete, got %d", len(tags))
	}
}
