//go:build !wasm

package tests

import (
	"testing"

	paymentmanager "github.com/veltylabs/payment_manager"
	"github.com/veltylabs/transbank"
)

func long(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = 'x'
	}
	return string(b)
}

func TestProvider_IsTransbank(t *testing.T) {
	if newFake(t).gateway().Provider() != "transbank" {
		t.Fatalf("Provider = %q", newFake(t).gateway().Provider())
	}
}

func TestStart_ReturnsTokenAndURL(t *testing.T) {
	f := newFake(t)
	f.respond(200, `{"token":"tok-64","url":"https://webpay3gint.transbank.cl/webpayserver/initTransaction"}`)
	w := transbank.New(transbank.Config{ReturnURL: "https://r", BaseURL: f.srv.URL})
	s, err := w.Start(paymentmanager.GatewayStart{PaymentId: "b1", Amount: 1000, Currency: paymentmanager.CurrencyCLP})
	if err != nil || s.Token != "tok-64" || s.RedirectURL != "https://webpay3gint.transbank.cl/webpayserver/initTransaction" {
		t.Fatalf("got %+v, %v (el token viaja por POST, no en la URL)", s, err)
	}
}

func TestStart_WithoutReturnUrlIsErrNoReturnUrl(t *testing.T) {
	if _, err := newFake(t).gateway().Start(paymentmanager.GatewayStart{PaymentId: "b1", Amount: 1000, Currency: paymentmanager.CurrencyCLP}); err != transbank.ErrNoReturnURL {
		t.Fatalf("err = %v", err)
	}
}

func TestStart_ReturnsTokenField(t *testing.T) {
	f := newFake(t)
	f.respond(200, `{"token":"tok-64","url":"https://webpay3gint.transbank.cl/webpayserver/initTransaction"}`)
	w := transbank.New(transbank.Config{ReturnURL: "https://r", BaseURL: f.srv.URL})
	s, _ := w.Start(paymentmanager.GatewayStart{PaymentId: "b1", Amount: 1000, Currency: paymentmanager.CurrencyCLP})
	if s.TokenField != "token_ws" {
		t.Fatalf("TokenField = %q", s.TokenField)
	}
}

func TestStart_FieldTooLongSkipsAPI(t *testing.T) {
	f := newFake(t)
	w := transbank.New(transbank.Config{ReturnURL: "https://r", BaseURL: f.srv.URL})
	wLongUrl := transbank.New(transbank.Config{ReturnURL: long(transbank.MaxReturnURL + 1), BaseURL: f.srv.URL})

	if _, err := w.Start(paymentmanager.GatewayStart{PaymentId: long(transbank.MaxBuyOrder + 1), Currency: paymentmanager.CurrencyCLP}); err != transbank.ErrFieldTooLong {
		t.Errorf("long buy order: %v", err)
	}
	if _, err := wLongUrl.Start(paymentmanager.GatewayStart{PaymentId: "b", Currency: paymentmanager.CurrencyCLP}); err != transbank.ErrFieldTooLong {
		t.Errorf("long return url: %v", err)
	}
	if len(f.requests) != 0 {
		t.Fatal("un inicio inválido nunca llega a Webpay")
	}
}

func TestStart_APIErrorPropagates(t *testing.T) {
	f := newFake(t)
	w := transbank.New(transbank.Config{ReturnURL: "https://r", BaseURL: f.srv.URL})
	f.respond(401, `{"error_message":"Not Authorized"}`)
	if _, err := w.Start(paymentmanager.GatewayStart{PaymentId: "b", Currency: paymentmanager.CurrencyCLP}); err == nil {
		t.Fatal("se esperaba el error de la API")
	}
}

func TestStatus_InitializedIsPending(t *testing.T) {
	f := newFake(t)
	f.respond(200, `{"status":"INITIALIZED","amount":25000}`)
	r, err := f.gateway().Status("tok")
	if err != nil || r.Outcome != paymentmanager.OutcomePending || r.CardLast4 != "" {
		t.Fatalf("got %+v, %v", r, err)
	}
	if req := f.last(t); req.method != "GET" {
		t.Fatalf("method %s", req.method)
	}
}

func TestStatus_TerminalStatesAreRejected(t *testing.T) {
	f := newFake(t)
	for _, st := range []string{transbank.StatusFailed, transbank.StatusReversed, transbank.StatusNullified, transbank.StatusPartiallyNullified} {
		f.respond(200, `{"status":"`+st+`"}`)
		if r, err := f.gateway().Status("tok"); err != nil || r.Outcome != paymentmanager.OutcomeRejected {
			t.Errorf("%s: %+v, %v", st, r, err)
		}
	}
}

func TestStatus_UnknownStatus(t *testing.T) {
	f := newFake(t)
	f.respond(200, `{"status":"WEIRD"}`)
	if _, err := f.gateway().Status("tok"); err != transbank.ErrUnknownStatus {
		t.Fatalf("err = %v", err)
	}
}

func TestRefund_Reversed(t *testing.T) {
	f := newFake(t)
	f.respond(200, `{"type":"REVERSED"}`)
	r, err := f.gateway().Refund("tok", 25000)
	if err != nil || r.Kind != paymentmanager.RefundKindReversed || r.Amount != 25000 {
		t.Fatalf("got %+v, %v", r, err)
	}
	if req := f.last(t); req.method != "POST" || req.path != transbank.TransactionsPath+"/tok/refunds" || req.body != `{"amount":25000}` {
		t.Fatalf("request = %+v", req)
	}
}

func TestRefund_Nullified(t *testing.T) {
	f := newFake(t)
	f.respond(200, `{"type":"NULLIFIED","response_code":0,"authorization_code":"123456","nullified_amount":1000,"balance":24000}`)
	r, err := f.gateway().Refund("tok", 1000)
	if err != nil || r.Kind != paymentmanager.RefundKindNullified || r.Amount != 1000 || r.AuthorizationCode != "123456" {
		t.Fatalf("got %+v, %v", r, err)
	}
}

func TestRefund_NullifiedRejected(t *testing.T) {
	f := newFake(t)
	f.respond(200, `{"type":"NULLIFIED","response_code":-1}`)
	if _, err := f.gateway().Refund("tok", 1000); err != transbank.ErrRefundRejected {
		t.Fatalf("err = %v", err)
	}
}

func TestRefund_UnknownType(t *testing.T) {
	f := newFake(t)
	f.respond(200, `{"type":"OTHER"}`)
	if _, err := f.gateway().Refund("tok", 1000); err != transbank.ErrUnknownRefundType {
		t.Fatalf("err = %v", err)
	}
}

func TestRefund_ErrorPropagates(t *testing.T) {
	f := newFake(t)
	w := transbank.New(transbank.Config{BaseURL: f.srv.URL})
	f.respond(500, ``)
	if _, err := w.Refund("tok", 1000); err == nil {
		t.Fatal("se esperaba un error")
	}
}
