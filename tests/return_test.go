package tests

import (
	"errors"
	"net/http"
	"strings"
	"testing"
	"github.com/veltylabs/payment_manager"
	"github.com/veltylabs/transbank"
	"webtyp.com/router"
	"webtyp.com/model"
)

type dummyRoute struct{}
func (d dummyRoute) Accepts(s model.Fielder) router.Route { return d }
func (d dummyRoute) AllowCORS() router.Route { return d }
func (d dummyRoute) Authenticated() router.Route { return d }
func (d dummyRoute) Public() router.Route { return d }
func (d dummyRoute) Describe(desc string) router.Route { return d }
func (d dummyRoute) Input(i model.Decodable) router.Route { return d }
func (d dummyRoute) Returns(code int, v model.Encodable) router.Route { return d }
func (d dummyRoute) Requires(res model.Resource, act model.Action) router.Route { return d }
func (d dummyRoute) RequiresMFA() router.Route { return d }
func (d dummyRoute) Handle(h router.HandlerFunc) router.Route { return d }

type dummyRouter struct {
	handler router.HandlerFunc
}

func (d *dummyRouter) Get(path string, h router.HandlerFunc) router.Route {
	d.handler = h
	return dummyRoute{}
}
func (d *dummyRouter) Post(path string, h router.HandlerFunc) router.Route {
	d.handler = h
	return dummyRoute{}
}
func (d *dummyRouter) Delete(path string, h router.HandlerFunc) router.Route { return dummyRoute{} }
func (d *dummyRouter) Put(path string, h router.HandlerFunc) router.Route { return dummyRoute{} }
func (d *dummyRouter) Patch(path string, h router.HandlerFunc) router.Route { return dummyRoute{} }
func (d *dummyRouter) Use(m ...router.Middleware) {}
func (d *dummyRouter) Mount(path string, r func(router.Router)) {}
func (d *dummyRouter) Handle(method, path string, h router.HandlerFunc) router.Route { return dummyRoute{} }
func (d *dummyRouter) ServeHTTP(w http.ResponseWriter, req *http.Request) {}
func (d *dummyRouter) Options(path string, h router.HandlerFunc) router.Route { return dummyRoute{} }
func (d *dummyRouter) Group(f func(router.Router)) router.Router { return d }
func (d *dummyRouter) PublicAsset(path string, h router.HandlerFunc) {}
func (d *dummyRouter) PublicDir(path string, dir string) {}
func (d *dummyRouter) PublicFile(path string, file string) {}
func (d *dummyRouter) Routes() []router.RouteInfo { return nil }
func (d *dummyRouter) Socket(path string, h router.SocketFunc) router.Route { return dummyRoute{} }
func (d *dummyRouter) NotFound(h router.HandlerFunc) {}
func (d *dummyRouter) RouteInfo(method, path string) router.RouteInfo { return router.RouteInfo{} }
func (d *dummyRouter) Stream(path string, h router.StreamFunc) router.Route { return dummyRoute{} }

type dummyContext struct {
	method  string
	path    string
	body    []byte
	headers map[string]string
	status  int
}

func (d *dummyContext) Method() string { return d.method }
func (d *dummyContext) Path() string { return d.path }
func (d *dummyContext) Body() []byte { return d.body }
func (d *dummyContext) SetHeader(k, v string) { d.headers[k] = v }
func (d *dummyContext) WriteStatus(code int) { d.status = code }
func (d *dummyContext) Write(b []byte) (int, error) { return len(b), nil }

func (d *dummyContext) RequestHeader(k string) string { return "" }
func (d *dummyContext) RemoteIP() string { return "" }
func (d *dummyContext) URL() string { return "" }
func (d *dummyContext) Cookie(name string) (router.Cookie, bool) { return router.Cookie{}, false }
func (d *dummyContext) SetCookie(cookie router.Cookie) {}
func (d *dummyContext) Next() {}
func (d *dummyContext) Abort() {}
func (d *dummyContext) Redirect(code int, url string) {}
func (d *dummyContext) QueryParam(key string) string {
	if d.method == "POST" {
		query := string(d.body)
		if strings.Contains(query, key+"=") {
			return strings.Split(strings.Split(query, key+"=")[1], "&")[0]
		}
	} else {
		if strings.Contains(d.path, key+"=") {
			return strings.Split(strings.Split(d.path, key+"=")[1], "&")[0]
		}
	}
	return ""
}
func (d *dummyContext) URLParam(key string) string { return "" }
func (d *dummyContext) Decode(v model.Decodable) error { return nil }
func (d *dummyContext) Encode(v model.Encodable) error { return nil }
func (d *dummyContext) WriteJSON(code int, v model.Encodable) {}
func (d *dummyContext) GetHeader(k string) string { return "" }
func (d *dummyContext) Param(key string) string { return "" }
func (d *dummyContext) DeleteCookie(cookie router.Cookie) {}
func (d *dummyContext) SetUserID(id string) {}
func (d *dummyContext) SetValue(k string, v string) {}
func (d *dummyContext) Value(k string) string { return "" }
func (d *dummyContext) UserID() string { return "" }
func (d *dummyContext) Assurance() model.Assurance { return 0 }
func (d *dummyContext) SetAssurance(a model.Assurance) {}

