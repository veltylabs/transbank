//go:build !wasm

package tests

import (
	"testing"

	"webtyp.com/fmt"

	paymentmanager "github.com/veltylabs/payment_manager"
	"github.com/veltylabs/transbank"
)

func TestCall_SendsHeadersAndBody(t *testing.T) {
	f := newFake(t)
	w := transbank.New(transbank.Config{CommerceCode: testCode, APIKey: testKey, ReturnURL: "https://clinic/return", BaseURL: f.srv.URL})
	f.respond(200, `{"token":"tok","url":"https://webpay.example/init"}`)
	if _, err := w.Start(paymentmanager.GatewayStart{PaymentId: "b1", Amount: 25000, Currency: paymentmanager.CurrencyCLP}); err != nil {
		t.Fatal(err)
	}
	r := f.last(t)
	if r.method != "POST" || r.path != transbank.TransactionsPath || r.keyID != testCode || r.keySecret != testKey || !fmt.Contains(r.contentType, "application/json") {
		t.Fatalf("request = %+v", r)
	}
	for _, want := range []string{`"buy_order":"b1"`, `"session_id":"b1"`, `"amount":25000`, `"return_url":"https://clinic/return"`} {
		if !fmt.Contains(r.body, want) {
			t.Errorf("el cuerpo %s carece de %s", r.body, want)
		}
	}
}

func TestCall_NetworkError(t *testing.T) {
	f := newFake(t)
	w := f.gateway()
	f.srv.Close()
	_, err := w.Status("tok")
	if err == nil || !fmt.Contains(err.Error(), "transbank: network") {
		t.Fatalf("err = %v", err)
	}
}

func TestCall_HTTPErrorWithMessage(t *testing.T) {
	f := newFake(t)
	f.respond(422, `{"error_message":"Invalid value for parameter: amount"}`)
	_, err := f.gateway().Status("tok")
	if err == nil || err.Error() != "transbank: HTTP 422: Invalid value for parameter: amount" {
		t.Fatalf("err = %v", err)
	}
}

func TestCall_HTTPErrorWithoutMessage(t *testing.T) {
	f := newFake(t)
	f.respond(500, `oops`)
	_, err := f.gateway().Status("tok")
	if err == nil || err.Error() != "transbank: HTTP 500" {
		t.Fatalf("err = %v", err)
	}
}

func TestCall_BadJSON(t *testing.T) {
	f := newFake(t)
	f.respond(200, `{not json`)
	if _, err := f.gateway().Status("tok"); err != transbank.ErrBadResponse {
		t.Fatalf("err = %v", err)
	}
}
