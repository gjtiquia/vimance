package service_test

import (
	"testing"
)

func TestRecordCRUD(t *testing.T) {
	s := setupTestService(t)

	user, _ := s.CreateUser(t.Context(), "testuser")
	currency, _ := s.CreateCurrency(t.Context(), "USD")

	record, err := s.CreateRecord(t.Context(), "2026-01-15", 1000, currency.ID, "test note", user.ID)
	if err != nil {
		t.Fatalf("failed to create record: %v", err)
	}

	if record.ID == 0 {
		t.Error("expected record ID to be set")
	}
	if record.Date != "2026-01-15" {
		t.Errorf("expected date '2026-01-15', got '%s'", record.Date)
	}
	if record.AmountCents != 1000 {
		t.Errorf("expected amount 1000, got %d", record.AmountCents)
	}

	fetched, err := s.GetRecord(t.Context(), record.ID)
	if err != nil {
		t.Fatalf("failed to get record: %v", err)
	}
	if fetched.Date != record.Date {
		t.Errorf("expected date '%s', got '%s'", record.Date, fetched.Date)
	}

	records, err := s.ListRecords(t.Context())
	if err != nil {
		t.Fatalf("failed to list records: %v", err)
	}
	if len(records) != 1 {
		t.Errorf("expected 1 record, got %d", len(records))
	}

	updated, err := s.UpdateRecord(t.Context(), record.ID, "2026-01-20", 2000, currency.ID, "updated note", user.ID)
	if err != nil {
		t.Fatalf("failed to update record: %v", err)
	}
	if updated.Date != "2026-01-20" {
		t.Errorf("expected date '2026-01-20', got '%s'", updated.Date)
	}
	if updated.AmountCents != 2000 {
		t.Errorf("expected amount 2000, got %d", updated.AmountCents)
	}

	err = s.SoftDeleteRecord(t.Context(), record.ID, user.ID)
	if err != nil {
		t.Fatalf("failed to soft delete record: %v", err)
	}

	activeRecords, err := s.ListActiveRecords(t.Context())
	if err != nil {
		t.Fatalf("failed to list active records: %v", err)
	}
	for _, r := range activeRecords {
		if r.ID == record.ID {
			t.Error("soft deleted record should not appear in active_records view")
		}
	}

	_, err = s.RestoreRecord(t.Context(), record.ID)
	if err != nil {
		t.Fatalf("failed to restore record: %v", err)
	}

	activeRecords, err = s.ListActiveRecords(t.Context())
	if err != nil {
		t.Fatalf("failed to list active records: %v", err)
	}
	found := false
	for _, r := range activeRecords {
		if r.ID == record.ID {
			found = true
			break
		}
	}
	if !found {
		t.Error("restored record should appear in active_records view")
	}

	err = s.HardDeleteRecord(t.Context(), record.ID)
	if err != nil {
		t.Fatalf("failed to hard delete record: %v", err)
	}

	_, err = s.GetRecord(t.Context(), record.ID)
	if err == nil {
		t.Error("expected error when getting hard deleted record")
	}
}

func TestCreateRecordWithTags(t *testing.T) {
	s := setupTestService(t)

	user, _ := s.CreateUser(t.Context(), "testuser")
	currency, _ := s.CreateCurrency(t.Context(), "USD")
	tag1, _ := s.CreateTag(t.Context(), "food", "", "", user.ID)
	tag2, _ := s.CreateTag(t.Context(), "dinner", "", "", user.ID)

	record, err := s.CreateRecordWithTags(t.Context(), "2026-01-15", 1000, currency.ID, "test note", user.ID, []int64{tag1.ID, tag2.ID})
	if err != nil {
		t.Fatalf("failed to create record with tags: %v", err)
	}

	tags, err := s.GetRecordTags(t.Context(), record.ID)
	if err != nil {
		t.Fatalf("failed to get record tags: %v", err)
	}
	if len(tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(tags))
	}
}

