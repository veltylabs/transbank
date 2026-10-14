package tests

import (
	"testing"
	"github.com/veltylabs/payment_manager"
	"github.com/veltylabs/transbank"
)

func TestWebpay_Start(t *testing.T) {
	fake := newFakeWebpay(t)
	fake.response = `{"token": "T-123", "url": "https://webpay.cl/pay"}`

	wp := transbank.New(transbank.Config{BaseURL: fake.server.URL})

	req := paymentmanager.GatewayStart{
		BuyOrder:  "O-1",
		SessionId: "S-1",
		Currency:  paymentmanager.CurrencyCLP,
		Amount:    1000,
		ReturnURL: "http://return",
	}

	res, err := wp.Start(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.Token != "T-123" {
		t.Fatalf("expected Token T-123, got %s", res.Token)
	}
	if res.RedirectURL != "https://webpay.cl/pay?token_ws=T-123" {
		t.Fatalf("expected URL https://webpay.cl/pay?token_ws=T-123, got %s", res.RedirectURL)
	}
}

func TestWebpay_Start_BadCurrency(t *testing.T) {
	wp := transbank.New(transbank.Config{})
	req := paymentmanager.GatewayStart{
		Currency: "EUR",
	}
	_, err := wp.Start(req)
	if err == nil {
		t.Fatalf("expected error for bad currency")
	}
}

func TestWebpay_Start_CallError(t *testing.T) {
	wp := transbank.New(transbank.Config{BaseURL: "http://localhost:0"})
	req := paymentmanager.GatewayStart{
		Currency: paymentmanager.CurrencyCLP,
	}
	_, err := wp.Start(req)
	if err == nil {
		t.Fatalf("expected error for bad call")
	}
}

func TestWebpay_Start_LengthError(t *testing.T) {
	wp := transbank.New(transbank.Config{BaseURL: "http://localhost:0"})
	req := paymentmanager.GatewayStart{
		BuyOrder: "123456789012345678901234567890", // 30 chars
	}
	_, err := wp.Start(req)
	if err == nil {
		t.Fatalf("expected error for too long buy order")
	}
}

func TestWebpay_Commit(t *testing.T) {
	fake := newFakeWebpay(t)
	fake.response = `{"vci": "TSY", "amount": 1000, "status": "AUTHORIZED", "buy_order": "O-1", "session_id": "S-1", "card_detail": {"card_number": "1234"}, "accounting_date": "1234", "transaction_date": "1234", "authorization_code": "1234", "payment_type_code": "VN", "response_code": 0, "installments_amount": 0, "installments_number": 0, "balance": 0}`

	wp := transbank.New(transbank.Config{BaseURL: fake.server.URL})

	res, err := wp.Commit("T-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.Outcome != paymentmanager.OutcomeAuthorized {
		t.Fatalf("expected OutcomeAuthorized, got %v", res.Outcome)
	}
	if res.Amount != 1000 {
		t.Fatalf("expected Amount 1000, got %d", res.Amount)
	}

	if len(fake.requests) != 1 {
		t.Fatalf("expected 1 request")
	}
	if fake.requests[0].method != "PUT" {
		t.Fatalf("expected PUT method, got %s", fake.requests[0].method)
	}
	if fake.requests[0].path != "/rswebpaytransaction/api/webpay/v1.2/transactions/T-123" {
		t.Fatalf("expected path /rswebpaytransaction/api/webpay/v1.2/transactions/T-123, got %s", fake.requests[0].path)
	}
}

func TestWebpay_Commit_CallError(t *testing.T) {
	wp := transbank.New(transbank.Config{BaseURL: "http://localhost:0"})
	_, err := wp.Commit("T-123")
	if err == nil {
		t.Fatalf("expected error for bad call")
	}
}

func TestWebpay_Commit_FailedStatus(t *testing.T) {
	fake := newFakeWebpay(t)
	fake.response = `{"status": "FAILED", "response_code": -1}`
	wp := transbank.New(transbank.Config{BaseURL: fake.server.URL})
	res, err := wp.Commit("T-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Outcome != paymentmanager.OutcomeRejected {
		t.Fatalf("expected OutcomeRejected, got %v", res.Outcome)
	}
}

func TestWebpay_Status(t *testing.T) {
	fake := newFakeWebpay(t)
	fake.response = `{"vci": "TSY", "amount": 1000, "status": "AUTHORIZED", "buy_order": "O-1", "session_id": "S-1", "card_detail": {"card_number": "1234"}, "accounting_date": "1234", "transaction_date": "1234", "authorization_code": "1234", "payment_type_code": "VN", "response_code": 0, "installments_amount": 0, "installments_number": 0, "balance": 0}`

	wp := transbank.New(transbank.Config{BaseURL: fake.server.URL})

	res, err := wp.Status("T-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.Outcome != paymentmanager.OutcomeAuthorized {
		t.Fatalf("expected OutcomeAuthorized, got %v", res.Outcome)
	}
	if res.Amount != 1000 {
		t.Fatalf("expected Amount 1000, got %d", res.Amount)
	}

	if len(fake.requests) != 1 {
		t.Fatalf("expected 1 request")
	}
	if fake.requests[0].method != "GET" {
		t.Fatalf("expected GET method, got %s", fake.requests[0].method)
	}
}

func TestWebpay_Status_CallError(t *testing.T) {
	wp := transbank.New(transbank.Config{BaseURL: "http://localhost:0"})
	_, err := wp.Status("T-123")
	if err == nil {
		t.Fatalf("expected error for bad call")
	}
}

func TestWebpay_Refund_Reversed(t *testing.T) {
	fake := newFakeWebpay(t)
	fake.response = `{"type": "REVERSED", "nullified_amount": 1000, "response_code": 0}`

	wp := transbank.New(transbank.Config{BaseURL: fake.server.URL})

	res, err := wp.Refund("T-123", 1000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.Kind != paymentmanager.RefundKindReversed {
		t.Fatalf("expected Reversed kind")
	}

	if len(fake.requests) != 1 {
		t.Fatalf("expected 1 request")
	}
	if fake.requests[0].method != "POST" {
		t.Fatalf("expected POST method, got %s", fake.requests[0].method)
	}
	if fake.requests[0].path != "/rswebpaytransaction/api/webpay/v1.2/transactions/T-123/refunds" {
		t.Fatalf("expected path /rswebpaytransaction/api/webpay/v1.2/transactions/T-123/refunds, got %s", fake.requests[0].path)
	}
}

func TestWebpay_Refund_Nullified(t *testing.T) {
	fake := newFakeWebpay(t)
	fake.response = `{"type": "NULLIFIED", "nullified_amount": 1000, "authorization_code": "1234", "response_code": 0}`

	wp := transbank.New(transbank.Config{BaseURL: fake.server.URL})

	res, err := wp.Refund("T-123", 1000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.Kind != paymentmanager.RefundKindNullified {
		t.Fatalf("expected Nullified kind")
	}
	if res.Amount != 1000 {
		t.Fatalf("expected amount 1000")
	}
}

func TestWebpay_Refund_CallError(t *testing.T) {
	wp := transbank.New(transbank.Config{BaseURL: "http://localhost:0"})
	_, err := wp.Refund("T-123", 1000)
	if err == nil {
		t.Fatalf("expected error for bad call")
	}
}

func TestWebpay_Refund_FailedStatus(t *testing.T) {
	fake := newFakeWebpay(t)
	fake.response = `{"type": "NULLIFIED", "response_code": -1}`
	wp := transbank.New(transbank.Config{BaseURL: fake.server.URL})
	_, err := wp.Refund("T-123", 1000)
	if err == nil {
		t.Fatalf("expected error for rejected refund")
	}
}

func TestWebpay_Refund_Unknown(t *testing.T) {
	fake := newFakeWebpay(t)
	fake.response = `{"type": "UNKNOWN", "response_code": 0}`
	wp := transbank.New(transbank.Config{BaseURL: fake.server.URL})
	_, err := wp.Refund("T-123", 1000)
	if err == nil {
		t.Fatalf("expected error for unknown refund type")
	}
}

func TestWebpay_Commit_UnknownStatus3(t *testing.T) {
	fake := newFakeWebpay(t)
	fake.response = `{"vci": "TSY", "amount": 1000, "status": "UNKNOWN_STATUS", "response_code": 0}`

	wp := transbank.New(transbank.Config{BaseURL: fake.server.URL})

	_, err := wp.Commit("T-123")
	if err == nil {
		t.Fatalf("expected error for unknown status")
	}
}

func TestMapStatus(t *testing.T) {
	if transbank.ExportMapStatus("INITIALIZED") != "FAILED" {
		t.Fatalf("expected FAILED")
	}
	if transbank.ExportMapStatus("OTHER") != "FAILED" {
		t.Fatalf("expected FAILED")
	}
}
