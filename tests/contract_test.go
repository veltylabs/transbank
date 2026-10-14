//go:build !wasm

package tests

import (
	"testing"

	paymentmanager "github.com/veltylabs/payment_manager"
	"github.com/veltylabs/transbank"
)

func TestWebpay_ImplementsGateway(t *testing.T) {
	var g paymentmanager.Gateway = transbank.New(transbank.Config{})
	if g == nil {
		t.Fatal("New returns a Gateway")
	}
}
