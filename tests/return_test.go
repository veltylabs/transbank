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

func mounted(c *fakeCompleter) *mock.Router {
	r := &mock.Router{}
	transbank.New(transbank.Config{}).MountReturn(r, c, page)
	return r
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
	r := mounted(&fakeCompleter{})
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
	ctx := invokeGet(mounted(c), "?token_ws=tok-1")
	if len(c.got) != 1 || c.got[0] != (paymentmanager.OnlineReturn{Token: "tok-1"}) {
		t.Fatalf("completer got %+v", c.got)
	}
	assertRedirect(t, ctx, page+"?payment=pay-1&status="+paymentmanager.StatusPaid+"#booking-7")
}

func TestReturn_PostTokenFromFormBody(t *testing.T) {
	c := &fakeCompleter{result: paid()}
	ctx := invokePost(mounted(c), "token_ws=tok-2")
	if len(c.got) != 1 || c.got[0].Token != "tok-2" {
		t.Fatalf("completer got %+v", c.got)
	}
	assertRedirect(t, ctx, page+"?payment=pay-1&status="+paymentmanager.StatusPaid+"#booking-7")
}

func TestReturn_PostAbortFromFormBody(t *testing.T) {
	c := &fakeCompleter{result: paymentmanager.Payment{Id: "pay-1", Status: paymentmanager.StatusFailed}}
	ctx := invokePost(mounted(c), "TBK_TOKEN=tok-a&TBK_ORDEN_COMPRA=b1&TBK_ID_SESION=s1")
	if len(c.got) != 1 || c.got[0] != (paymentmanager.OnlineReturn{AbortToken: "tok-a"}) {
		t.Fatalf("completer got %+v", c.got)
	}
	assertRedirect(t, ctx, page+"?payment=pay-1&status="+paymentmanager.StatusFailed)
}

func TestReturn_TokenAndAbortIsAbort(t *testing.T) {
	c := &fakeCompleter{result: paid()}
	invokeGet(mounted(c), "?token_ws=tok-1&TBK_TOKEN=tok-1")
	if len(c.got) != 1 || c.got[0] != (paymentmanager.OnlineReturn{AbortToken: "tok-1"}) {
		t.Fatalf("both tokens = the user clicked \"volver al sitio\" after an error; got %+v", c.got)
	}
}

func TestReturn_TimeoutPassesBuyOrder(t *testing.T) {
	c := &fakeCompleter{result: paymentmanager.Payment{Id: "pay-1", Status: paymentmanager.StatusFailed}}
	invokePost(mounted(c), "TBK_ORDEN_COMPRA=b1&TBK_ID_SESION=s1")
	if len(c.got) != 1 || c.got[0] != (paymentmanager.OnlineReturn{BuyOrder: "b1"}) {
		t.Fatalf("completer got %+v", c.got)
	}
}

func TestReturn_NoParamsRedirectsError(t *testing.T) {
	c := &fakeCompleter{result: paid()}
	ctx := invokeGet(mounted(c), "")
	if len(c.got) != 0 {
		t.Fatal("a return without parameters never reaches payment_manager")
	}
	assertRedirect(t, ctx, page+"?status="+transbank.StatusError)
}

func TestReturn_CompleterErrorRedirectsError(t *testing.T) {
	c := &fakeCompleter{err: paymentmanager.ErrNotFound}
	ctx := invokeGet(mounted(c), "?token_ws=forged")
	assertRedirect(t, ctx, page+"?status="+transbank.StatusError)
	if l := ctx.GetHeader("Location"); l != page+"?status=ERROR" {
		t.Fatalf("Location %q leaks details", l)
	}
}

func TestReturn_NoReturnToOmitsHash(t *testing.T) {
	c := &fakeCompleter{result: paymentmanager.Payment{Id: "pay-1", Status: paymentmanager.StatusPaid}}
	ctx := invokeGet(mounted(c), "?token_ws=tok-1")
	assertRedirect(t, ctx, page+"?payment=pay-1&status="+paymentmanager.StatusPaid)
}

// Un error que igual identifica el pago (p. ej. Webpay no respondió al
// confirmar) lleva al usuario a la pantalla del pago con su estado, no a ERROR.
func TestReturn_ErrorWithPaymentReportsItsStatus(t *testing.T) {
	c := &fakeCompleter{result: paymentmanager.Payment{Id: "pay-1", Status: paymentmanager.StatusPending}, err: paymentmanager.ErrNotFound}
	ctx := invokeGet(mounted(c), "?token_ws=tok-1")
	assertRedirect(t, ctx, page+"?payment=pay-1&status="+paymentmanager.StatusPending)
}
