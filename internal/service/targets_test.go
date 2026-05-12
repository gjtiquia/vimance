package service_test

import (
	"testing"

	service "github.com/gjtiquia/vimance/internal/service"
)

func TestTarget_CRUD(t *testing.T) {
	s := setupTestService(t)

	user, _ := s.CreateUser(t.Context(), "testuser")

	sq, _ := s.CreateSavedQueryWithTags(t.Context(), "food budget", "2026-05-01", "2026-05-31", nil, "", user.ID, nil)

	// Create
	target, err := s.CreateTarget(t.Context(), "food budget", sq.ID, -50000, user.ID)
	if err != nil {
		t.Fatalf("CreateTarget() error: %v", err)
	}
	if target.Name != "food budget" {
		t.Errorf("expected name 'food budget', got %q", target.Name)
	}
	if target.SavedQueryID != sq.ID {
		t.Errorf("expected saved_query_id %d, got %d", sq.ID, target.SavedQueryID)
	}
	if target.TargetCents != -50000 {
		t.Errorf("expected target_cents -50000, got %d", target.TargetCents)
	}

	// Get
	got, err := s.GetTarget(t.Context(), target.ID)
	if err != nil {
		t.Fatalf("GetTarget() error: %v", err)
	}
	if got.Name != target.Name {
		t.Errorf("expected name %q, got %q", target.Name, got.Name)
	}

	// List
	targets, err := s.ListTargets(t.Context())
	if err != nil {
		t.Fatalf("ListTargets() error: %v", err)
	}
	if len(targets) != 1 {
		t.Errorf("expected 1 target, got %d", len(targets))
	}

	// Update
	_, err = s.UpdateTarget(t.Context(), target.ID, "food budget updated", sq.ID, -60000, user.ID)
	if err != nil {
		t.Fatalf("UpdateTarget() error: %v", err)
	}
	got, _ = s.GetTarget(t.Context(), target.ID)
	if got.Name != "food budget updated" {
		t.Errorf("expected name 'food budget updated', got %q", got.Name)
	}
	if got.TargetCents != -60000 {
		t.Errorf("expected target_cents -60000, got %d", got.TargetCents)
	}
}

func TestTarget_WithActual_HasData(t *testing.T) {
	s := setupTestService(t)

	user, _ := s.CreateUser(t.Context(), "testuser")
	currency, _ := s.CreateCurrency(t.Context(), "USD")
	food, _ := s.CreateTag(t.Context(), "food", "", "", user.ID)

	// Create records: food totaling -$320 (32000 cents)
	r1, _ := s.CreateRecord(t.Context(), "2026-05-01", -20000, currency.ID, "groceries", user.ID)
	s.AddRecordTag(t.Context(), r1.ID, food.ID, user.ID)
	r2, _ := s.CreateRecord(t.Context(), "2026-05-05", -12000, currency.ID, "lunch", user.ID)
	s.AddRecordTag(t.Context(), r2.ID, food.ID, user.ID)

	// Create saved query for food in May
	sq, _ := s.CreateSavedQueryWithTags(t.Context(), "food budget", "2026-05-01", "2026-05-31", nil, "", user.ID, []int64{food.ID})

	// Create target
	target, _ := s.CreateTarget(t.Context(), "food budget", sq.ID, -50000, user.ID)

	// Get with actual
	result, err := s.GetTargetWithActual(t.Context(), target.ID)
	if err != nil {
		t.Fatalf("GetTargetWithActual() error: %v", err)
	}

	if !result.HasData {
		t.Error("expected HasData=true")
	}
	if result.ActualAmount == nil {
		t.Fatal("expected ActualAmount to be non-nil")
	}
	if *result.ActualAmount != -32000 {
		t.Errorf("expected actual=-32000, got %d", *result.ActualAmount)
	}
	if result.Target.TargetCents != -50000 {
		t.Errorf("expected target=-50000, got %d", result.Target.TargetCents)
	}

	_ = currency
}

