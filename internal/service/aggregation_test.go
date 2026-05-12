package service_test

import (
	"testing"

	service "github.com/gjtiquia/vimance/internal/service"
)

func TestAggregate_EmptyResult(t *testing.T) {
	s := setupTestService(t)

	result, err := s.Aggregate(t.Context(), "2026-01-01", "2026-01-31", nil, nil, "")
	if err != nil {
		t.Fatalf("Aggregate() error: %v", err)
	}

	if result.HasData {
		t.Error("expected HasData=false for empty result")
	}
	if result.TotalAmount != 0 {
		t.Errorf("expected TotalAmount=0, got %d", result.TotalAmount)
	}
	if result.IncomeSum != 0 {
		t.Errorf("expected IncomeSum=0, got %d", result.IncomeSum)
	}
	if result.ExpenseSum != 0 {
		t.Errorf("expected ExpenseSum=0, got %d", result.ExpenseSum)
	}
	if result.RecordCount != 0 {
		t.Errorf("expected RecordCount=0, got %d", result.RecordCount)
	}
	if len(result.ByTag) != 0 {
		t.Errorf("expected empty ByTag, got %d entries", len(result.ByTag))
	}
}

func TestAggregate_SingleRecord(t *testing.T) {
	s := setupTestService(t)

	user, _ := s.CreateUser(t.Context(), "testuser")
	currency, _ := s.CreateCurrency(t.Context(), "USD")

	s.CreateRecord(t.Context(), "2026-01-15", -5000, currency.ID, "lunch", user.ID)

	result, err := s.Aggregate(t.Context(), "2026-01-01", "2026-01-31", nil, nil, "")
	if err != nil {
		t.Fatalf("Aggregate() error: %v", err)
	}

	if !result.HasData {
		t.Error("expected HasData=true")
	}
	if result.TotalAmount != -5000 {
		t.Errorf("expected TotalAmount=-5000, got %d", result.TotalAmount)
	}
	if result.IncomeSum != 0 {
		t.Errorf("expected IncomeSum=0, got %d", result.IncomeSum)
	}
	if result.ExpenseSum != -5000 {
		t.Errorf("expected ExpenseSum=-5000, got %d", result.ExpenseSum)
	}
	if result.RecordCount != 1 {
		t.Errorf("expected RecordCount=1, got %d", result.RecordCount)
	}
}

func TestAggregate_Totals_IncomeExpense(t *testing.T) {
	s := setupTestService(t)

	user, _ := s.CreateUser(t.Context(), "testuser")
	currency, _ := s.CreateCurrency(t.Context(), "USD")

	// 3 expenses, 2 income
	s.CreateRecord(t.Context(), "2026-05-01", -10000, currency.ID, "rent", user.ID)
	s.CreateRecord(t.Context(), "2026-05-02", -5000, currency.ID, "food", user.ID)
	s.CreateRecord(t.Context(), "2026-05-03", -2000, currency.ID, "transport", user.ID)
	s.CreateRecord(t.Context(), "2026-05-10", 350000, currency.ID, "salary", user.ID)
	s.CreateRecord(t.Context(), "2026-05-15", 50000, currency.ID, "freelance", user.ID)

	result, err := s.Aggregate(t.Context(), "2026-05-01", "2026-05-31", nil, nil, "")
	if err != nil {
		t.Fatalf("Aggregate() error: %v", err)
	}

	if !result.HasData {
		t.Fatal("expected HasData=true")
	}
	if result.RecordCount != 5 {
		t.Errorf("expected RecordCount=5, got %d", result.RecordCount)
	}
	// Total = -10000 - 5000 - 2000 + 350000 + 50000 = 383000
	if result.TotalAmount != 383000 {
		t.Errorf("expected TotalAmount=383000, got %d", result.TotalAmount)
	}
	if result.IncomeSum != 400000 {
		t.Errorf("expected IncomeSum=400000, got %d", result.IncomeSum)
	}
	if result.ExpenseSum != -17000 {
		t.Errorf("expected ExpenseSum=-17000, got %d", result.ExpenseSum)
	}
}

