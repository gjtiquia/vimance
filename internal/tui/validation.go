package tui

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var amountRegex = regexp.MustCompile(`^\d+(\.\d{1,2})?$`)

func validateAmount(amount string) error {
	amount = strings.TrimSpace(amount)
	if amount == "" {
		return fmt.Errorf("amount is required")
	}

	if !amountRegex.MatchString(amount) {
		return fmt.Errorf("invalid amount format (use format like 12.50)")
	}

	return nil
}

func parseAmountToCents(amount string) (int64, error) {
	if err := validateAmount(amount); err != nil {
		return 0, err
	}

	parts := strings.Split(amount, ".")
	cents := int64(0)

	if len(parts) == 1 {
		dollars, _ := strconv.ParseInt(parts[0], 10, 64)
		cents = dollars * 100
	} else {
		dollars, _ := strconv.ParseInt(parts[0], 10, 64)
		cents = dollars * 100

		decimalPart := parts[1]
		if len(decimalPart) == 1 {
			decimalCents, _ := strconv.ParseInt(decimalPart, 10, 64)
			cents += decimalCents * 10
		} else {
			decimalCents, _ := strconv.ParseInt(decimalPart, 10, 64)
			cents += decimalCents
		}
	}

	return cents, nil
}

func validateDate(year, month, day string) error {
	year = strings.TrimSpace(year)
	month = strings.TrimSpace(month)
	day = strings.TrimSpace(day)

	if year == "" || month == "" || day == "" {
		return fmt.Errorf("date is required")
	}

	dateStr := fmt.Sprintf("%s-%s-%s", year, month, day)

	_, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return fmt.Errorf("invalid date format")
	}

	return nil
}

func formatDate(year, month, day string) string {
	return fmt.Sprintf("%s-%s-%s", year, month, day)
}

type ValidationError struct {
	Field   string
	Message string
}

type ValidationErrors []ValidationError

func (e ValidationErrors) HasErrors() bool {
	return len(e) > 0
}

func (e ValidationErrors) Get(field string) string {
	for _, err := range e {
		if err.Field == field {
			return err.Message
		}
	}
	return ""
}

func (m RecordModel) Validate() ValidationErrors {
	var errors ValidationErrors

	if err := validateDate(m.DateYearInput.Value(), m.DateMonthInput.Value(), m.DateDayInput.Value()); err != nil {
		errors = append(errors, ValidationError{Field: "date", Message: err.Error()})
	}

	if m.CurrencyInput.Selected == nil {
		errors = append(errors, ValidationError{Field: "currency", Message: "currency is required"})
	}

	if err := validateAmount(m.AmountInput.Value()); err != nil {
		errors = append(errors, ValidationError{Field: "amount", Message: err.Error()})
	}

	return errors
}

func (m RecordModel) GetWarnings() ValidationErrors {
	var warnings ValidationErrors

	if len(m.TagsInput.SelectedTags) == 0 {
		warnings = append(warnings, ValidationError{Field: "tags", Message: "no tags selected"})
	}

	return warnings
}
