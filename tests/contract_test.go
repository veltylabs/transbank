package tests

import (
	"testing"
	"github.com/veltylabs/payment_manager"
	"github.com/veltylabs/transbank"
)

func TestWebpay_ImplementsGateway(t *testing.T) {
	var _ paymentmanager.Gateway = (*transbank.Webpay)(nil)
}
