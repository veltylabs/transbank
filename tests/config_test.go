//go:build !wasm

package tests

import (
	"testing"
	"time"

	"webtyp.com/fmt"

	"github.com/veltylabs/transbank"
)

func TestOpen_WrongScheme(t *testing.T) {
	for _, raw := range []string{"", "http://a:b@integration", "webpay:/a:b@integration"} {
		if _, err := transbank.Open(raw); err != transbank.ErrConfigScheme {
			t.Errorf("%q: %v", raw, err)
		}
	}
}

func TestOpen_BadFormat(t *testing.T) {
	for _, raw := range []string{"webpay://codekey@integration", "webpay://code:key", "webpay://code@key:integration"} {
		if _, err := transbank.Open(raw); err != transbank.ErrConfigFormat {
			t.Errorf("%q: %v", raw, err)
		}
	}
}

func TestOpen_EmptyCredentials(t *testing.T) {
	for _, raw := range []string{"webpay://:key@integration", "webpay://code:@integration"} {
		if _, err := transbank.Open(raw); err != transbank.ErrConfigCredentials {
			t.Errorf("%q: %v", raw, err)
		}
	}
}

func TestOpen_Environments(t *testing.T) {
	for _, env := range []string{transbank.EnvironmentIntegration, transbank.EnvironmentProduction} {
		w, err := transbank.Open("webpay://" + transbank.IntegrationCommerceCode + ":" + transbank.IntegrationAPIKey + "@" + env)
		if err != nil || w == nil {
			t.Errorf("%s: %v", env, err)
		}
	}
}

func TestOpen_UnknownEnvironment(t *testing.T) {
	if _, err := transbank.Open("webpay://a:b@staging"); err != transbank.ErrConfigEnvironment {
		t.Fatalf("err = %v", err)
	}
}

func TestOpen_ErrorsNeverContainKey(t *testing.T) {
	for _, raw := range []string{"webpay://a:" + testKey, "webpay://a:" + testKey + "@staging", "webpay://:" + testKey + "@integration"} {
		_, err := transbank.Open(raw)
		if err == nil || fmt.Contains(err.Error(), testKey) {
			t.Errorf("%q: %v", raw, err)
		}
	}
}

// Con TimeoutMs 0 rige DefaultTimeoutMs (30 s): un servidor que tarda 300 ms
// responde; con un tope explícito de 50 ms, la misma llamada falla por red.
func TestNew_DefaultTimeout(t *testing.T) {
	f := newFake(t)
	f.respond(200, `{"status":"INITIALIZED"}`)
	slow := func() { time.Sleep(300 * time.Millisecond) }
	f.srv.Config.Handler = delayed(f.srv.Config.Handler, slow)
	if _, err := transbank.New(transbank.Config{CommerceCode: "c", APIKey: "k", BaseURL: f.srv.URL}).Status("t"); err != nil {
		t.Fatalf("default timeout: %v", err)
	}
	_, err := transbank.New(transbank.Config{CommerceCode: "c", APIKey: "k", BaseURL: f.srv.URL, TimeoutMs: 50}).Status("t")
	if err == nil || !fmt.Contains(err.Error(), "transbank: network") {
		t.Fatalf("short timeout: %v", err)
	}
}
