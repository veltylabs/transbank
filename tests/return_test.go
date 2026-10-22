//go:build !wasm

package tests

import (
	"testing"

	"webtyp.com/model"
	"webtyp.com/router/mock"

	paymentmanager "github.com/veltylabs/payment_manager"
	"github.com/veltylabs/transbank"
)

const page = "/reserva"

func mounted(t *testing.T, c *fakeCompleter) (*mock.Router, *fakeWebpay) {
	r := &mock.Router{}
	fw := newFake(t)
	fw.gateway().MountReturn(r, c, page)
	return r, fw
}

func invokeGet(r *mock.Router, query string) *mock.Context {
	ctx := &mock.Context{InMethod: "GET", InPath: transbank.ReturnPath + query}
	r.Invoke("GET", transbank.ReturnPath, ctx)
	return ctx
}

func invokePost(r *mock.Router, form string) *mock.Context {
	ctx := &mock.Context{InMethod: "POST", InPath: transbank.ReturnPath, InBody: []byte(form)}
	r.Invoke("POST", transbank.ReturnPath, ctx)
	return ctx
}

func assertRedirect(t *testing.T, ctx *mock.Context, want string) {
	t.Helper()
	if ctx.Status != 303 || ctx.GetHeader("Location") != want {
		t.Fatalf("status %d, Location %q; want 303 %q", ctx.Status, ctx.GetHeader("Location"), want)
	}
}

func paid() paymentmanager.Payment {
	return paymentmanager.Payment{Id: "pay-1", Status: paymentmanager.StatusPaid, ReturnTo: "booking-7"}
}

func TestMountReturn_RegistersPublicGetAndPost(t *testing.T) {
	r, _ := mounted(t, &fakeCompleter{})
	var get, post bool
	for _, info := range r.Routes() {
		if info.Path != transbank.ReturnPath {
			continue
		}
		if info.Access != model.AccessPublic {
			t.Errorf("%s %s must be public (Webpay brings no session)", info.Method, info.Path)
		}
		get = get || info.Method == "GET"
		post = post || info.Method == "POST"
	}
	if !get || !post {
		t.Fatalf("GET %v, POST %v", get, post)
	}
	if err := r.Verify(); err != nil {
		t.Fatal(err)
	}
}

func TestReturn_GetTokenRedirectsWithStatus(t *testing.T) {
	c := &fakeCompleter{result: paid()}
	r, fw := mounted(t, c)
	fw.respond(200, `{"status":"AUTHORIZED","response_code":0,"buy_order":"p1","amount":25000}`)
	ctx := invokeGet(r, "?token_ws=tok-1")
	if len(c.got) != 1 || c.got[0].PaymentId != "p1" || c.got[0].Result.Outcome != paymentmanager.OutcomeAuthorized {
		t.Fatalf("completer got %+v", c.got)
	}
	assertRedirect(t, ctx, page+"?payment=pay-1&status="+paymentmanager.StatusPaid+"#booking-7")
}

func TestReturn_PostTokenFromFormBody(t *testing.T) {
	c := &fakeCompleter{result: paid()}
	r, fw := mounted(t, c)
	fw.respond(200, `{"status":"AUTHORIZED","response_code":0,"buy_order":"p1","amount":25000}`)
	ctx := invokePost(r, "token_ws=tok-2")
	if len(c.got) != 1 || c.got[0].PaymentId != "p1" || c.got[0].Result.Outcome != paymentmanager.OutcomeAuthorized {
		t.Fatalf("completer got %+v", c.got)
	}
	assertRedirect(t, ctx, page+"?payment=pay-1&status="+paymentmanager.StatusPaid+"#booking-7")
}

func TestReturn_PostAbortFromFormBody(t *testing.T) {
	c := &fakeCompleter{result: paymentmanager.Payment{Id: "pay-1", Status: paymentmanager.StatusFailed}}
	r, _ := mounted(t, c)
	ctx := invokePost(r, "TBK_TOKEN=tok-a&TBK_ORDEN_COMPRA=b1&TBK_ID_SESION=s1")
	if len(c.got) != 1 || c.got[0].PaymentId != "b1" || c.got[0].Result.Outcome != paymentmanager.OutcomeCancelled {
		t.Fatalf("completer got %+v", c.got)
	}
	assertRedirect(t, ctx, page+"?payment=pay-1&status="+paymentmanager.StatusFailed)
}

