package service_test

import (
	"testing"

	service "github.com/gjtiquia/vimance/internal/service"
)

// Journey 1: "I just installed, now what?" → empty DB, then first record
func TestJourney_FirstRecord(t *testing.T) {
	s := setupTestService(t)

	// Empty DB: query returns HasData=false
	result, err := s.Aggregate(t.Context(), "2026-01-01", "2026-12-31", nil, nil, "")
	if err != nil {
		t.Fatalf("Aggregate() error: %v", err)
	}
	if result.HasData {
		t.Error("expected HasData=false for empty DB")
	}
	if result.RecordCount != 0 {
		t.Errorf("expected 0 records, got %d", result.RecordCount)
	}

	// Create first record
	user, _ := s.CreateUser(t.Context(), "testuser")
	currency, _ := s.CreateCurrency(t.Context(), "USD")
	checking, _ := s.CreateTag(t.Context(), "checking", "", "", user.ID)

	s.CreateRecordWithTags(t.Context(), "2026-05-01", 500000, currency.ID, "approx balance", user.ID, []int64{checking.ID})

	// Query now returns valid data
	result2, err := s.Aggregate(t.Context(), "2026-05-01", "2026-05-31", nil, nil, "")
	if err != nil {
		t.Fatalf("Aggregate() error: %v", err)
	}
	if !result2.HasData {
		t.Error("expected HasData=true after creating record")
	}
	if result2.RecordCount != 1 {
		t.Errorf("expected 1 record, got %d", result2.RecordCount)
	}
	if result2.TotalAmount != 500000 {
		t.Errorf("expected TotalAmount=500000, got %d", result2.TotalAmount)
	}
}

// Journey 3: "How much did I spend this month?" → aggregation with mixed income/expense
func TestJourney_MonthlySpending(t *testing.T) {
	s := setupTestService(t)

	user, _ := s.CreateUser(t.Context(), "testuser")
	currency, _ := s.CreateCurrency(t.Context(), "USD")
	food, _ := s.CreateTag(t.Context(), "food", "", "", user.ID)
	rent, _ := s.CreateTag(t.Context(), "rent", "", "", user.ID)
	salary, _ := s.CreateTag(t.Context(), "salary", "", "", user.ID)

	// Expenses
	r1, _ := s.CreateRecord(t.Context(), "2026-05-01", -100000, currency.ID, "rent", user.ID)
	s.AddRecordTag(t.Context(), r1.ID, rent.ID, user.ID)
	r2, _ := s.CreateRecord(t.Context(), "2026-05-05", -5000, currency.ID, "lunch", user.ID)
	s.AddRecordTag(t.Context(), r2.ID, food.ID, user.ID)
	r3, _ := s.CreateRecord(t.Context(), "2026-05-10", -3200, currency.ID, "groceries", user.ID)
	s.AddRecordTag(t.Context(), r3.ID, food.ID, user.ID)

	// Income
	r4, _ := s.CreateRecord(t.Context(), "2026-05-15", 350000, currency.ID, "salary", user.ID)
	s.AddRecordTag(t.Context(), r4.ID, salary.ID, user.ID)

	result, err := s.Aggregate(t.Context(), "2026-05-01", "2026-05-31", nil, nil, "")
	if err != nil {
		t.Fatalf("Aggregate() error: %v", err)
	}

	if result.RecordCount != 4 {
		t.Errorf("expected 4 records, got %d", result.RecordCount)
	}
	// Total = -100000 - 5000 - 3200 + 350000 = 241800
	if result.TotalAmount != 241800 {
		t.Errorf("expected TotalAmount=241800, got %d", result.TotalAmount)
	}
	if result.IncomeSum != 350000 {
		t.Errorf("expected IncomeSum=350000, got %d", result.IncomeSum)
	}
	if result.ExpenseSum != -108200 {
		t.Errorf("expected ExpenseSum=-108200, got %d", result.ExpenseSum)
	}

	// Check tag breakdown
	tagMap := make(map[string]int64)
	for _, ts := range result.ByTag {
		tagMap[ts.TagName] = ts.Amount
	}
	if tagMap["food"] != -8200 {
		t.Errorf("expected food=-8200, got %d", tagMap["food"])
	}
	if tagMap["rent"] != -100000 {
		t.Errorf("expected rent=-100000, got %d", tagMap["rent"])
	}
	if tagMap["salary"] != 350000 {
		t.Errorf("expected salary=350000, got %d", tagMap["salary"])
	}
}

