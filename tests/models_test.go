//go:build !wasm

package tests

import (
	"testing"

	"webtyp.com/json"
	"webtyp.com/model"

	"github.com/veltylabs/transbank"
)

// roundTrip codifica y decodifica v en una instancia nueva y compara el JSON.
func roundTrip(t *testing.T, name string, v model.Encodable, fresh model.Decodable) {
	t.Helper()
	var first, second []byte
	if err := json.Encode(v, &first); err != nil {
		t.Fatalf("%s encode: %v", name, err)
	}
	if err := json.Decode(string(first), fresh); err != nil {
		t.Fatalf("%s decode: %v", name, err)
	}
	if err := json.Encode(fresh.(model.Encodable), &second); err != nil {
		t.Fatalf("%s re-encode: %v", name, err)
	}
	if string(first) != string(second) {
		t.Fatalf("%s round trip:\n%s\n%s", name, first, second)
	}
}

// Los nombres JSON son los de la API de Transbank: un cambio en model.go que
// los rompa lo atrapa esta prueba, no producción.
func TestModels_RoundTrip(t *testing.T) {
	roundTrip(t, "create_request", &transbank.CreateRequest{BuyOrder: "b1", SessionId: "s1", Amount: 1000, ReturnUrl: "https://r"}, &transbank.CreateRequest{})
	roundTrip(t, "create_response", &transbank.CreateResponse{Token: "t", Url: "https://u"}, &transbank.CreateResponse{})
	roundTrip(t, "card_detail", &transbank.CardDetail{CardNumber: "6623"}, &transbank.CardDetail{})
	roundTrip(t, "transaction_response", &transbank.TransactionResponse{Vci: "TSY", Amount: 1000, Status: transbank.StatusAuthorized, BuyOrder: "b1", SessionId: "s1",
		CardDetail: transbank.CardDetail{CardNumber: "6623"}, AccountingDate: "0522", TransactionDate: "2026-05-22T19:00:00.000Z", AuthorizationCode: "1213",
		PaymentTypeCode: "VN", ResponseCode: 0, InstallmentsAmount: 500, InstallmentsNumber: 2, Balance: 1000}, &transbank.TransactionResponse{})
	roundTrip(t, "refund_request", &transbank.RefundRequest{Amount: 1000}, &transbank.RefundRequest{})
	roundTrip(t, "refund_response", &transbank.RefundResponse{Type: transbank.RefundTypeNullified, AuthorizationCode: "1", AuthorizationDate: "d", Balance: 0, NullifiedAmount: 1000, ResponseCode: 0}, &transbank.RefundResponse{})
	roundTrip(t, "api_error", &transbank.APIError{ErrorMessage: "m"}, &transbank.APIError{})
}

// fielderList es la forma que ormc genera para cada lista de modelos.
type fielderList interface {
	Len() int
	At(i int) model.Fielder
	Append() model.Fielder
	IsNil() bool
	EncodeFields(w model.FieldWriter)
	DecodeFields(r model.FieldReader)
}

func TestModels_GeneratedLists(t *testing.T) {
	cases := []struct {
		name string
		list fielderList
	}{
		{"create_request", &transbank.CreateRequestList{}}, {"create_response", &transbank.CreateResponseList{}},
		{"card_detail", &transbank.CardDetailList{}}, {"transaction_response", &transbank.TransactionResponseList{}},
		{"refund_request", &transbank.RefundRequestList{}}, {"refund_response", &transbank.RefundResponseList{}},
		{"api_error", &transbank.APIErrorList{}},
	}
	for _, c := range cases {
		if c.list.IsNil() || c.list.Len() != 0 {
			t.Errorf("%s: a new list is empty and non-nil", c.name)
		}
		added := c.list.Append()
		if c.list.Len() != 1 || c.list.At(0) != added {
			t.Errorf("%s: Append then At(0) returns the appended record", c.name)
		}
		got, ok := added.(model.Model)
		if !ok || got.ModelName() != c.name {
			t.Errorf("%s: element model name = %v", c.name, added)
			continue
		}
		if got.IsNil() || len(got.Schema()) != len(got.Pointers()) {
			t.Errorf("%s: %d schema fields, %d pointers", c.name, len(got.Schema()), len(got.Pointers()))
		}
		if v, ok := added.(interface{ Validate(byte) error }); !ok || v.Validate(model.ActionCreate) != nil {
			t.Errorf("%s: a zero value is a valid Webpay payload", c.name)
		}
		// Encode/DecodeFields de la lista no escriben nada: el codec recorre los elementos.
		c.list.EncodeFields(nil)
		c.list.DecodeFields(nil)
		if c.list.Len() != 1 {
			t.Errorf("%s: list Encode/DecodeFields must not change it", c.name)
		}
	}
}
