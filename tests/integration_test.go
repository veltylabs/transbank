//go:build transbank_integration

package tests

import (
	"testing"

	paymentmanager "github.com/veltylabs/payment_manager"
	"github.com/veltylabs/transbank"
)

// Corre contra el ambiente de integración real de Transbank con sus
// credenciales públicas: go test -tags transbank_integration ./tests/
func TestIntegration_StartThenStatusIsPending(t *testing.T) {
	w, err := transbank.Open("webpay://" + transbank.IntegrationCommerceCode + ":" + transbank.IntegrationAPIKey + "@" + transbank.EnvironmentIntegration)
	if err != nil {
		t.Fatal(err)
	}
	s, err := w.Start(paymentmanager.GatewayStart{BuyOrder: "it-transbank", SessionId: "it-transbank", Amount: 1000, Currency: paymentmanager.CurrencyCLP, ReturnURL: "https://example.com" + transbank.ReturnPath})
	if err != nil || s.Token == "" || s.RedirectURL == "" {
		t.Fatalf("Start: %+v, %v", s, err)
	}
	r, err := w.Status(s.Token)
	if err != nil || r.Outcome != paymentmanager.OutcomePending {
		t.Fatalf("Status: %+v, %v", r, err)
	}
}