// Journey 14: "I want to budget $500/month for food" → full budget tracking flow
func TestJourney_BudgetTracking(t *testing.T) {
	s := setupTestService(t)

	user, _ := s.CreateUser(t.Context(), "testuser")
	currency, _ := s.CreateCurrency(t.Context(), "USD")
	food, _ := s.CreateTag(t.Context(), "food", "", "", user.ID)

	// Create food records totaling -$320
	r1, _ := s.CreateRecord(t.Context(), "2026-05-01", -20000, currency.ID, "groceries", user.ID)
	s.AddRecordTag(t.Context(), r1.ID, food.ID, user.ID)
	r2, _ := s.CreateRecord(t.Context(), "2026-05-05", -12000, currency.ID, "lunch", user.ID)
	s.AddRecordTag(t.Context(), r2.ID, food.ID, user.ID)

	// Save query and create target
	sq, _ := s.CreateSavedQueryWithTags(t.Context(), "food budget", "2026-05-01", "2026-05-31", nil, "", user.ID, []int64{food.ID})
	target, _ := s.CreateTarget(t.Context(), "food budget", sq.ID, -50000, user.ID)

	// Check target: actual should be -$320 (under budget)
	result, _ := s.GetTargetWithActual(t.Context(), target.ID)
	if !result.HasData {
		t.Fatal("expected HasData=true")
	}
	if *result.ActualAmount != -32000 {
		t.Errorf("expected actual=-32000, got %d", *result.ActualAmount)
	}
	// Gap: -32000 - (-50000) = +18000 (under budget by $180)

	// Add more food expense: -$250 (over budget now)
	r3, _ := s.CreateRecord(t.Context(), "2026-05-15", -25000, currency.ID, "dinner", user.ID)
	s.AddRecordTag(t.Context(), r3.ID, food.ID, user.ID)

	// Check again: actual should be -$570
	result2, _ := s.GetTargetWithActual(t.Context(), target.ID)
	if !result2.HasData {
		t.Fatal("expected HasData=true after adding expense")
	}
	if *result2.ActualAmount != -57000 {
		t.Errorf("expected actual=-57000, got %d", *result2.ActualAmount)
	}
	// Gap: -57000 - (-50000) = -7000 (over budget by $70)

	_ = currency
}

// Journey 5: "Is my spending growing?" → monthly trend
func TestJourney_MonthlyTrend(t *testing.T) {
	s := setupTestService(t)

	user, _ := s.CreateUser(t.Context(), "testuser")
	currency, _ := s.CreateCurrency(t.Context(), "USD")

	// Jan: net -$200
	s.CreateRecord(t.Context(), "2026-01-15", -20000, currency.ID, "expense", user.ID)

	// Feb: net +$300
	s.CreateRecord(t.Context(), "2026-02-10", 50000, currency.ID, "income", user.ID)
	s.CreateRecord(t.Context(), "2026-02-20", -20000, currency.ID, "expense", user.ID)

	// Mar: net +$900
	s.CreateRecord(t.Context(), "2026-03-01", 50000, currency.ID, "income", user.ID)
	s.CreateRecord(t.Context(), "2026-03-15", 40000, currency.ID, "income", user.ID)

	result, err := s.AggregateByPeriod(t.Context(), "2026-01-01", "2026-12-31", nil, nil, "", service.PeriodByMonth)
	if err != nil {
		t.Fatalf("AggregateByPeriod() error: %v", err)
	}

	if len(result.Periods) != 3 {
		t.Fatalf("expected 3 monthly periods, got %d", len(result.Periods))
	}

	if result.Periods[0].Amount != -20000 {
		t.Errorf("expected Jan net=-20000, got %d", result.Periods[0].Amount)
	}
	if result.Periods[1].Amount != 30000 {
		t.Errorf("expected Feb net=30000, got %d", result.Periods[1].Amount)
	}
	if result.Periods[2].Amount != 90000 {
		t.Errorf("expected Mar net=90000, got %d", result.Periods[2].Amount)
	}

	// Trend: growing! Jan < Feb < Mar
	_ = user
}

