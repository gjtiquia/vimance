package service_test

import (
	"testing"
)

func TestLinkRecords(t *testing.T) {
	s := setupTestService(t)

	user, err := s.CreateUser(t.Context(), "testuser")
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	currency, err := s.CreateCurrency(t.Context(), "USD")
	if err != nil {
		t.Fatalf("failed to create currency: %v", err)
	}

	parent, err := s.CreateRecord(t.Context(), "2026-01-01", 500000, currency.ID, "january balance", user.ID)
	if err != nil {
		t.Fatalf("failed to create parent record: %v", err)
	}

	child, err := s.CreateRecord(t.Context(), "2026-01-15", 120000, currency.ID, "credit card bill", user.ID)
	if err != nil {
		t.Fatalf("failed to create child record: %v", err)
	}

	err = s.LinkRecords(t.Context(), parent.ID, child.ID, user.ID)
	if err != nil {
		t.Fatalf("failed to link records: %v", err)
	}

	parents, err := s.GetRecordParents(t.Context(), child.ID)
	if err != nil {
		t.Fatalf("failed to get record parents: %v", err)
	}
	if len(parents) != 1 {
		t.Errorf("expected 1 parent, got %d", len(parents))
	}
	if parents[0].ID != parent.ID {
		t.Errorf("expected parent ID %d, got %d", parent.ID, parents[0].ID)
	}

	children, err := s.GetRecordChildren(t.Context(), parent.ID)
	if err != nil {
		t.Fatalf("failed to get record children: %v", err)
	}
	if len(children) != 1 {
		t.Errorf("expected 1 child, got %d", len(children))
	}
	if children[0].ID != child.ID {
		t.Errorf("expected child ID %d, got %d", child.ID, children[0].ID)
	}
}

func TestMultipleParents(t *testing.T) {
	s := setupTestService(t)

	user, _ := s.CreateUser(t.Context(), "testuser")
	currency, _ := s.CreateCurrency(t.Context(), "USD")

	parent1, _ := s.CreateRecord(t.Context(), "2026-01-01", 500000, currency.ID, "january balance", user.ID)
	parent2, _ := s.CreateRecord(t.Context(), "2026-02-01", 520000, currency.ID, "february balance", user.ID)
	child, _ := s.CreateRecord(t.Context(), "2026-01-15", 120000, currency.ID, "credit card bill", user.ID)

	if err := s.LinkRecords(t.Context(), parent1.ID, child.ID, user.ID); err != nil {
		t.Fatalf("failed to link parent1: %v", err)
	}
	if err := s.LinkRecords(t.Context(), parent2.ID, child.ID, user.ID); err != nil {
		t.Fatalf("failed to link parent2: %v", err)
	}

	parents, err := s.GetRecordParents(t.Context(), child.ID)
	if err != nil {
		t.Fatalf("failed to get parents: %v", err)
	}
	if len(parents) != 2 {
		t.Errorf("expected 2 parents, got %d", len(parents))
	}
}

func TestUnlinkRecords(t *testing.T) {
	s := setupTestService(t)

	user, _ := s.CreateUser(t.Context(), "testuser")
	currency, _ := s.CreateCurrency(t.Context(), "USD")

	parent, _ := s.CreateRecord(t.Context(), "2026-01-01", 500000, currency.ID, "january balance", user.ID)
	child, _ := s.CreateRecord(t.Context(), "2026-01-15", 120000, currency.ID, "credit card bill", user.ID)

	s.LinkRecords(t.Context(), parent.ID, child.ID, user.ID)

	err := s.UnlinkRecords(t.Context(), parent.ID, child.ID)
	if err != nil {
		t.Fatalf("failed to unlink records: %v", err)
	}

	parents, _ := s.GetRecordParents(t.Context(), child.ID)
	if len(parents) != 0 {
		t.Errorf("expected 0 parents after unlink, got %d", len(parents))
	}
}

