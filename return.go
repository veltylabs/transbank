package transbank

import (
	"webtyp.com/router"

	paymentmanager "github.com/veltylabs/payment_manager"
)

// ReturnPath es la ruta a la que Webpay devuelve el navegador.
const ReturnPath = "/payments/webpay/return"

// Parámetros con que Webpay vuelve y el estado que informa un retorno inválido.
const (
	ParamToken    = "token_ws"
	ParamAbort    = "TBK_TOKEN"
	ParamBuyOrder = "TBK_ORDEN_COMPRA"
	ParamSession  = "TBK_ID_SESION"
	StatusError   = "ERROR"
)

// Completer es lo que la ruta necesita de payment_manager (lo satisface
// *paymentmanager.Module).
type Completer interface {
	CompleteOnlinePayment(ret paymentmanager.OnlineReturn) (paymentmanager.Payment, error)
}

// MountReturn registra GET y POST en ReturnPath, públicos: Webpay no trae
// sesión y el token solo sirve con la llave secreta. page es la página de la
// app a la que se redirige con ?payment=<id>&status=<estado>#<ReturnTo>.
func (w *Webpay) MountReturn(r router.Router, c Completer, page string) {
	handler := func(ctx router.Context) {
		source := ctx.Path() // R1: GET → query
		if ctx.Method() == methodPost {
			source = "?" + string(ctx.Body()) // R1: POST → formulario
		}
		token, _ := router.QueryParam(source, ParamToken)
		abort, _ := router.QueryParam(source, ParamAbort)
		order, _ := router.QueryParam(source, ParamBuyOrder)

		var ret paymentmanager.OnlineReturn
		switch { // R2
		case abort != "":
			ret.AbortToken = abort // abandono, o error con "volver al sitio" (trae ambos)
		case token != "":
			ret.Token = token
		case order != "":
			ret.BuyOrder = order // venció el formulario
		default:
			redirect(ctx, page+"?"+paymentmanager.QueryStatus+"="+StatusError)
			return
		}

		p, _ := c.CompleteOnlinePayment(ret)
		if p.Id == "" { // R3 no
			redirect(ctx, page+"?"+paymentmanager.QueryStatus+"="+StatusError)
			return
		}
		dest := page + "?" + paymentmanager.QueryPayment + "=" + p.Id + "&" + paymentmanager.QueryStatus + "=" + p.Status
		if p.ReturnTo != "" {
			dest += "#" + p.ReturnTo
		}
		redirect(ctx, dest)
	}
	r.Get(ReturnPath, handler).Public()
	r.Post(ReturnPath, handler).Public()
}

func redirect(ctx router.Context, location string) {
	ctx.SetHeader("Location", location)
	ctx.WriteStatus(303)
}