// Journey 15+19+20: Progressive disclosure — target with no data, then data appears
func TestJourney_ProgressiveDisclosure(t *testing.T) {
	s := setupTestService(t)

	user, _ := s.CreateUser(t.Context(), "testuser")

	investment, _ := s.CreateTag(t.Context(), "investment", "", "", user.ID)

	// Create target for investment category with no data
	sq, _ := s.CreateSavedQueryWithTags(t.Context(), "investment budget", "2026-05-01", "2026-05-31", nil, "", user.ID, []int64{investment.ID})
	target, _ := s.CreateTarget(t.Context(), "investment budget", sq.ID, -50000, user.ID)

	// No data: target shows ???
	result, _ := s.GetTargetWithActual(t.Context(), target.ID)
	if result.HasData {
		t.Error("expected HasData=false for empty category")
	}
	if result.ActualAmount != nil {
		t.Errorf("expected ActualAmount=nil, got %d", *result.ActualAmount)
	}

	// Add first investment record
	currency, _ := s.CreateCurrency(t.Context(), "USD")
	r, _ := s.CreateRecord(t.Context(), "2026-05-10", -20000, currency.ID, "ETF", user.ID)
	s.AddRecordTag(t.Context(), r.ID, investment.ID, user.ID)

	// Now data appears
	result2, _ := s.GetTargetWithActual(t.Context(), target.ID)
	if !result2.HasData {
		t.Error("expected HasData=true after adding record")
	}
	if result2.ActualAmount == nil {
		t.Fatal("expected ActualAmount non-nil")
	}
	if *result2.ActualAmount != -20000 {
		t.Errorf("expected actual=-20000, got %d", *result2.ActualAmount)
	}
}

// Journey 10: Multi-tag record — both tags get the amount, total not double-counted
func TestJourney_MultiTag(t *testing.T) {
	s := setupTestService(t)

	user, _ := s.CreateUser(t.Context(), "testuser")
	currency, _ := s.CreateCurrency(t.Context(), "USD")
	food, _ := s.CreateTag(t.Context(), "food", "", "", user.ID)
	dining, _ := s.CreateTag(t.Context(), "dining", "", "", user.ID)

	r, _ := s.CreateRecord(t.Context(), "2026-05-01", -5000, currency.ID, "", user.ID)
	s.AddRecordTag(t.Context(), r.ID, food.ID, user.ID)
	s.AddRecordTag(t.Context(), r.ID, dining.ID, user.ID)

	result, err := s.Aggregate(t.Context(), "2026-05-01", "2026-05-31", nil, nil, "")
	if err != nil {
		t.Fatalf("Aggregate() error: %v", err)
	}

	// Total is -$50 (not -$100, no double-counting)
	if result.TotalAmount != -5000 {
		t.Errorf("expected TotalAmount=-5000, got %d", result.TotalAmount)
	}
	if result.RecordCount != 1 {
		t.Errorf("expected RecordCount=1, got %d", result.RecordCount)
	}

	tagMap := make(map[string]int64)
	for _, ts := range result.ByTag {
		tagMap[ts.TagName] = ts.Amount
	}
	// Both tags show the full amount
	if tagMap["food"] != -5000 {
		t.Errorf("expected food=-5000, got %d", tagMap["food"])
	}
	if tagMap["dining"] != -5000 {
		t.Errorf("expected dining=-5000, got %d", tagMap["dining"])
	}
}

// Journey 24: Mixed currencies — V1 just sums cents regardless
func TestJourney_MixedCurrencies(t *testing.T) {
	s := setupTestService(t)

	user, _ := s.CreateUser(t.Context(), "testuser")
	usd, _ := s.CreateCurrency(t.Context(), "USD")
	php, _ := s.CreateCurrency(t.Context(), "PHP")

	// $50 USD + 2000 PHP — just sums cents (meaningless but V1 behavior)
	s.CreateRecord(t.Context(), "2026-05-01", 5000, usd.ID, "USD income", user.ID)
	s.CreateRecord(t.Context(), "2026-05-02", 200000, php.ID, "PHP income", user.ID)

	result, err := s.Aggregate(t.Context(), "2026-05-01", "2026-05-31", nil, nil, "")
	if err != nil {
		t.Fatalf("Aggregate() error: %v", err)
	}

	// V1: just sums without currency awareness
	if !result.HasData {
		t.Error("expected HasData=true")
	}
	if result.TotalAmount != 205000 {
		t.Errorf("expected TotalAmount=205000 (V1: raw cents sum), got %d", result.TotalAmount)
	}
}

