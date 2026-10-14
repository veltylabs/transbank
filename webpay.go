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

// Start crea la transacción; RedirectURL es la URL a la que el navegador hace
// POST con token_ws.
func (w *Webpay) Start(req paymentmanager.GatewayStart) (paymentmanager.GatewayStarted, error) {
	if len(req.BuyOrder) > MaxBuyOrder || len(req.SessionId) > MaxSessionID || len(req.ReturnURL) > MaxReturnURL {
		return paymentmanager.GatewayStarted{}, ErrFieldTooLong
	}
	amount, err := toWebpay(req.Amount, req.Currency)
	if err != nil {
		return paymentmanager.GatewayStarted{}, err
	}
	var res CreateResponse
	if err := w.call(methodPost, TransactionsPath, &CreateRequest{BuyOrder: req.BuyOrder, SessionId: req.SessionId, Amount: amount, ReturnUrl: req.ReturnURL}, &res); err != nil {
		return paymentmanager.GatewayStarted{}, err
	}
	return paymentmanager.GatewayStarted{Token: res.Token, RedirectURL: res.Url}, nil
}

// Commit confirma la transacción al volver el cliente (PUT).
func (w *Webpay) Commit(token string) (paymentmanager.GatewayResult, error) {
	return w.transaction(methodPut, token)
}

// Status consulta la transacción (GET); Transbank la guarda 7 días.
func (w *Webpay) Status(token string) (paymentmanager.GatewayResult, error) {
	return w.transaction(methodGet, token)
}

func (w *Webpay) transaction(method, token string) (paymentmanager.GatewayResult, error) {
	var tr TransactionResponse
	if err := w.call(method, TransactionsPath+"/"+token, nil, &tr); err != nil {
		return paymentmanager.GatewayResult{}, err
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
		return paymentmanager.GatewayResult{}, ErrUnknownStatus
	}
	return res, nil
}

// Refund devuelve amount (pesos). Webpay decide reversa (total, antes de 3 h)
// o anulación (diagrama §5).
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