func TestUpdateRecordWithTags(t *testing.T) {
	s := setupTestService(t)

	user, _ := s.CreateUser(t.Context(), "testuser")
	currency, _ := s.CreateCurrency(t.Context(), "USD")
	tag1, _ := s.CreateTag(t.Context(), "food", "", "", user.ID)
	tag2, _ := s.CreateTag(t.Context(), "dinner", "", "", user.ID)
	tag3, _ := s.CreateTag(t.Context(), "lunch", "", "", user.ID)

	record, _ := s.CreateRecordWithTags(t.Context(), "2026-01-15", 1000, currency.ID, "test note", user.ID, []int64{tag1.ID, tag2.ID})

	updated, err := s.UpdateRecordWithTags(t.Context(), record.ID, "2026-01-20", 2000, currency.ID, "updated note", user.ID, []int64{tag2.ID, tag3.ID})
	if err != nil {
		t.Fatalf("failed to update record with tags: %v", err)
	}

	if updated.Date != "2026-01-20" {
		t.Errorf("expected date '2026-01-20', got '%s'", updated.Date)
	}

	tags, err := s.GetRecordTags(t.Context(), record.ID)
	if err != nil {
		t.Fatalf("failed to get record tags: %v", err)
	}
	if len(tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(tags))
	}

	tagIDs := make(map[int64]bool)
	for _, tag := range tags {
		tagIDs[tag.ID] = true
	}
	if tagIDs[tag1.ID] {
		t.Error("tag1 should have been removed")
	}
	if !tagIDs[tag2.ID] {
		t.Error("tag2 should still be present")
	}
	if !tagIDs[tag3.ID] {
		t.Error("tag3 should have been added")
	}
}

func TestRecordDateRange(t *testing.T) {
	s := setupTestService(t)

	user, _ := s.CreateUser(t.Context(), "testuser")
	currency, _ := s.CreateCurrency(t.Context(), "USD")

	s.CreateRecord(t.Context(), "2026-01-10", 1000, currency.ID, "", user.ID)
	s.CreateRecord(t.Context(), "2026-01-15", 2000, currency.ID, "", user.ID)
	s.CreateRecord(t.Context(), "2026-01-20", 3000, currency.ID, "", user.ID)
	s.CreateRecord(t.Context(), "2026-01-25", 4000, currency.ID, "", user.ID)

	records, err := s.ListActiveRecordsByDateRange(t.Context(), "2026-01-12", "2026-01-22")
	if err != nil {
		t.Fatalf("failed to list records by date range: %v", err)
	}
	if len(records) != 2 {
		t.Errorf("expected 2 records in range, got %d", len(records))
	}
}

func TestAddRemoveRecordTag(t *testing.T) {
	s := setupTestService(t)

	user, _ := s.CreateUser(t.Context(), "testuser")
	currency, _ := s.CreateCurrency(t.Context(), "USD")
	tag, _ := s.CreateTag(t.Context(), "food", "", "", user.ID)

	record, _ := s.CreateRecord(t.Context(), "2026-01-15", 1000, currency.ID, "", user.ID)

	err := s.AddRecordTag(t.Context(), record.ID, tag.ID, user.ID)
	if err != nil {
		t.Fatalf("failed to add record tag: %v", err)
	}

	tags, _ := s.GetRecordTags(t.Context(), record.ID)
	if len(tags) != 1 {
		t.Errorf("expected 1 tag, got %d", len(tags))
	}

	err = s.RemoveRecordTag(t.Context(), record.ID, tag.ID)
	if err != nil {
		t.Fatalf("failed to remove record tag: %v", err)
	}

	tags, _ = s.GetRecordTags(t.Context(), record.ID)
	if len(tags) != 0 {
		t.Errorf("expected 0 tags, got %d", len(tags))
	}
}