func TestTarget_WithActual_NoData(t *testing.T) {
	s := setupTestService(t)

	user, _ := s.CreateUser(t.Context(), "testuser")

	// Create saved query for a tag with no records
	food, _ := s.CreateTag(t.Context(), "food", "", "", user.ID)
	sq, _ := s.CreateSavedQueryWithTags(t.Context(), "food budget", "2026-05-01", "2026-05-31", nil, "", user.ID, []int64{food.ID})

	target, _ := s.CreateTarget(t.Context(), "food budget", sq.ID, -50000, user.ID)

	result, err := s.GetTargetWithActual(t.Context(), target.ID)
	if err != nil {
		t.Fatalf("GetTargetWithActual() error: %v", err)
	}

	if result.HasData {
		t.Error("expected HasData=false for no matching records")
	}
	if result.ActualAmount != nil {
		t.Errorf("expected ActualAmount=nil, got %d", *result.ActualAmount)
	}
}

func TestTarget_WithActual_TotalZero(t *testing.T) {
	s := setupTestService(t)

	user, _ := s.CreateUser(t.Context(), "testuser")
	currency, _ := s.CreateCurrency(t.Context(), "USD")
	food, _ := s.CreateTag(t.Context(), "food", "", "", user.ID)

	// Create expense and exact income refund in food
	r1, _ := s.CreateRecord(t.Context(), "2026-05-01", -5000, currency.ID, "groceries", user.ID)
	s.AddRecordTag(t.Context(), r1.ID, food.ID, user.ID)
	r2, _ := s.CreateRecord(t.Context(), "2026-05-02", 5000, currency.ID, "refund", user.ID)
	s.AddRecordTag(t.Context(), r2.ID, food.ID, user.ID)

	sq, _ := s.CreateSavedQueryWithTags(t.Context(), "food net", "2026-05-01", "2026-05-31", nil, "", user.ID, []int64{food.ID})

	target, _ := s.CreateTarget(t.Context(), "food net zero", sq.ID, 0, user.ID)

	result, _ := s.GetTargetWithActual(t.Context(), target.ID)

	if !result.HasData {
		t.Error("expected HasData=true when records exist")
	}
	if result.ActualAmount == nil {
		t.Fatal("expected ActualAmount to be non-nil")
	}
	if *result.ActualAmount != 0 {
		t.Errorf("expected actual=0, got %d", *result.ActualAmount)
	}
}

func TestTarget_CascadeDelete(t *testing.T) {
	s := setupTestService(t)

	user, _ := s.CreateUser(t.Context(), "testuser")

	sq, _ := s.CreateSavedQueryWithTags(t.Context(), "test query", "2026-05-01", "2026-05-31", nil, "", user.ID, nil)
	target, _ := s.CreateTarget(t.Context(), "test target", sq.ID, -50000, user.ID)

	// Delete the saved query
	err := s.DeleteSavedQuery(t.Context(), sq.ID)
	if err != nil {
		t.Fatalf("DeleteSavedQuery() error: %v", err)
	}

	// Target should be cascade-deleted
	_, err = s.GetTarget(t.Context(), target.ID)
	if err == nil {
		t.Error("expected error getting cascade-deleted target, got nil")
	}
}

func TestTarget_ListWithActuals(t *testing.T) {
	s := setupTestService(t)

	user, _ := s.CreateUser(t.Context(), "testuser")
	currency, _ := s.CreateCurrency(t.Context(), "USD")
	food, _ := s.CreateTag(t.Context(), "food", "", "", user.ID)
	transport, _ := s.CreateTag(t.Context(), "transport", "", "", user.ID)

	// Create food records
	r1, _ := s.CreateRecord(t.Context(), "2026-05-01", -32000, currency.ID, "groceries", user.ID)
	s.AddRecordTag(t.Context(), r1.ID, food.ID, user.ID)

	// Food target (has data)
	sq1, _ := s.CreateSavedQueryWithTags(t.Context(), "food budget", "2026-05-01", "2026-05-31", nil, "", user.ID, []int64{food.ID})
	s.CreateTarget(t.Context(), "food budget", sq1.ID, -50000, user.ID)

	// Transport target (no data)
	sq2, _ := s.CreateSavedQueryWithTags(t.Context(), "transport budget", "2026-05-01", "2026-05-31", nil, "", user.ID, []int64{transport.ID})
	s.CreateTarget(t.Context(), "transport budget", sq2.ID, -20000, user.ID)

	results, err := s.ListTargetsWithActuals(t.Context())
	if err != nil {
		t.Fatalf("ListTargetsWithActuals() error: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 targets, got %d", len(results))
	}

	// Find food and transport targets
	var foodResult, transportResult *service.TargetWithActual
	for i := range results {
		if results[i].Target.Name == "food budget" {
			foodResult = &results[i]
		} else if results[i].Target.Name == "transport budget" {
			transportResult = &results[i]
		}
	}

	if foodResult == nil {
		t.Fatal("expected to find food budget target")
	}
	if !foodResult.HasData {
		t.Error("expected food budget HasData=true")
	}
	if foodResult.ActualAmount == nil || *foodResult.ActualAmount != -32000 {
		t.Errorf("expected food actual=-32000, got %v", foodResult.ActualAmount)
	}

	if transportResult == nil {
		t.Fatal("expected to find transport budget target")
	}
	if transportResult.HasData {
		t.Error("expected transport budget HasData=false (no data)")
	}
	if transportResult.ActualAmount != nil {
		t.Errorf("expected transport ActualAmount=nil, got %d", *transportResult.ActualAmount)
	}
}

