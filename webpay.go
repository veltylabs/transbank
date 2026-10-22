package transbank

import paymentmanager "github.com/veltylabs/payment_manager"

var _ paymentmanager.Gateway = (*Webpay)(nil)

// Límites de Webpay para la creación.
const (
	MaxBuyOrder  = 26
	MaxSessionID = 61
	MaxReturnURL = 256
)

// Estados y tipos que informa Webpay.
const (
	StatusInitialized        = "INITIALIZED"
	StatusAuthorized         = "AUTHORIZED"
	StatusFailed             = "FAILED"
	StatusReversed           = "REVERSED"
	StatusNullified          = "NULLIFIED"
	StatusPartiallyNullified = "PARTIALLY_NULLIFIED"
	RefundTypeReversed       = "REVERSED"
	RefundTypeNullified      = "NULLIFIED"
)

// ProviderName es el nombre que payment_manager guarda en cada pago de este adaptador.
const ProviderName = "transbank"

// TokenField es el campo con que el navegador envía el token a Webpay (POST).
const TokenField = "token_ws"

// Provider devuelve el nombre del proveedor ("transbank").
// Ejemplo: name := wp.Provider()
func (w *Webpay) Provider() string { return ProviderName }

// Start crea la transacción; el navegador hace POST a RedirectURL con TokenField.
// Ejemplo: s, err := wp.Start(req)
func (w *Webpay) Start(req paymentmanager.GatewayStart) (paymentmanager.GatewayStarted, error) {
	if w.cfg.ReturnURL == "" {
		return paymentmanager.GatewayStarted{}, ErrNoReturnURL
	}
	if len(req.PaymentId) > MaxBuyOrder || len(w.cfg.ReturnURL) > MaxReturnURL {
		return paymentmanager.GatewayStarted{}, ErrFieldTooLong
	}
	amount, err := toWebpay(req.Amount, req.Currency)
	if err != nil {
		return paymentmanager.GatewayStarted{}, err
	}
	var res CreateResponse
	if err := w.call(methodPost, TransactionsPath, &CreateRequest{BuyOrder: req.PaymentId, SessionId: req.PaymentId, Amount: amount, ReturnUrl: w.cfg.ReturnURL}, &res); err != nil {
		return paymentmanager.GatewayStarted{}, err
	}
	return paymentmanager.GatewayStarted{Token: res.Token, RedirectURL: res.Url, TokenField: TokenField}, nil
}

// commit confirma la transacción al volver el cliente (PUT).
func (w *Webpay) commit(token string) (paymentmanager.GatewayResult, string, error) {
	return w.transaction(methodPut, token)
}

// Status consulta la transacción (GET); Transbank la guarda 7 días.
// Ejemplo: res, err := wp.Status(token)
func (w *Webpay) Status(token string) (paymentmanager.GatewayResult, error) {
	res, _, err := w.transaction(methodGet, token)
	return res, err
}

func (w *Webpay) transaction(method, token string) (paymentmanager.GatewayResult, string, error) {
	var tr TransactionResponse
	if err := w.call(method, TransactionsPath+"/"+token, nil, &tr); err != nil {
		return paymentmanager.GatewayResult{}, "", err
	}
	res := paymentmanager.GatewayResult{
		Amount:            fromWebpay(tr.Amount),
		AuthorizationCode: tr.AuthorizationCode,
		PaymentTypeCode:   tr.PaymentTypeCode,
		Installments:      tr.InstallmentsNumber,
	}
	if n := len(tr.CardDetail.CardNumber); n >= 4 {
		res.CardLast4 = tr.CardDetail.CardNumber[n-4:]
	}
	switch tr.Status { // diagrama §4
	case StatusAuthorized:
		res.Outcome = paymentmanager.OutcomeAuthorized
		if tr.ResponseCode != 0 { // S2
			res.Outcome = paymentmanager.OutcomeRejected
		}
	case StatusInitialized:
		res.Outcome = paymentmanager.OutcomePending
	case StatusFailed, StatusReversed, StatusNullified, StatusPartiallyNullified:
		res.Outcome = paymentmanager.OutcomeRejected
	default:
		return paymentmanager.GatewayResult{}, "", ErrUnknownStatus
	}
	return res, tr.BuyOrder, nil
}

// Refund devuelve amount (pesos). Webpay decide reversa (total, antes de 3 h)
// o anulación (diagrama §5).
// Ejemplo: res, err := wp.Refund(token, 1000)
func (w *Webpay) Refund(token string, amount int64) (paymentmanager.GatewayRefund, error) {
	var res RefundResponse
	if err := w.call(methodPost, TransactionsPath+"/"+token+refundsPathSuffix, &RefundRequest{Amount: float64(amount)}, &res); err != nil {
		return paymentmanager.GatewayRefund{}, err
	}
	switch res.Type {
	case RefundTypeReversed:
		return paymentmanager.GatewayRefund{Kind: paymentmanager.RefundKindReversed, Amount: amount}, nil
	case RefundTypeNullified:
		if res.ResponseCode != 0 {
			return paymentmanager.GatewayRefund{}, ErrRefundRejected
		}
		return paymentmanager.GatewayRefund{Kind: paymentmanager.RefundKindNullified, AuthorizationCode: res.AuthorizationCode, Amount: fromWebpay(res.NullifiedAmount)}, nil
	}
	return paymentmanager.GatewayRefund{}, ErrUnknownRefundType
}