func TestAggregate_ByTag_Basic(t *testing.T) {
	s := setupTestService(t)

	user, _ := s.CreateUser(t.Context(), "testuser")
	currency, _ := s.CreateCurrency(t.Context(), "USD")
	food, _ := s.CreateTag(t.Context(), "food", "", "", user.ID)
	rent, _ := s.CreateTag(t.Context(), "rent", "", "", user.ID)

	r1, _ := s.CreateRecord(t.Context(), "2026-05-01", -10000, currency.ID, "", user.ID)
	s.AddRecordTag(t.Context(), r1.ID, rent.ID, user.ID)

	r2, _ := s.CreateRecord(t.Context(), "2026-05-02", -5000, currency.ID, "", user.ID)
	s.AddRecordTag(t.Context(), r2.ID, food.ID, user.ID)

	r3, _ := s.CreateRecord(t.Context(), "2026-05-03", -3000, currency.ID, "", user.ID)
	s.AddRecordTag(t.Context(), r3.ID, food.ID, user.ID)

	result, err := s.Aggregate(t.Context(), "2026-05-01", "2026-05-31", nil, nil, "")
	if err != nil {
		t.Fatalf("Aggregate() error: %v", err)
	}

	tagMap := make(map[string]service.TagSum)
	for _, ts := range result.ByTag {
		tagMap[ts.TagName] = ts
	}

	if foodSum, ok := tagMap["food"]; !ok {
		t.Error("expected food tag in ByTag")
	} else {
		if foodSum.Amount != -8000 {
			t.Errorf("expected food amount=-8000, got %d", foodSum.Amount)
		}
		if foodSum.Count != 2 {
			t.Errorf("expected food count=2, got %d", foodSum.Count)
		}
	}

	if rentSum, ok := tagMap["rent"]; !ok {
		t.Error("expected rent tag in ByTag")
	} else {
		if rentSum.Amount != -10000 {
			t.Errorf("expected rent amount=-10000, got %d", rentSum.Amount)
		}
		if rentSum.Count != 1 {
			t.Errorf("expected rent count=1, got %d", rentSum.Count)
		}
	}
}

func TestAggregate_ByTag_SameTag(t *testing.T) {
	s := setupTestService(t)

	user, _ := s.CreateUser(t.Context(), "testuser")
	currency, _ := s.CreateCurrency(t.Context(), "USD")
	food, _ := s.CreateTag(t.Context(), "food", "", "", user.ID)

	for i := 0; i < 10; i++ {
		r, _ := s.CreateRecord(t.Context(), "2026-05-01", -1000, currency.ID, "", user.ID)
		s.AddRecordTag(t.Context(), r.ID, food.ID, user.ID)
	}

	result, err := s.Aggregate(t.Context(), "2026-05-01", "2026-05-31", nil, nil, "")
	if err != nil {
		t.Fatalf("Aggregate() error: %v", err)
	}

	if len(result.ByTag) != 1 {
		t.Fatalf("expected 1 tag in ByTag, got %d", len(result.ByTag))
	}
	if result.ByTag[0].TagName != "food" {
		t.Errorf("expected tag name 'food', got %s", result.ByTag[0].TagName)
	}
	if result.ByTag[0].Amount != -10000 {
		t.Errorf("expected food amount=-10000, got %d", result.ByTag[0].Amount)
	}
	if result.ByTag[0].Count != 10 {
		t.Errorf("expected food count=10, got %d", result.ByTag[0].Count)
	}
}

func TestAggregate_ByTag_MultiTag(t *testing.T) {
	s := setupTestService(t)

	user, _ := s.CreateUser(t.Context(), "testuser")
	currency, _ := s.CreateCurrency(t.Context(), "USD")
	food, _ := s.CreateTag(t.Context(), "food", "", "", user.ID)
	dining, _ := s.CreateTag(t.Context(), "dining", "", "", user.ID)

	// single record with 2 tags, -$50
	r, _ := s.CreateRecord(t.Context(), "2026-05-01", -5000, currency.ID, "", user.ID)
	s.AddRecordTag(t.Context(), r.ID, food.ID, user.ID)
	s.AddRecordTag(t.Context(), r.ID, dining.ID, user.ID)

	result, err := s.Aggregate(t.Context(), "2026-05-01", "2026-05-31", nil, nil, "")
	if err != nil {
		t.Fatalf("Aggregate() error: %v", err)
	}

	// Total should NOT double-count: -5000, not -10000
	if result.TotalAmount != -5000 {
		t.Errorf("expected TotalAmount=-5000 (no double-counting), got %d", result.TotalAmount)
	}
	if result.RecordCount != 1 {
		t.Errorf("expected RecordCount=1, got %d", result.RecordCount)
	}

	tagMap := make(map[string]service.TagSum)
	for _, ts := range result.ByTag {
		tagMap[ts.TagName] = ts
	}

	// Each tag shows the full amount of the record
	if foodSum, ok := tagMap["food"]; !ok {
		t.Error("expected food tag in ByTag")
	} else if foodSum.Amount != -5000 {
		t.Errorf("expected food amount=-5000, got %d", foodSum.Amount)
	}

	if diningSum, ok := tagMap["dining"]; !ok {
		t.Error("expected dining tag in ByTag")
	} else if diningSum.Amount != -5000 {
		t.Errorf("expected dining amount=-5000, got %d", diningSum.Amount)
	}
}