func TestReturn_TokenAndAbortIsAbort(t *testing.T) {
	c := &fakeCompleter{result: paymentmanager.Payment{Id: "pay-1", Status: paymentmanager.StatusFailed}}
	r, _ := mounted(t, c)
	invokeGet(r, "?token_ws=tok-1&TBK_TOKEN=tok-1&TBK_ORDEN_COMPRA=b1")
	if len(c.got) != 1 || c.got[0].PaymentId != "b1" || c.got[0].Result.Outcome != paymentmanager.OutcomeCancelled {
		t.Fatalf("ambos tokens = el usuario hizo clic en \"volver al sitio\" después de un error; got %+v", c.got)
	}
}

func TestReturn_TimeoutPassesBuyOrder(t *testing.T) {
	c := &fakeCompleter{result: paymentmanager.Payment{Id: "pay-1", Status: paymentmanager.StatusFailed}}
	r, _ := mounted(t, c)
	invokePost(r, "TBK_ORDEN_COMPRA=b1&TBK_ID_SESION=s1")
	if len(c.got) != 1 || c.got[0].PaymentId != "b1" || c.got[0].Result.Outcome != paymentmanager.OutcomeExpired {
		t.Fatalf("completer got %+v", c.got)
	}
}

func TestReturn_NoParamsRedirectsError(t *testing.T) {
	c := &fakeCompleter{result: paid()}
	r, _ := mounted(t, c)
	ctx := invokeGet(r, "")
	if len(c.got) != 0 {
		t.Fatal("un retorno sin parámetros nunca llega a payment_manager")
	}
	assertRedirect(t, ctx, page+"?status="+transbank.StatusError)
}

func TestReturn_CommitErrorRedirectsError(t *testing.T) {
	c := &fakeCompleter{result: paid()}
	r, fw := mounted(t, c)
	fw.respond(500, ``)
	ctx := invokeGet(r, "?token_ws=tok-err")
	assertRedirect(t, ctx, page+"?status="+transbank.StatusError)
	if len(c.got) != 0 {
		t.Fatal("un commit fallido nunca llega al completer")
	}
}

func TestReturn_CompleterErrorRedirectsError(t *testing.T) {
	c := &fakeCompleter{err: paymentmanager.ErrNotFound}
	r, fw := mounted(t, c)
	fw.respond(200, `{"status":"AUTHORIZED","response_code":0,"buy_order":"p1","amount":25000}`)
	ctx := invokeGet(r, "?token_ws=forged")
	assertRedirect(t, ctx, page+"?status="+transbank.StatusError)
	if l := ctx.GetHeader("Location"); l != page+"?status=ERROR" {
		t.Fatalf("Location %q filtra detalles", l)
	}
}

func TestReturn_NoReturnToOmitsHash(t *testing.T) {
	c := &fakeCompleter{result: paymentmanager.Payment{Id: "pay-1", Status: paymentmanager.StatusPaid}}
	r, fw := mounted(t, c)
	fw.respond(200, `{"status":"AUTHORIZED","response_code":0,"buy_order":"p1","amount":25000}`)
	ctx := invokeGet(r, "?token_ws=tok-1")
	assertRedirect(t, ctx, page+"?payment=pay-1&status="+paymentmanager.StatusPaid)
}

// Un error que igual identifica el pago (p. ej. Webpay no respondió al
// confirmar) lleva al usuario a la pantalla del pago con su estado, no a ERROR.
func TestReturn_ErrorWithPaymentReportsItsStatus(t *testing.T) {
	c := &fakeCompleter{result: paymentmanager.Payment{Id: "pay-1", Status: paymentmanager.StatusPending}, err: paymentmanager.ErrNotFound}
	r, fw := mounted(t, c)
	fw.respond(200, `{"status":"AUTHORIZED","response_code":0,"buy_order":"p1","amount":25000}`)
	ctx := invokeGet(r, "?token_ws=tok-1")
	assertRedirect(t, ctx, page+"?payment=pay-1&status="+paymentmanager.StatusPending)
}
