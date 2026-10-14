package transbank

import paymentmanager "github.com/veltylabs/payment_manager"

// Webpay Plus nacional opera en pesos. El puerto Gateway no informa la moneda
// en Commit, Status ni Refund, así que el adaptador solo acepta CLP: un monto en
// otra moneda se rechaza al iniciar (Start) en vez de convertirse mal después.

// toWebpay pasa unidades mínimas al monto de Webpay (CLP no tiene decimales).
func toWebpay(amount int64, currency string) (float64, error) {
	if currency != paymentmanager.CurrencyCLP {
		return 0, ErrCurrency
	}
	return float64(amount), nil
}

// fromWebpay pasa un monto de Webpay a unidades mínimas (CLP, redondeado).
func fromWebpay(amount float64) int64 { return int64(amount + 0.5) }