func TestAggregate_Untagged(t *testing.T) {
	s := setupTestService(t)

	user, _ := s.CreateUser(t.Context(), "testuser")
	currency, _ := s.CreateCurrency(t.Context(), "USD")

	// record with no tags
	s.CreateRecord(t.Context(), "2026-05-01", -5000, currency.ID, "", user.ID)

	result, err := s.Aggregate(t.Context(), "2026-05-01", "2026-05-31", nil, nil, "")
	if err != nil {
		t.Fatalf("Aggregate() error: %v", err)
	}

	if result.TotalAmount != -5000 {
		t.Errorf("expected TotalAmount=-5000, got %d", result.TotalAmount)
	}

	// untagged record should appear under "(untagged)"
	found := false
	for _, ts := range result.ByTag {
		if ts.TagName == "(untagged)" {
			found = true
			if ts.Amount != -5000 {
				t.Errorf("expected (untagged) amount=-5000, got %d", ts.Amount)
			}
			if ts.Count != 1 {
				t.Errorf("expected (untagged) count=1, got %d", ts.Count)
			}
		}
	}
	if !found {
		t.Error("expected (untagged) entry in ByTag")
	}
}

func TestAggregate_ByTag_MixedSigns(t *testing.T) {
	s := setupTestService(t)

	user, _ := s.CreateUser(t.Context(), "testuser")
	currency, _ := s.CreateCurrency(t.Context(), "USD")
	food, _ := s.CreateTag(t.Context(), "food", "", "", user.ID)

	// refund +$50 (5000 cents)
	r1, _ := s.CreateRecord(t.Context(), "2026-05-01", 5000, currency.ID, "refund", user.ID)
	s.AddRecordTag(t.Context(), r1.ID, food.ID, user.ID)

	// groceries -$200 (20000 cents)
	r2, _ := s.CreateRecord(t.Context(), "2026-05-02", -20000, currency.ID, "groceries", user.ID)
	s.AddRecordTag(t.Context(), r2.ID, food.ID, user.ID)

	result, err := s.Aggregate(t.Context(), "2026-05-01", "2026-05-31", nil, nil, "")
	if err != nil {
		t.Fatalf("Aggregate() error: %v", err)
	}

	// food tag sum = 5000 + (-20000) = -15000
	foodSum := result.ByTag[0]
	if foodSum.TagName != "food" {
		t.Errorf("expected food tag, got %s", foodSum.TagName)
	}
	if foodSum.Amount != -15000 {
		t.Errorf("expected food amount=-15000, got %d", foodSum.Amount)
	}
	if foodSum.Count != 2 {
		t.Errorf("expected food count=2, got %d", foodSum.Count)
	}
}

func TestAggregate_WithFilters(t *testing.T) {
	s := setupTestService(t)

	user, _ := s.CreateUser(t.Context(), "testuser")
	usd, _ := s.CreateCurrency(t.Context(), "USD")
	php, _ := s.CreateCurrency(t.Context(), "PHP")
	food, _ := s.CreateTag(t.Context(), "food", "", "", user.ID)

	// USD food in May
	r1, _ := s.CreateRecord(t.Context(), "2026-05-01", -5000, usd.ID, "lunch", user.ID)
	s.AddRecordTag(t.Context(), r1.ID, food.ID, user.ID)

	// PHP food in May
	r2, _ := s.CreateRecord(t.Context(), "2026-05-02", -3000, php.ID, "merienda", user.ID)
	s.AddRecordTag(t.Context(), r2.ID, food.ID, user.ID)

	// USD food in April
	r3, _ := s.CreateRecord(t.Context(), "2026-04-15", -7000, usd.ID, "dinner", user.ID)
	s.AddRecordTag(t.Context(), r3.ID, food.ID, user.ID)

	// Filter: USD only, May only, tag: food
	usdID := usd.ID
	result, err := s.Aggregate(t.Context(), "2026-05-01", "2026-05-31", &usdID, []int64{food.ID}, "")
	if err != nil {
		t.Fatalf("Aggregate() error: %v", err)
	}

	if result.RecordCount != 1 {
		t.Errorf("expected 1 record with USD+food+May filter, got %d", result.RecordCount)
	}
	if result.TotalAmount != -5000 {
		t.Errorf("expected TotalAmount=-5000, got %d", result.TotalAmount)
	}

	// Filter: fuzzy search "lunch"
	result2, err := s.Aggregate(t.Context(), "2026-01-01", "2026-12-31", nil, nil, "lunch")
	if err != nil {
		t.Fatalf("Aggregate() error: %v", err)
	}

	if result2.RecordCount != 1 {
		t.Errorf("expected 1 record with fuzzy 'lunch', got %d", result2.RecordCount)
	}
}

