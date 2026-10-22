//go:build !wasm

package tests

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/veltylabs/transbank"
)

// request es un pedido que recibió el servidor falso.
type request struct {
	method, path, keyID, keySecret, contentType, body string
}

// fakeWebpay es un servidor Webpay de prueba: guarda cada pedido y responde
// status/body configurados por la prueba.
type fakeWebpay struct {
	srv      *httptest.Server
	requests []request
	status   int
	body     string
}

func newFake(t *testing.T) *fakeWebpay {
	t.Helper()
	f := &fakeWebpay{status: 200, body: `{}`}
	f.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		f.requests = append(f.requests, request{r.Method, r.URL.Path, r.Header.Get(transbank.HeaderKeyID), r.Header.Get(transbank.HeaderKeySecret), r.Header.Get("Content-Type"), string(b)})
		w.WriteHeader(f.status)
		_, _ = w.Write([]byte(f.body))
	}))
	t.Cleanup(f.srv.Close)
	return f
}

func (f *fakeWebpay) respond(status int, body string) { f.status, f.body = status, body }

func (f *fakeWebpay) gateway() *transbank.Webpay {
	return transbank.New(transbank.Config{CommerceCode: testCode, APIKey: testKey, BaseURL: f.srv.URL})
}

func (f *fakeWebpay) last(t *testing.T) request {
	t.Helper()
	if len(f.requests) == 0 {
		t.Fatal("ningún pedido llegó al Webpay falso")
	}
	return f.requests[len(f.requests)-1]
}

func delayed(h http.Handler, before func()) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		before()
		h.ServeHTTP(w, r)
	})
}