func TestMountReturn_Token(t *testing.T) {
	wp := transbank.New(transbank.Config{BaseURL: "http://localhost"})
	comp := &fakeCompleter{
		payment: paymentmanager.Payment{Id: "P-1", Status: "SUCCESS", ReturnTo: "ok"},
	}

	r := &dummyRouter{}
	wp.MountReturn(r, comp, "http://checkout")

	ctx := &dummyContext{
		method:  "GET",
		path:    transbank.ReturnPath + "?token_ws=T-123",
		headers: make(map[string]string),
	}
	r.handler(ctx)

	if ctx.status != 303 {
		t.Fatalf("expected 303, got %d", ctx.status)
	}
	loc := ctx.headers["Location"]
	if loc != "http://checkout?payment=P-1&status=SUCCESS#ok" {
		t.Fatalf("expected http://checkout?payment=P-1&status=SUCCESS#ok, got %s", loc)
	}
	if comp.lastReturn.Token != "T-123" {
		t.Fatalf("expected token T-123, got %v", comp.lastReturn)
	}
}

func TestMountReturn_Abort(t *testing.T) {
	wp := transbank.New(transbank.Config{BaseURL: "http://localhost"})
	comp := &fakeCompleter{
		payment: paymentmanager.Payment{Id: "P-1", Status: "FAILED"},
	}

	r := &dummyRouter{}
	wp.MountReturn(r, comp, "http://checkout")

	ctx := &dummyContext{
		method:  "POST",
		path:    transbank.ReturnPath,
		body:    []byte("TBK_TOKEN=T-123"),
		headers: make(map[string]string),
	}
	r.handler(ctx)

	if ctx.status != 303 {
		t.Fatalf("expected 303, got %d", ctx.status)
	}
	loc := ctx.headers["Location"]
	if loc != "http://checkout?payment=P-1&status=FAILED" {
		t.Fatalf("expected http://checkout?payment=P-1&status=FAILED, got %s", loc)
	}
	if comp.lastReturn.AbortToken != "T-123" {
		t.Fatalf("expected abort token T-123, got %v", comp.lastReturn)
	}
}

func TestMountReturn_BuyOrder(t *testing.T) {
	wp := transbank.New(transbank.Config{BaseURL: "http://localhost"})
	comp := &fakeCompleter{
		payment: paymentmanager.Payment{Id: "P-1", Status: "FAILED"},
	}

	r := &dummyRouter{}
	wp.MountReturn(r, comp, "http://checkout")

	ctx := &dummyContext{
		method:  "GET",
		path:    transbank.ReturnPath + "?TBK_ORDEN_COMPRA=O-1",
		headers: make(map[string]string),
	}
	r.handler(ctx)

	if ctx.status != 303 {
		t.Fatalf("expected 303, got %d", ctx.status)
	}
	if comp.lastReturn.BuyOrder != "O-1" {
		t.Fatalf("expected buy order O-1, got %v", comp.lastReturn)
	}
}

func TestMountReturn_Empty(t *testing.T) {
	wp := transbank.New(transbank.Config{BaseURL: "http://localhost"})
	comp := &fakeCompleter{}

	r := &dummyRouter{}
	wp.MountReturn(r, comp, "http://checkout")

	ctx := &dummyContext{
		method:  "GET",
		path:    transbank.ReturnPath,
		headers: make(map[string]string),
	}
	r.handler(ctx)

	if ctx.status != 303 {
		t.Fatalf("expected 303, got %d", ctx.status)
	}
	loc := ctx.headers["Location"]
	if loc != "http://checkout?status=ERROR" {
		t.Fatalf("expected http://checkout?status=ERROR, got %s", loc)
	}
}

func TestMountReturn_CompleteErrorWithPayment(t *testing.T) {
	wp := transbank.New(transbank.Config{BaseURL: "http://localhost"})
	comp := &fakeCompleter{
		payment: paymentmanager.Payment{Id: "P-1", Status: "FAILED", ReturnTo: "err"},
		err:     errors.New("test error"),
	}

	r := &dummyRouter{}
	wp.MountReturn(r, comp, "http://checkout")

	ctx := &dummyContext{
		method:  "GET",
		path:    transbank.ReturnPath + "?token_ws=T-123",
		headers: make(map[string]string),
	}
	r.handler(ctx)

	if ctx.status != 303 {
		t.Fatalf("expected 303, got %d", ctx.status)
	}
	loc := ctx.headers["Location"]
	if loc != "http://checkout?payment=P-1&status=FAILED#err" {
		t.Fatalf("expected payment error redirect, got %s", loc)
	}
}

func TestMountReturn_CompleteErrorNoPayment(t *testing.T) {
	wp := transbank.New(transbank.Config{BaseURL: "http://localhost"})
	comp := &fakeCompleter{
		err: errors.New("test error"),
	}

	r := &dummyRouter{}
	wp.MountReturn(r, comp, "http://checkout")

	ctx := &dummyContext{
		method:  "GET",
		path:    transbank.ReturnPath + "?token_ws=T-123",
		headers: make(map[string]string),
	}
	r.handler(ctx)

	if ctx.status != 303 {
		t.Fatalf("expected 303, got %d", ctx.status)
	}
	loc := ctx.headers["Location"]
	if loc != "http://checkout?status=ERROR" {
		t.Fatalf("expected error redirect, got %s", loc)
	}
}