func TestAggregateByPeriod_Monthly(t *testing.T) {
	s := setupTestService(t)

	user, _ := s.CreateUser(t.Context(), "testuser")
	currency, _ := s.CreateCurrency(t.Context(), "USD")

	// Jan: -$200 total
	s.CreateRecord(t.Context(), "2026-01-15", -20000, currency.ID, "", user.ID)

	// Feb: -$600 total
	s.CreateRecord(t.Context(), "2026-02-10", -30000, currency.ID, "", user.ID)
	s.CreateRecord(t.Context(), "2026-02-20", -30000, currency.ID, "", user.ID)

	// Mar: +$900 total
	s.CreateRecord(t.Context(), "2026-03-01", 50000, currency.ID, "", user.ID)
	s.CreateRecord(t.Context(), "2026-03-15", 40000, currency.ID, "", user.ID)

	result, err := s.AggregateByPeriod(t.Context(), "2026-01-01", "2026-12-31", nil, nil, "", service.PeriodByMonth)
	if err != nil {
		t.Fatalf("AggregateByPeriod() error: %v", err)
	}

	if !result.HasData {
		t.Fatal("expected HasData=true")
	}
	if len(result.Periods) != 3 {
		t.Fatalf("expected 3 periods, got %d", len(result.Periods))
	}

	// Periods sorted chronologically
	if result.Periods[0].Period != "2026-01" {
		t.Errorf("expected period 0 = 2026-01, got %s", result.Periods[0].Period)
	}
	if result.Periods[0].Amount != -20000 {
		t.Errorf("expected 2026-01 amount=-20000, got %d", result.Periods[0].Amount)
	}
	if result.Periods[0].Count != 1 {
		t.Errorf("expected 2026-01 count=1, got %d", result.Periods[0].Count)
	}

	if result.Periods[1].Period != "2026-02" {
		t.Errorf("expected period 1 = 2026-02, got %s", result.Periods[1].Period)
	}
	if result.Periods[1].Amount != -60000 {
		t.Errorf("expected 2026-02 amount=-60000, got %d", result.Periods[1].Amount)
	}

	if result.Periods[2].Period != "2026-03" {
		t.Errorf("expected period 2 = 2026-03, got %s", result.Periods[2].Period)
	}
	if result.Periods[2].Amount != 90000 {
		t.Errorf("expected 2026-03 amount=90000, got %d", result.Periods[2].Amount)
	}
}

func TestAggregateByPeriod_Weekly(t *testing.T) {
	s := setupTestService(t)

	user, _ := s.CreateUser(t.Context(), "testuser")
	currency, _ := s.CreateCurrency(t.Context(), "USD")

	// Jan 6, 2026 = ISO week 2026-W02, Jan 8 = also W02
	s.CreateRecord(t.Context(), "2026-01-06", -10000, currency.ID, "", user.ID)
	s.CreateRecord(t.Context(), "2026-01-08", -5000, currency.ID, "", user.ID)

	// Jan 13, 2026 = ISO week 2026-W03
	s.CreateRecord(t.Context(), "2026-01-13", -30000, currency.ID, "", user.ID)

	result, err := s.AggregateByPeriod(t.Context(), "2026-01-01", "2026-12-31", nil, nil, "", service.PeriodByWeek)
	if err != nil {
		t.Fatalf("AggregateByPeriod() error: %v", err)
	}

	if !result.HasData {
		t.Fatal("expected HasData=true")
	}

	if len(result.Periods) != 2 {
		t.Fatalf("expected 2 weekly periods, got %d", len(result.Periods))
	}

	// First period: W02 with two records totaling -15000
	if result.Periods[0].Period != "2026-W02" {
		t.Errorf("expected period 0 = 2026-W02, got %s", result.Periods[0].Period)
	}
	if result.Periods[0].Amount != -15000 {
		t.Errorf("expected W02 amount=-15000, got %d", result.Periods[0].Amount)
	}
	if result.Periods[0].Count != 2 {
		t.Errorf("expected W02 count=2, got %d", result.Periods[0].Count)
	}

	// Second period: W03 with one record
	if result.Periods[1].Period != "2026-W03" {
		t.Errorf("expected period 1 = 2026-W03, got %s", result.Periods[1].Period)
	}
	if result.Periods[1].Amount != -30000 {
		t.Errorf("expected W03 amount=-30000, got %d", result.Periods[1].Amount)
	}
	if result.Periods[1].Count != 1 {
		t.Errorf("expected W03 count=1, got %d", result.Periods[1].Count)
	}
}