func TestDuplicateRecordTag(t *testing.T) {
	s := setupTestService(t)

	user, _ := s.CreateUser(t.Context(), "testuser")
	currency, _ := s.CreateCurrency(t.Context(), "USD")
	tag, _ := s.CreateTag(t.Context(), "food", "", "", user.ID)

	record, _ := s.CreateRecord(t.Context(), "2026-01-15", 1000, currency.ID, "", user.ID)

	err := s.AddRecordTag(t.Context(), record.ID, tag.ID, user.ID)
	if err != nil {
		t.Fatalf("failed to add record tag: %v", err)
	}

	err = s.AddRecordTag(t.Context(), record.ID, tag.ID, user.ID)
	if err == nil {
		t.Error("expected error when adding duplicate record-tag pair")
	}
}

func TestGetRecordFull(t *testing.T) {
	s := setupTestService(t)

	user, _ := s.CreateUser(t.Context(), "testuser")
	currency, _ := s.CreateCurrency(t.Context(), "USD")
	tag1, _ := s.CreateTag(t.Context(), "food", "", "", user.ID)
	tag2, _ := s.CreateTag(t.Context(), "drink", "", "", user.ID)

	parent, _ := s.CreateRecord(t.Context(), "2026-01-01", 500000, currency.ID, "january balance", user.ID)
	child, _ := s.CreateRecordWithTags(t.Context(), "2026-01-15", 120000, currency.ID, "credit card bill", user.ID, []int64{tag1.ID, tag2.ID})
	s.LinkRecords(t.Context(), parent.ID, child.ID, user.ID)

	full, err := s.GetRecordFull(t.Context(), child.ID)
	if err != nil {
		t.Fatalf("failed to get record full: %v", err)
	}

	if full.Record.ID != child.ID {
		t.Errorf("expected record ID %d, got %d", child.ID, full.Record.ID)
	}
	if full.CurrencyCode != "USD" {
		t.Errorf("expected currency code USD, got %s", full.CurrencyCode)
	}
	if len(full.Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(full.Tags))
	}
	if len(full.Parents) != 1 {
		t.Errorf("expected 1 parent, got %d", len(full.Parents))
	}
	if full.Parents[0].ID != parent.ID {
		t.Errorf("expected parent ID %d, got %d", parent.ID, full.Parents[0].ID)
	}
}

func TestUpdateRecordWithTagsAndLinks(t *testing.T) {
	s := setupTestService(t)

	user, _ := s.CreateUser(t.Context(), "testuser")
	currency, _ := s.CreateCurrency(t.Context(), "USD")
	tag1, _ := s.CreateTag(t.Context(), "food", "", "", user.ID)
	tag2, _ := s.CreateTag(t.Context(), "drink", "", "", user.ID)

	parent, _ := s.CreateRecord(t.Context(), "2026-01-01", 500000, currency.ID, "january", user.ID)
	child, _ := s.CreateRecordWithTags(t.Context(), "2026-01-15", 120000, currency.ID, "bill", user.ID, []int64{tag1.ID})
	s.LinkRecords(t.Context(), parent.ID, child.ID, user.ID)

	// update: change tags (remove tag1, add tag2), change links (remove parent)
	updated, err := s.UpdateRecordWithTagsAndLinks(t.Context(), child.ID, "2026-01-20", 200000, currency.ID, "updated bill", user.ID, []int64{tag2.ID}, nil)
	if err != nil {
		t.Fatalf("failed to update record: %v", err)
	}

	if updated.Date != "2026-01-20" {
		t.Errorf("expected date '2026-01-20', got '%s'", updated.Date)
	}
	if updated.AmountCents != 200000 {
		t.Errorf("expected amount 200000, got %d", updated.AmountCents)
	}

	// verify old tag removed
	tags, _ := s.GetRecordTags(t.Context(), child.ID)
	if len(tags) != 1 {
		t.Errorf("expected 1 tag after update, got %d", len(tags))
	}
	if tags[0].ID != tag2.ID {
		t.Errorf("expected tag2 (%d) after update, got %d", tag2.ID, tags[0].ID)
	}

	// verify old link removed
	parents, _ := s.GetRecordParents(t.Context(), child.ID)
	if len(parents) != 0 {
		t.Errorf("expected 0 parents after link removal, got %d", len(parents))
	}
}
