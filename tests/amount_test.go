//go:build !wasm

package tests

import (
	"testing"

	"webtyp.com/fmt"

	paymentmanager "github.com/veltylabs/payment_manager"
	"github.com/veltylabs/transbank"
)

func TestAmount_CLP(t *testing.T) {
	f := newFake(t)
	w := transbank.New(transbank.Config{ReturnURL: "https://r", BaseURL: f.srv.URL})
	if _, err := w.Start(paymentmanager.GatewayStart{PaymentId: "b1", Amount: 25000, Currency: paymentmanager.CurrencyCLP}); err != nil {
		t.Fatal(err)
	}
	if body := f.last(t).body; !fmt.Contains(body, `"amount":25000`) {
		t.Fatalf("CLP viaja como entero: %s", body)
	}
	f.respond(200, `{"status":"AUTHORIZED","amount":25000,"response_code":0,"buy_order":"b1"}`)
	r, err := w.Status("tok")
	if err != nil || r.Amount != 25000 {
		t.Fatalf("got %+v, %v", r, err)
	}
}

// USD se rechaza: Commit, Status y Refund del Gateway no traen la moneda, así
// que un monto USD no se podría devolver a unidades mínimas (desviación del
// plan, ver docs/ARCHITECTURE.md).
func TestAmount_USD(t *testing.T) {
	f := newFake(t)
	w := transbank.New(transbank.Config{ReturnURL: "https://r", BaseURL: f.srv.URL})
	_, err := w.Start(paymentmanager.GatewayStart{PaymentId: "b1", Amount: 1250, Currency: paymentmanager.CurrencyUSD})
	if err != transbank.ErrCurrency || len(f.requests) != 0 {
		t.Fatalf("err = %v, requests = %d", err, len(f.requests))
	}
}

func TestAmount_Unsupported(t *testing.T) {
	f := newFake(t)
	w := transbank.New(transbank.Config{ReturnURL: "https://r", BaseURL: f.srv.URL})
	_, err := w.Start(paymentmanager.GatewayStart{PaymentId: "b1", Amount: 1250, Currency: "EUR"})
	if err != transbank.ErrCurrency || len(f.requests) != 0 {
		t.Fatalf("err = %v, requests = %d (un monto que no sea CLP nunca llega a Webpay)", err, len(f.requests))
	}
}
