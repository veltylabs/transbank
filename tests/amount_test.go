package tests

import (
	"testing"
	"github.com/veltylabs/payment_manager"
	"github.com/veltylabs/transbank"
)

func TestAmount_CLP(t *testing.T) {
	enc, err := transbank.ExportEncodeAmount(1500, paymentmanager.CurrencyCLP)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if enc != 1500.0 {
		t.Fatalf("expected 1500.0, got %f", enc)
	}

	dec, err := transbank.ExportDecodeAmount(1500.0, paymentmanager.CurrencyCLP)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dec != 1500 {
		t.Fatalf("expected 1500, got %d", dec)
	}
}

func TestAmount_USD(t *testing.T) {
	enc, err := transbank.ExportEncodeAmount(1500, paymentmanager.CurrencyUSD)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if enc != 15.00 {
		t.Fatalf("expected 15.00, got %f", enc)
	}

	dec, err := transbank.ExportDecodeAmount(15.00, paymentmanager.CurrencyUSD)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dec != 1500 {
		t.Fatalf("expected 1500, got %d", dec)
	}
}

func TestAmount_Unsupported(t *testing.T) {
	_, err := transbank.ExportEncodeAmount(1500, "EUR")
	if err != transbank.ErrCurrency {
		t.Fatalf("expected ErrCurrency, got %v", err)
	}

	_, err = transbank.ExportDecodeAmount(15.00, "EUR")
	if err != transbank.ErrCurrency {
		t.Fatalf("expected ErrCurrency, got %v", err)
	}
}
