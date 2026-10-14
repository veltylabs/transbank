package tests

import (
	"github.com/veltylabs/payment_manager"
)

type fakeCompleter struct {
	lastReturn paymentmanager.OnlineReturn
	payment    paymentmanager.Payment
	err        error
}

func (f *fakeCompleter) CompleteOnlinePayment(ret paymentmanager.OnlineReturn) (paymentmanager.Payment, error) {
	f.lastReturn = ret
	return f.payment, f.err
}