func TestAggregateByPeriod_Yearly(t *testing.T) {
	s := setupTestService(t)

	user, _ := s.CreateUser(t.Context(), "testuser")
	currency, _ := s.CreateCurrency(t.Context(), "USD")

	s.CreateRecord(t.Context(), "2025-06-15", -50000, currency.ID, "", user.ID)
	s.CreateRecord(t.Context(), "2026-01-15", -20000, currency.ID, "", user.ID)
	s.CreateRecord(t.Context(), "2026-05-15", 80000, currency.ID, "", user.ID)

	result, err := s.AggregateByPeriod(t.Context(), "2025-01-01", "2026-12-31", nil, nil, "", service.PeriodByYear)
	if err != nil {
		t.Fatalf("AggregateByPeriod() error: %v", err)
	}

	if len(result.Periods) != 2 {
		t.Fatalf("expected 2 yearly periods, got %d", len(result.Periods))
	}

	if result.Periods[0].Period != "2025" {
		t.Errorf("expected period 0 = 2025, got %s", result.Periods[0].Period)
	}
	if result.Periods[0].Amount != -50000 {
		t.Errorf("expected 2025 amount=-50000, got %d", result.Periods[0].Amount)
	}

	if result.Periods[1].Period != "2026" {
		t.Errorf("expected period 1 = 2026, got %s", result.Periods[1].Period)
	}
	if result.Periods[1].Amount != 60000 {
		t.Errorf("expected 2026 amount=60000, got %d", result.Periods[1].Amount)
	}
}

func TestAggregateByPeriod_Daily(t *testing.T) {
	s := setupTestService(t)

	user, _ := s.CreateUser(t.Context(), "testuser")
	currency, _ := s.CreateCurrency(t.Context(), "USD")

	// Two records on same day, one on different day
	s.CreateRecord(t.Context(), "2026-05-01", -5000, currency.ID, "lunch", user.ID)
	s.CreateRecord(t.Context(), "2026-05-01", -3000, currency.ID, "coffee", user.ID)
	s.CreateRecord(t.Context(), "2026-05-03", -10000, currency.ID, "rent", user.ID)

	result, err := s.AggregateByPeriod(t.Context(), "2026-05-01", "2026-05-31", nil, nil, "", service.PeriodByDay)
	if err != nil {
		t.Fatalf("AggregateByPeriod() error: %v", err)
	}

	if len(result.Periods) != 2 {
		t.Fatalf("expected 2 daily periods, got %d", len(result.Periods))
	}

	if result.Periods[0].Period != "2026-05-01" {
		t.Errorf("expected period 0 = 2026-05-01, got %s", result.Periods[0].Period)
	}
	if result.Periods[0].Amount != -8000 {
		t.Errorf("expected 2026-05-01 amount=-8000, got %d", result.Periods[0].Amount)
	}
	if result.Periods[0].Count != 2 {
		t.Errorf("expected 2026-05-01 count=2, got %d", result.Periods[0].Count)
	}

	if result.Periods[1].Period != "2026-05-03" {
		t.Errorf("expected period 1 = 2026-05-03, got %s", result.Periods[1].Period)
	}
	if result.Periods[1].Amount != -10000 {
		t.Errorf("expected 2026-05-03 amount=-10000, got %d", result.Periods[1].Amount)
	}
}

func TestAggregateByPeriod_Empty(t *testing.T) {
	s := setupTestService(t)

	result, err := s.AggregateByPeriod(t.Context(), "2026-01-01", "2026-01-31", nil, nil, "", service.PeriodByMonth)
	if err != nil {
		t.Fatalf("AggregateByPeriod() error: %v", err)
	}

	if result.HasData {
		t.Error("expected HasData=false for empty result")
	}
	if result.TotalAmount != 0 {
		t.Errorf("expected TotalAmount=0, got %d", result.TotalAmount)
	}
	if len(result.Periods) != 0 {
		t.Errorf("expected empty Periods, got %d entries", len(result.Periods))
	}
}