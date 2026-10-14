package transbank

import (
	"github.com/veltylabs/payment_manager"
	"webtyp.com/router"
)

const ReturnPath = "/payments/webpay/return"

const (
	ParamToken    = "token_ws"
	ParamAbort    = "TBK_TOKEN"
	ParamBuyOrder = "TBK_ORDEN_COMPRA"
	ParamSession  = "TBK_ID_SESION"
	StatusError   = "ERROR"
)

type Completer interface {
	CompleteOnlinePayment(ret paymentmanager.OnlineReturn) (paymentmanager.Payment, error)
}

func (w *Webpay) MountReturn(r router.Router, c Completer, page string) {
	handler := func(ctx router.Context) {
		var token, abort, order string

		if ctx.Method() == "POST" {
			query := "?" + string(ctx.Body())
			token, _ = router.QueryParam(query, ParamToken)
			abort, _ = router.QueryParam(query, ParamAbort)
			order, _ = router.QueryParam(query, ParamBuyOrder)
		} else {
			token, _ = router.QueryParam(ctx.Path(), ParamToken)
			abort, _ = router.QueryParam(ctx.Path(), ParamAbort)
			order, _ = router.QueryParam(ctx.Path(), ParamBuyOrder)
		}

		var ret paymentmanager.OnlineReturn
		if token != "" && abort == "" {
			ret = paymentmanager.OnlineReturn{Token: token}
		} else if abort != "" {
			ret = paymentmanager.OnlineReturn{AbortToken: abort}
		} else if order != "" {
			ret = paymentmanager.OnlineReturn{BuyOrder: order}
		} else {
			ctx.SetHeader("Location", page+"?"+paymentmanager.QueryStatus+"="+StatusError)
			ctx.WriteStatus(303)
			return
		}

		payment, err := c.CompleteOnlinePayment(ret)
		if err != nil {
			if payment.Id != "" {
				dest := page + "?" + paymentmanager.QueryPayment + "=" + payment.Id + "&" + paymentmanager.QueryStatus + "=" + string(payment.Status)
				if payment.ReturnTo != "" {
					dest += "#" + payment.ReturnTo
				}
				ctx.SetHeader("Location", dest)
			} else {
				ctx.SetHeader("Location", page+"?"+paymentmanager.QueryStatus+"="+StatusError)
			}
		} else {
			dest := page + "?" + paymentmanager.QueryPayment + "=" + payment.Id + "&" + paymentmanager.QueryStatus + "=" + string(payment.Status)
			if payment.ReturnTo != "" {
				dest += "#" + payment.ReturnTo
			}
			ctx.SetHeader("Location", dest)
		}
		ctx.WriteStatus(303)
	}

	r.Get(ReturnPath, handler).Public()
	r.Post(ReturnPath, handler).Public()
}
