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

func TestStart_ReturnsTokenAndURL(t *testing.T) {
	f := newFake(t)
	f.respond(200, `{"token":"tok-64","url":"https://webpay3gint.transbank.cl/webpayserver/initTransaction"}`)
	s, err := f.gateway().Start(paymentmanager.GatewayStart{BuyOrder: "b1", SessionId: "b1", Amount: 1000, Currency: paymentmanager.CurrencyCLP, ReturnURL: "https://r"})
	if err != nil || s.Token != "tok-64" || s.RedirectURL != "https://webpay3gint.transbank.cl/webpayserver/initTransaction" {
		t.Fatalf("got %+v, %v (the token travels by POST, not in the URL)", s, err)
	}
}

func TestStart_FieldTooLongSkipsAPI(t *testing.T) {
	f := newFake(t)
	cases := []paymentmanager.GatewayStart{
		{BuyOrder: long(transbank.MaxBuyOrder + 1), Currency: paymentmanager.CurrencyCLP},
		{BuyOrder: "b", SessionId: long(transbank.MaxSessionID + 1), Currency: paymentmanager.CurrencyCLP},
		{BuyOrder: "b", ReturnURL: long(transbank.MaxReturnURL + 1), Currency: paymentmanager.CurrencyCLP},
	}
	for i, c := range cases {
		if _, err := f.gateway().Start(c); err != transbank.ErrFieldTooLong {
			t.Errorf("case %d: %v", i, err)
		}
	}
	if len(f.requests) != 0 {
		t.Fatal("an invalid start never reaches Webpay")
	}
}

func TestStart_APIErrorPropagates(t *testing.T) {
	f := newFake(t)
	f.respond(401, `{"error_message":"Not Authorized"}`)
	if _, err := f.gateway().Start(paymentmanager.GatewayStart{BuyOrder: "b", Currency: paymentmanager.CurrencyCLP}); err == nil {
		t.Fatal("expected the API error")
	}
}

func TestCommit_Authorized(t *testing.T) {
	f := newFake(t)
	f.respond(200, `{"status":"AUTHORIZED","amount":25000,"response_code":0,"authorization_code":"1213","payment_type_code":"VN","installments_number":3,"card_detail":{"card_number":"XXXXXXXXXXXX6623"}}`)
	r, err := f.gateway().Commit("tok")
	if err != nil || r.Outcome != paymentmanager.OutcomeAuthorized || r.Amount != 25000 || r.AuthorizationCode != "1213" || r.PaymentTypeCode != "VN" || r.Installments != 3 || r.CardLast4 != "6623" {
		t.Fatalf("got %+v, %v", r, err)
	}
	if req := f.last(t); req.method != "PUT" || req.path != transbank.TransactionsPath+"/tok" {
		t.Fatalf("request = %+v", req)
	}
}

func TestCommit_AuthorizedWithNonZeroCodeIsRejected(t *testing.T) {
	f := newFake(t)
	f.respond(200, `{"status":"AUTHORIZED","amount":25000,"response_code":-1}`)
	if r, _ := f.gateway().Commit("tok"); r.Outcome != paymentmanager.OutcomeRejected {
		t.Fatalf("outcome = %s", r.Outcome)
	}
}

func TestCommit_ErrorPropagates(t *testing.T) {
	f := newFake(t)
	f.respond(500, ``)
	if _, err := f.gateway().Commit("tok"); err == nil {
		t.Fatal("expected an error")
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
	f.respond(500, ``)
	if _, err := f.gateway().Refund("tok", 1000); err == nil {
		t.Fatal("expected an error")
	}
}
