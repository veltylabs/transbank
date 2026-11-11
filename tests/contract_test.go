//go:build !wasm

package tests

import (
	"testing"

	paymentmanager "github.com/veltylabs/payment_manager"
	"github.com/veltylabs/transbank"
)

func TestWebpay_ImplementsGateway(t *testing.T) {
	var _ paymentmanager.Gateway = transbank.New(transbank.Config{ReturnURL: "https://example.com/return"})
}
