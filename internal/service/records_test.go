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