func TestTarget_CumulativeScope(t *testing.T) {
	s := setupTestService(t)

	user, _ := s.CreateUser(t.Context(), "testuser")
	currency, _ := s.CreateCurrency(t.Context(), "USD")
	savings, _ := s.CreateTag(t.Context(), "savings", "", "", user.ID)

	// Create savings records across multiple months (cumulative)
	s.CreateRecordWithTags(t.Context(), "2026-01-15", 100000, currency.ID, "", user.ID, []int64{savings.ID})
	s.CreateRecordWithTags(t.Context(), "2026-02-15", 100000, currency.ID, "", user.ID, []int64{savings.ID})
	s.CreateRecordWithTags(t.Context(), "2026-03-15", 100000, currency.ID, "", user.ID, []int64{savings.ID})

	// Saved query with wide date range (all-time savings)
	sq, _ := s.CreateSavedQueryWithTags(t.Context(), "all savings", "2026-01-01", "2026-12-31", nil, "", user.ID, []int64{savings.ID})

	target, _ := s.CreateTarget(t.Context(), "savings goal", sq.ID, 1000000, user.ID)

	result, _ := s.GetTargetWithActual(t.Context(), target.ID)

	if !result.HasData {
		t.Fatal("expected HasData=true")
	}
	if result.ActualAmount == nil {
		t.Fatal("expected ActualAmount to be non-nil")
	}
	// 3 x 100000 = 300000 cents
	if *result.ActualAmount != 300000 {
		t.Errorf("expected cumulative actual=300000, got %d", *result.ActualAmount)
	}
}

func TestTarget_UpdateAmount(t *testing.T) {
	s := setupTestService(t)

	user, _ := s.CreateUser(t.Context(), "testuser")
	sq, _ := s.CreateSavedQueryWithTags(t.Context(), "test query", "2026-05-01", "2026-05-31", nil, "", user.ID, nil)

	target, _ := s.CreateTarget(t.Context(), "budget", sq.ID, -50000, user.ID)

	updated, err := s.UpdateTarget(t.Context(), target.ID, "budget updated", sq.ID, -60000, user.ID)
	if err != nil {
		t.Fatalf("UpdateTarget() error: %v", err)
	}
	if updated.TargetCents != -60000 {
		t.Errorf("expected target_cents=-60000, got %d", updated.TargetCents)
	}
	if updated.Name != "budget updated" {
		t.Errorf("expected name='budget updated', got %q", updated.Name)
	}

	// Verify persisted
	got, _ := s.GetTarget(t.Context(), target.ID)
	if got.TargetCents != -60000 {
		t.Errorf("expected persisted target_cents=-60000, got %d", got.TargetCents)
	}
}

func TestTarget_Delete(t *testing.T) {
	s := setupTestService(t)

	user, _ := s.CreateUser(t.Context(), "testuser")
	sq, _ := s.CreateSavedQueryWithTags(t.Context(), "test query", "2026-05-01", "2026-05-31", nil, "", user.ID, nil)

	target, _ := s.CreateTarget(t.Context(), "budget", sq.ID, -50000, user.ID)

	err := s.DeleteTarget(t.Context(), target.ID)
	if err != nil {
		t.Fatalf("DeleteTarget() error: %v", err)
	}

	_, err = s.GetTarget(t.Context(), target.ID)
	if err == nil {
		t.Error("expected error getting deleted target")
	}
}