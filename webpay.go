package transbank

import (
	"github.com/veltylabs/payment_manager"
	"webtyp.com/fmt"
)

type Webpay struct {
	cfg Config
}

var _ paymentmanager.Gateway = (*Webpay)(nil)

func (w *Webpay) Start(req paymentmanager.GatewayStart) (paymentmanager.GatewayStarted, error) {
	if len(req.BuyOrder) > 26 || len(req.SessionId) > 61 || len(req.ReturnURL) > 256 {
		return paymentmanager.GatewayStarted{}, fmt.Err("transbank: field exceeds Webpay length")
	}

	amount, err := encodeAmount(req.Amount, req.Currency)
	if err != nil {
		return paymentmanager.GatewayStarted{}, err
	}

	creq := &CreateRequest{
		BuyOrder:  req.BuyOrder,
		SessionId: req.SessionId,
		Amount:    amount,
		ReturnUrl: req.ReturnURL,
	}

	var res CreateResponse
	err = w.call("POST", "/rswebpaytransaction/api/webpay/v1.2/transactions", creq, &res)
	if err != nil {
		return paymentmanager.GatewayStarted{}, err
	}

	return paymentmanager.GatewayStarted{
		Token:       res.Token,
		RedirectURL: res.Url + "?token_ws=" + res.Token,
	}, nil
}

func (w *Webpay) Commit(token string) (paymentmanager.GatewayResult, error) {
	var tr TransactionResponse
	err := w.call("PUT", "/rswebpaytransaction/api/webpay/v1.2/transactions/"+token, nil, &tr)
	if err != nil {
		return paymentmanager.GatewayResult{}, err
	}
	return w.translateStatus(&tr, token)
}

func (w *Webpay) Status(token string) (paymentmanager.GatewayResult, error) {
	var tr TransactionResponse
	err := w.call("GET", "/rswebpaytransaction/api/webpay/v1.2/transactions/"+token, nil, &tr)
	if err != nil {
		return paymentmanager.GatewayResult{}, err
	}
	return w.translateStatus(&tr, token)
}

func (w *Webpay) Refund(token string, amount int64) (paymentmanager.GatewayRefund, error) {
	req := &RefundRequest{
		Amount: float64(amount),
	}
	var res RefundResponse
	err := w.call("POST", "/rswebpaytransaction/api/webpay/v1.2/transactions/"+token+"/refunds", req, &res)
	if err != nil {
		return paymentmanager.GatewayRefund{}, err
	}

	switch res.Type {
	case "REVERSED":
		return paymentmanager.GatewayRefund{
			Kind:   paymentmanager.RefundKindReversed,
			Amount: amount,
		}, nil
	case "NULLIFIED":
		if res.ResponseCode != 0 {
			return paymentmanager.GatewayRefund{}, fmt.Err("transbank: refund rejected")
		}
		return paymentmanager.GatewayRefund{
			Kind:              paymentmanager.RefundKindNullified,
			Amount:            int64(res.NullifiedAmount),
			AuthorizationCode: res.AuthorizationCode,
		}, nil
	default:
		return paymentmanager.GatewayRefund{}, fmt.Err("transbank: unknown refund type")
	}
}

func (w *Webpay) translateStatus(tr *TransactionResponse, token string) (paymentmanager.GatewayResult, error) {
	res := paymentmanager.GatewayResult{
		Amount:            int64(tr.Amount),
		AuthorizationCode: tr.AuthorizationCode,
		PaymentTypeCode:   tr.PaymentTypeCode,
		Installments:      tr.InstallmentsNumber,
	}

	if len(tr.CardDetail.CardNumber) >= 4 {
		res.CardLast4 = tr.CardDetail.CardNumber[len(tr.CardDetail.CardNumber)-4:]
	}

	if tr.ResponseCode != 0 {
		res.Outcome = paymentmanager.OutcomeRejected
		return res, nil
	}

	switch tr.Status {
	case "AUTHORIZED":
		res.Outcome = paymentmanager.OutcomeAuthorized
	case "FAILED":
		res.Outcome = paymentmanager.OutcomeRejected
	default:
		return paymentmanager.GatewayResult{}, fmt.Err("transbank: unknown status")
	}

	return res, nil
}

func mapStatus(status string) string {
	switch status {
	case "INITIALIZED":
		return "FAILED"
	default:
		return "FAILED"
	}
}
