//go:build !wasm
package tests

import (
	"testing"
	"webtyp.com/json"
	"github.com/veltylabs/transbank"
)

func TestCall_SendsHeadersAndBody(t *testing.T) {
	fake := newFakeWebpay(t)
	fake.response = `{"token": "T-123"}`

	wp := transbank.New(transbank.Config{
		CommerceCode: "C-123",
		APIKey:       "K-123",
		BaseURL:      fake.server.URL,
		TimeoutMs:    1000,
	})

	req := &transbank.CreateRequest{
		BuyOrder: "O-1",
	}
	var res transbank.CreateResponse

	err := wp.ExportCall("POST", "/test", req, &res)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fake.requests) != 1 {
		t.Fatalf("expected 1 request, got %d", len(fake.requests))
	}

	r := fake.requests[0]
	if r.method != "POST" {
		t.Fatalf("expected POST, got %s", r.method)
	}
	if r.path != "/test" {
		t.Fatalf("expected /test, got %s", r.path)
	}
	if r.header.Get("Tbk-Api-Key-Id") != "C-123" {
		t.Fatalf("expected Tbk-Api-Key-Id C-123, got %s", r.header.Get("Tbk-Api-Key-Id"))
	}
	if r.header.Get("Tbk-Api-Key-Secret") != "K-123" {
		t.Fatalf("expected Tbk-Api-Key-Secret K-123, got %s", r.header.Get("Tbk-Api-Key-Secret"))
	}
	if r.header.Get("Content-Type") != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %s", r.header.Get("Content-Type"))
	}

	var reqBody transbank.CreateRequest
	_ = reqBody
	if err := json.Decode([]byte(r.body), &reqBody); err != nil { t.Logf("decode error: %v", err) }
	if reqBody.BuyOrder != "O-1" {
		t.Fatalf("expected body with BuyOrder O-1, got %s", r.body)
	}
}

func TestCall_NetworkError(t *testing.T) {
	wp := transbank.New(transbank.Config{
		CommerceCode: "C-123",
		APIKey:       "K-123",
		BaseURL:      "http://localhost:0", // invalid URL for network error
		TimeoutMs:    1000,
	})

	var res transbank.CreateResponse
	err := wp.ExportCall("POST", "/test", nil, &res)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if len(err.Error()) < 19 || err.Error()[:19] != "transbank: network:" {
		t.Fatalf("expected transbank: network: <err>, got %v", err)
	}
}

func TestCall_HTTPErrorWithMessage(t *testing.T) {
	fake := newFakeWebpay(t)
	fake.status = 400
	fake.response = `{"error_message": "Invalid amount"}`

	wp := transbank.New(transbank.Config{
		CommerceCode: "C-123",
		APIKey:       "K-123",
		BaseURL:      fake.server.URL,
		TimeoutMs:    1000,
	})

	var res transbank.CreateResponse
	err := wp.ExportCall("POST", "/test", nil, &res)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if err.Error() != "transbank: HTTP 400: Invalid amount" {
		t.Fatalf("expected transbank: HTTP 400: Invalid amount, got %v", err)
	}
}

func TestCall_HTTPErrorWithoutMessage(t *testing.T) {
	fake := newFakeWebpay(t)
	fake.status = 500
	fake.response = `{"other": "data"}`

	wp := transbank.New(transbank.Config{
		CommerceCode: "C-123",
		APIKey:       "K-123",
		BaseURL:      fake.server.URL,
		TimeoutMs:    1000,
	})

	var res transbank.CreateResponse
	err := wp.ExportCall("POST", "/test", nil, &res)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if err.Error() != "transbank: HTTP 500" {
		t.Fatalf("expected transbank: HTTP 500, got %v", err)
	}
}

func TestCall_BadJSON(t *testing.T) {
	fake := newFakeWebpay(t)
	fake.response = `{"token": invalid}`

	wp := transbank.New(transbank.Config{
		CommerceCode: "C-123",
		APIKey:       "K-123",
		BaseURL:      fake.server.URL,
		TimeoutMs:    1000,
	})

	var res transbank.CreateResponse
	err := wp.ExportCall("POST", "/test", nil, &res)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if len(err.Error()) < 23 || err.Error()[:23] != "transbank: bad response" {
		t.Fatalf("expected transbank: bad response, got %v", err)
	}
}

func TestCall_HTTPErrorWithBadJSONForApiErr(t *testing.T) {
	fake := newFakeWebpay(t)
	fake.status = 400
	fake.response = `bad`

	wp := transbank.New(transbank.Config{
		CommerceCode: "C-123",
		APIKey:       "K-123",
		BaseURL:      fake.server.URL,
		TimeoutMs:    1000,
	})

	var res transbank.CreateResponse
	err := wp.ExportCall("POST", "/test", nil, &res)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if err.Error() != "transbank: HTTP 400" {
		t.Fatalf("expected transbank: HTTP 400, got %v", err)
	}
}

func TestCall_UnexpectedResponse(t *testing.T) {
	fake := newFakeWebpay(t)
	fake.response = `bad`

	wp := transbank.New(transbank.Config{
		CommerceCode: "C-123",
		APIKey:       "K-123",
		BaseURL:      fake.server.URL,
		TimeoutMs:    1000,
	})

	var res transbank.CreateResponse
	err := wp.ExportCall("POST", "/test", nil, &res)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestCall_NetworkErrorViaErrCh(t *testing.T) {
	wp := transbank.New(transbank.Config{
		CommerceCode: "C-123",
		APIKey:       "K-123",
		BaseURL:      "http://10.255.255.1", // unroutable, should timeout or err
		TimeoutMs:    1,
	})

	var res transbank.CreateResponse
	err := wp.ExportCall("POST", "/test", nil, &res)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestCall_DeleteMethod(t *testing.T) {
	wp := transbank.New(transbank.Config{BaseURL: "http://localhost:0"})
	err := wp.ExportCall("DELETE", "/test", nil, nil)
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestCall_PutMethod(t *testing.T) {
	wp := transbank.New(transbank.Config{BaseURL: "http://localhost:0"})
	err := wp.ExportCall("PUT", "/test", nil, nil)
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestCall_GetMethod(t *testing.T) {
	wp := transbank.New(transbank.Config{BaseURL: "http://localhost:0"})
	err := wp.ExportCall("GET", "/test", nil, nil)
	if err == nil {
		t.Fatalf("expected error")
	}
}
