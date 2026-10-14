package transbank

import (
	"github.com/veltylabs/payment_manager"
	"webtyp.com/fmt"
)

var ErrCurrency = fmt.Err("transbank: unsupported currency")

func encodeAmount(amount int64, currency string) (float64, error) {
	switch currency {
	case paymentmanager.CurrencyCLP:
		return float64(amount), nil
	case paymentmanager.CurrencyUSD:
		return float64(amount) / 100.0, nil
	default:
		return 0, ErrCurrency
	}
}

func decodeAmount(amount float64, currency string) (int64, error) {
	switch currency {
	case paymentmanager.CurrencyCLP:
		return int64(amount), nil
	case paymentmanager.CurrencyUSD:
		return int64(amount * 100.0), nil
	default:
		return 0, ErrCurrency
	}
}