// Journey 22: From query results create target
func TestJourney_TargetFromQuery(t *testing.T) {
	s := setupTestService(t)

	user, _ := s.CreateUser(t.Context(), "testuser")
	currency, _ := s.CreateCurrency(t.Context(), "USD")
	food, _ := s.CreateTag(t.Context(), "food", "", "", user.ID)

	// Create records
	r1, _ := s.CreateRecord(t.Context(), "2026-05-01", -20000, currency.ID, "groceries", user.ID)
	s.AddRecordTag(t.Context(), r1.ID, food.ID, user.ID)

	// Step 1: Run aggregation manually (simulates query)
	agg, err := s.Aggregate(t.Context(), "2026-05-01", "2026-05-31", nil, []int64{food.ID}, "")
	if err != nil {
		t.Fatalf("Aggregate() error: %v", err)
	}
	if !agg.HasData {
		t.Fatal("expected HasData=true")
	}
	if agg.TotalAmount != -20000 {
		t.Errorf("expected aggregation total=-20000, got %d", agg.TotalAmount)
	}

	// Step 2: Save query
	sq, _ := s.CreateSavedQueryWithTags(t.Context(), "food budget", "2026-05-01", "2026-05-31", nil, "", user.ID, []int64{food.ID})

	// Step 3: Create target from saved query
	target, _ := s.CreateTarget(t.Context(), "food budget", sq.ID, -50000, user.ID)

	// Step 4: Verify target matches aggregation
	result, _ := s.GetTargetWithActual(t.Context(), target.ID)
	if !result.HasData {
		t.Fatal("expected HasData=true")
	}
	if *result.ActualAmount != agg.TotalAmount {
		t.Errorf("target actual (%d) should match aggregation (%d)", *result.ActualAmount, agg.TotalAmount)
	}
}

// Journey 23: Snapshot convention — #snapshot tagged records are summed normally
func TestJourney_SnapshotConvention(t *testing.T) {
	s := setupTestService(t)

	user, _ := s.CreateUser(t.Context(), "testuser")
	currency, _ := s.CreateCurrency(t.Context(), "USD")
	checking, _ := s.CreateTag(t.Context(), "checking", "", "", user.ID)
	snapshot, _ := s.CreateTag(t.Context(), "snapshot", "", "", user.ID)

	// Snapshot: checking balance $5000 on May 1
	r1, _ := s.CreateRecord(t.Context(), "2026-05-01", 500000, currency.ID, "checking balance", user.ID)
	s.AddRecordTag(t.Context(), r1.ID, checking.ID, user.ID)
	s.AddRecordTag(t.Context(), r1.ID, snapshot.ID, user.ID)

	// Transactions in May
	r2, _ := s.CreateRecord(t.Context(), "2026-05-10", -50000, currency.ID, "rent", user.ID)
	s.AddRecordTag(t.Context(), r2.ID, checking.ID, user.ID)
	r3, _ := s.CreateRecord(t.Context(), "2026-05-15", -30000, currency.ID, "food", user.ID)
	s.AddRecordTag(t.Context(), r3.ID, checking.ID, user.ID)

	// Query #checking for May: includes snapshot + transactions
	result, err := s.Aggregate(t.Context(), "2026-05-01", "2026-05-31", nil, []int64{checking.ID}, "")
	if err != nil {
		t.Fatalf("Aggregate() error: %v", err)
	}

	// V1: snapshot is just summed like any other record
	// Total = 500000 - 50000 - 30000 = 420000
	if result.TotalAmount != 420000 {
		t.Errorf("expected TotalAmount=420000 (including snapshot), got %d", result.TotalAmount)
	}

	// Future: detect #snapshot and display differently (start + activity + implied)
}