func TestSearchLinkCandidates(t *testing.T) {
	s := setupTestService(t)

	user, _ := s.CreateUser(t.Context(), "testuser")
	currency1, _ := s.CreateCurrency(t.Context(), "USD")
	currency2, _ := s.CreateCurrency(t.Context(), "PHP")

	tag1, _ := s.CreateTag(t.Context(), "balance", "", "", user.ID)
	tag2, _ := s.CreateTag(t.Context(), "expense", "", "", user.ID)

	p1, _ := s.CreateRecord(t.Context(), "2026-01-01", 500000, currency1.ID, "january balance", user.ID)
	p2, _ := s.CreateRecord(t.Context(), "2026-02-01", 520000, currency1.ID, "february balance", user.ID)
	pOtherCurrency, _ := s.CreateRecord(t.Context(), "2026-01-10", 10000, currency2.ID, "php balance", user.ID)

	s.CreateRecordWithTags(t.Context(), "2026-01-15", 1000, currency1.ID, "unrelated", user.ID, []int64{tag1.ID})

	_, _ = s.CreateRecordWithTagsAndLinks(t.Context(), "2026-01-15", 200, currency1.ID, "coffee", user.ID, []int64{tag2.ID}, []int64{p1.ID})

	candidates, err := s.SearchLinkCandidates(t.Context(), "2026-01-01", "2026-01-31", currency1.ID, 0)
	if err != nil {
		t.Fatalf("failed to search link candidates: %v", err)
	}

	found := false
	for _, c := range candidates {
		if c.ID == p1.ID {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected parent p1 to be in candidates")
	}

	for _, c := range candidates {
		if c.ID == pOtherCurrency.ID {
			t.Error("did not expect record with different currency in candidates")
		}
	}

	for _, c := range candidates {
		if c.ID == p2.ID {
			t.Error("did not expect february record in january date range")
		}
	}
}

func TestCreateRecordWithTagsAndLinks(t *testing.T) {
	s := setupTestService(t)

	user, _ := s.CreateUser(t.Context(), "testuser")
	currency, _ := s.CreateCurrency(t.Context(), "USD")
	tag, _ := s.CreateTag(t.Context(), "expense", "", "", user.ID)

	parent, _ := s.CreateRecord(t.Context(), "2026-01-01", 500000, currency.ID, "january balance", user.ID)

	child, err := s.CreateRecordWithTagsAndLinks(t.Context(), "2026-01-15", 120000, currency.ID, "credit card bill", user.ID, []int64{tag.ID}, []int64{parent.ID})
	if err != nil {
		t.Fatalf("failed to create record with tags and links: %v", err)
	}

	parents, _ := s.GetRecordParents(t.Context(), child.ID)
	if len(parents) != 1 {
		t.Errorf("expected 1 parent, got %d", len(parents))
	}

	tags, _ := s.GetRecordTags(t.Context(), child.ID)
	if len(tags) != 1 {
		t.Errorf("expected 1 tag, got %d", len(tags))
	}
}

func TestCascadeDeleteRecordLinks(t *testing.T) {
	s := setupTestService(t)

	user, _ := s.CreateUser(t.Context(), "testuser")
	currency, _ := s.CreateCurrency(t.Context(), "USD")

	parent, _ := s.CreateRecord(t.Context(), "2026-01-01", 500000, currency.ID, "january balance", user.ID)
	child, _ := s.CreateRecord(t.Context(), "2026-01-15", 120000, currency.ID, "credit card bill", user.ID)

	s.LinkRecords(t.Context(), parent.ID, child.ID, user.ID)

	s.HardDeleteRecord(t.Context(), parent.ID)

	parents, _ := s.GetRecordParents(t.Context(), child.ID)
	if len(parents) != 0 {
		t.Errorf("expected 0 parents after cascade delete, got %d", len(parents))
	}

	children, _ := s.GetRecordChildren(t.Context(), parent.ID)
	if len(children) != 0 {
		t.Errorf("expected 0 children after cascade delete, got %d", len(children))
	}
}
