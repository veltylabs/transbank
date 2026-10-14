//go:build !wasm

package tests

import (
	paymentmanager "github.com/veltylabs/payment_manager"
)

const (
	testCode = "597055555532"
	testKey  = "SECRET-KEY-123"
)

// fakeCompleter registra con qué la llamó la ruta de retorno y devuelve lo
// configurado.
type fakeCompleter struct {
	got    []paymentmanager.OnlineReturn
	result paymentmanager.Payment
	err    error
}

func (c *fakeCompleter) CompleteOnlinePayment(ret paymentmanager.OnlineReturn) (paymentmanager.Payment, error) {
	c.got = append(c.got, ret)
	return c.result, c.err
}
