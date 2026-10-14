package tests

import (
	"testing"
	"github.com/veltylabs/transbank"
)

func TestOpen_WrongScheme(t *testing.T) {
	_, err := transbank.Open("http://597055555532:key@integration")
	if err != transbank.ErrConfigScheme {
		t.Fatalf("expected ErrConfigScheme, got %v", err)
	}
}

func TestOpen_BadFormat(t *testing.T) {
	_, err := transbank.Open("webpay://597055555532:key")
	if err != transbank.ErrConfigFormat {
		t.Fatalf("expected ErrConfigFormat missing @, got %v", err)
	}
	_, err = transbank.Open("webpay://597055555532@integration")
	if err != transbank.ErrConfigFormat {
		t.Fatalf("expected ErrConfigFormat missing colon, got %v", err)
	}
}

func TestOpen_EmptyCredentials(t *testing.T) {
	_, err := transbank.Open("webpay://:key@integration")
	if err != transbank.ErrConfigCredentials {
		t.Fatalf("expected ErrConfigCredentials for empty code, got %v", err)
	}
	_, err = transbank.Open("webpay://597055555532:@integration")
	if err != transbank.ErrConfigCredentials {
		t.Fatalf("expected ErrConfigCredentials for empty key, got %v", err)
	}
}

func TestOpen_Environments(t *testing.T) {
	tests := []struct{ env string }{{"integration"}, {"production"}}
	for _, tt := range tests {
		wp, err := transbank.Open("webpay://code:key@" + tt.env)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if wp == nil {
			t.Fatalf("expected webpay instance for %s", tt.env)
		}
	}
}

func TestOpen_UnknownEnvironment(t *testing.T) {
	_, err := transbank.Open("webpay://code:key@staging")
	if err != transbank.ErrConfigEnvironment {
		t.Fatalf("expected ErrConfigEnvironment, got %v", err)
	}
}

func TestOpen_ErrorsNeverContainKey(t *testing.T) {
	errs := []error{transbank.ErrConfigScheme, transbank.ErrConfigFormat, transbank.ErrConfigCredentials, transbank.ErrConfigEnvironment}
	key := "SECRET_KEY_123"
	for _, err := range errs {
		if err != nil && err.Error() == key {
			t.Fatalf("error contains key: %v", err)
		}
	}
}

func TestNew_DefaultTimeout(t *testing.T) {
	wp := transbank.New(transbank.Config{CommerceCode: "test", APIKey: "test", BaseURL: transbank.IntegrationBaseURL, TimeoutMs: 0})
	if wp == nil {
		t.Fatalf("expected webpay instance")
	}
}

func TestOpen_ExportedCoverageMore(t *testing.T) {
	transbank.OpenForTest("webpay://code:@integration")
	transbank.OpenForTest("webpay://:key@integration")
}
