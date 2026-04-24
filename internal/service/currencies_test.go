package service_test

import (
	"testing"
)

func TestCurrencyCRUD(t *testing.T) {
	s := setupTestService(t)

	currency, err := s.CreateCurrency(t.Context(), "USD")
	if err != nil {
		t.Fatalf("failed to create currency: %v", err)
	}

	if currency.ID == 0 {
		t.Error("expected currency ID to be set")
	}
	if currency.Code != "USD" {
		t.Errorf("expected code 'USD', got '%s'", currency.Code)
	}

	fetched, err := s.GetCurrency(t.Context(), currency.ID)
	if err != nil {
		t.Fatalf("failed to get currency: %v", err)
	}
	if fetched.Code != currency.Code {
		t.Errorf("expected code '%s', got '%s'", currency.Code, fetched.Code)
	}

	byCode, err := s.GetCurrencyByCode(t.Context(), "USD")
	if err != nil {
		t.Fatalf("failed to get currency by code: %v", err)
	}
	if byCode.ID != currency.ID {
		t.Errorf("expected ID %d, got %d", currency.ID, byCode.ID)
	}

	currencies, err := s.ListCurrencies(t.Context())
	if err != nil {
		t.Fatalf("failed to list currencies: %v", err)
	}
	if len(currencies) != 1 {
		t.Errorf("expected 1 currency, got %d", len(currencies))
	}

	updated, err := s.UpdateCurrency(t.Context(), currency.ID, "EUR")
	if err != nil {
		t.Fatalf("failed to update currency: %v", err)
	}
	if updated.Code != "EUR" {
		t.Errorf("expected code 'EUR', got '%s'", updated.Code)
	}

	err = s.DeleteCurrency(t.Context(), currency.ID)
	if err != nil {
		t.Fatalf("failed to delete currency: %v", err)
	}

	_, err = s.GetCurrency(t.Context(), currency.ID)
	if err == nil {
		t.Error("expected error when getting deleted currency")
	}
}

func TestGetOrCreateCurrency(t *testing.T) {
	s := setupTestService(t)

	currency1, created1, err := s.GetOrCreateCurrency(t.Context(), "USD")
	if err != nil {
		t.Fatalf("failed to get or create currency: %v", err)
	}
	if !created1 {
		t.Error("expected currency to be created")
	}
	if currency1.Code != "USD" {
		t.Errorf("expected code 'USD', got '%s'", currency1.Code)
	}

	currency2, created2, err := s.GetOrCreateCurrency(t.Context(), "USD")
	if err != nil {
		t.Fatalf("failed to get or create currency: %v", err)
	}
	if created2 {
		t.Error("expected currency to already exist")
	}
	if currency2.ID != currency1.ID {
		t.Errorf("expected same currency ID, got %d and %d", currency1.ID, currency2.ID)
	}

	currency3, created3, err := s.GetOrCreateCurrency(t.Context(), "EUR")
	if err != nil {
		t.Fatalf("failed to get or create currency: %v", err)
	}
	if !created3 {
		t.Error("expected currency to be created")
	}
	if currency3.Code != "EUR" {
		t.Errorf("expected code 'EUR', got '%s'", currency3.Code)
	}
}
