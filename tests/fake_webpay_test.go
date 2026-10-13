package tests

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

type requestLog struct {
	method string
	path   string
	header http.Header
	body   string
}

type fakeWebpay struct {
	t        *testing.T
	server   *httptest.Server
	requests []requestLog
	status   int
	response string
}

func newFakeWebpay(t *testing.T) *fakeWebpay {
	fake := &fakeWebpay{
		t:      t,
		status: 200,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, _ := io.ReadAll(r.Body)
		defer r.Body.Close()

		fake.requests = append(fake.requests, requestLog{
			method: r.Method,
			path:   r.URL.Path,
			header: r.Header,
			body:   string(bodyBytes),
		})

		w.WriteHeader(fake.status)
		fmt.Fprint(w, fake.response)
	})

	fake.server = httptest.NewServer(mux)
	t.Cleanup(fake.server.Close)

	return fake
}
