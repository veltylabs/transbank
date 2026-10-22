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
	got    []paymentmanager.OnlineResult
	result paymentmanager.Payment
	err    error
}

func (c *fakeCompleter) CompleteOnlinePayment(r paymentmanager.OnlineResult) (paymentmanager.Payment, error) {
	c.got = append(c.got, r)
	return c.result, c.err
}
