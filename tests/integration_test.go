//go:build transbank_integration
package tests

import (
	"context"
	"testing"
	"github.com/veltylabs/payment_manager"
	"github.com/veltylabs/transbank"
)

func TestIntegration(t *testing.T) {
	wp, err := transbank.Open("webpay://597055555532:579B532A7440BB0C9079DED94D31EA1615BACEB56610332264630D42D0A36B1C@integration")
	if err != nil {
		t.Fatalf("Failed to open Transbank Webpay: %v", err)
	}

	req := paymentmanager.GatewayStart{
		BuyOrder:  "O-INTEGRATION-TEST",
		SessionId: "S-1",
		Currency:  paymentmanager.CurrencyCLP,
		Amount:    1000,
		ReturnURL: "http://localhost:8080/return",
	}

	res, err := wp.Start(req)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	if res.Token == "" {
		t.Fatalf("Expected token, got empty")
	}
	if res.RedirectURL == "" {
		t.Fatalf("Expected redirect URL, got empty")
	}

	statusRes, err := wp.Status(res.Token)
	// It usually complains because we haven't hit the URL in a browser
	if err != nil {
		t.Logf("Status failed as expected before user interaction: %v", err)
	} else {
		t.Logf("Status returned outcome: %v", statusRes.Outcome)
	}

	_ = context.Background()
}
