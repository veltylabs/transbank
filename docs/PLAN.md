---
PLAN: "fix!: Webpay return passes only identification and Confirm asks Transbank (payment_manager Gateway.Confirm, router v0.5.0 RawQuery)"
EXECUTOR: jules
REVIEWER: none
---

> Read AGENTS.md at the repo root before any change.
> **Finish by yourself:** when the delivery check passes, commit, push and open the pull request. Never ask whether to
> proceed or finalize — nobody answers, and a session that waits ends without your work. Commit after every stage.
> Never run `git reset`, `git restore`, `git checkout -- <file>` or `git stash` on your own work.
> **Publishing:** you publish only by OPENING THE PULL REQUEST (a `git push` from your sandbox does not reach GitHub). Submit after Stage 2 and keep committing.
> **Never ask a yes/no question about progress**; ask only when a signature in this plan is wrong, and name the symbol.
> **Before you submit:** `git status` must be clean and `git diff --stat main` must list the files of the stages. A pull request that carries only go.mod, go.sum or docs/PLAN.md is a failed delivery: `git add -A`, commit and submit again.
> No python, sed or regex over the source: edit file by file.

# Plan: `transbank` — the Webpay return never trusts the request

## Why (read once)

`payment_manager` (main, next tag after v0.4.47) changed two contracts this adapter implements and calls: `Gateway` gained `Confirm(token string) (GatewayResult, error)`, and `CompleteOnlinePayment` now takes `OnlineReturn{PaymentId, Token}`, **no outcome** (`OnlineResult` is gone): the module asks the gateway itself. Today `MountReturn` builds a «cancelled» or «expired» outcome from the request alone, so anyone who knows a payment id can cancel or expire it with `GET /payments/webpay/return?TBK_TOKEN=x&TBK_ORDEN_COMPRA=<id>`. Also `router` v0.5.0 made `Context.Path()` the path ONLY: the query is `ctx.RawQuery()`; reading it from `Path()` loses every parameter.

## Tools

```bash
go install webtyp.com/devflow/cmd/gotest@latest
go install webtyp.com/lang/cmd/langc@latest
go get github.com/veltylabs/payment_manager@main webtyp.com/router@v0.5.0
go mod tidy
```
(`@main` because the tag with `Confirm` and `OnlineReturn` is not cut yet; `gopush` rewrites it to the tag.) No `replace`. If `go doc github.com/veltylabs/payment_manager OnlineReturn` fails, stop and say so in the PR.

## APIs used

| symbol | exact signature | where |
|---|---|---|
| `paymentmanager.OnlineReturn` | `struct{ PaymentId string; Token string }` | payment_manager `ports.go` |
| `paymentmanager.Gateway` | `Provider() string; Start(GatewayStart) (GatewayStarted, error); Status(token string) (GatewayResult, error); Confirm(token string) (GatewayResult, error); Refund(token string, amount int64) (GatewayRefund, error)` | `ports.go` |
| `router.Context.RawQuery` | `RawQuery() string` ("a=1&b=2", no "?") | router v0.5.0 |
| `router.QueryParam` | `func QueryParam(path, key string) (value string, ok bool)` (reads the query of a string that has a `?`) | router |
| `(*Webpay).commit` | `func (w *Webpay) commit(token string) (paymentmanager.GatewayResult, string, error)` (PUT) | `webpay.go:57` |
| `(*Webpay).Status` | `func (w *Webpay) Status(token string) (paymentmanager.GatewayResult, error)` (GET) | `webpay.go:63` |
If a signature differs, stop and name the symbol in the PR; never adapt by guessing.

## Stage 1 — `webpay.go`: `Confirm` (append after `Status`)

```go
// Confirm finalizes the transaction (commit, PUT) and reports its real state. A transaction that cannot be committed (the buyer
// cancelled, the form expired, it was already committed) is reported as Transbank sees it (Status, GET). Idempotent.
//
// Example: res, err := wp.Confirm(token)
func (w *Webpay) Confirm(token string) (paymentmanager.GatewayResult, error) {
	res, _, err := w.commit(token)
	if err == nil {
		return res, nil
	}
	return w.Status(token)
}
```
If `Status` also fails its error is returned (the payment stays pending; reconciliation settles it later).

## Stage 2 — `return.go`: identification only

Replace the `Completer` interface and the whole `handler` (keep `redirect`, the constants and the two `r.Get`/`r.Post` lines):

```go
// Completer es lo que la ruta necesita de payment_manager (lo satisface *paymentmanager.Module).
type Completer interface {
	CompleteOnlinePayment(ret paymentmanager.OnlineReturn) (paymentmanager.Payment, error)
}
```

```go
	handler := func(ctx router.Context) {
		source := "?" + ctx.RawQuery() // GET → query
		if ctx.Method() == methodPost {
			source = "?" + string(ctx.Body()) // POST → formulario
		}
		token, _ := router.QueryParam(source, ParamToken)
		abort, _ := router.QueryParam(source, ParamAbort)
		order, _ := router.QueryParam(source, ParamBuyOrder)
		if abort != "" { // el cliente volvió sin pagar: el token que trae es TBK_TOKEN
			token = abort
		}
		if token == "" && order == "" {
			redirect(ctx, page+"?"+paymentmanager.QueryStatus+"="+StatusError)
			return
		}
		// La vuelta solo IDENTIFICA el pago: lo que pasó lo pregunta payment_manager a Webpay (Confirm).
		p, _ := c.CompleteOnlinePayment(paymentmanager.OnlineReturn{PaymentId: order, Token: token})
		if p.Id == "" {
			redirect(ctx, page+"?"+paymentmanager.QueryStatus+"="+StatusError)
			return
		}
		dest := page + "?" + paymentmanager.QueryPayment + "=" + p.Id + "&" + paymentmanager.QueryStatus + "=" + p.Status
		if p.ReturnTo != "" {
			dest += "#" + p.ReturnTo
		}
		redirect(ctx, dest)
	}
```
Check: `grep -n "OutcomeCancelled\|OutcomeExpired\|w.commit(\|OnlineResult\|ctx.Path()" return.go` → empty. `w.commit` stays used by `Confirm`.

## Stage 3 — tests (`tests/`, rows in `docs/TESTS.md`)

**3.1** `tests/setup_test.go`: change `fakeCompleter` to

```go
type fakeCompleter struct {
	got    []paymentmanager.OnlineReturn
	result paymentmanager.Payment
	err    error
}

func (c *fakeCompleter) CompleteOnlinePayment(r paymentmanager.OnlineReturn) (paymentmanager.Payment, error) {
	c.got = append(c.got, r)
	return c.result, c.err
}
```

**3.2** `tests/fake_webpay_test.go`: answer per method when the test asks for it. Add the field `byMethod map[string]response` to `fakeWebpay`, the type `type response struct { status int; body string }`, and in the handler, right after the `f.requests = append(...)` line:

```go
		status, body := f.status, f.body
		if resp, ok := f.byMethod[r.Method]; ok {
			status, body = resp.status, resp.body
		}
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
```
(replacing the two lines that wrote `f.status` and `f.body`), and the method

```go
func (f *fakeWebpay) respondTo(method string, status int, body string) {
	if f.byMethod == nil {
		f.byMethod = map[string]response{}
	}
	f.byMethod[method] = response{status, body}
}
```

**3.3** `tests/return_test.go`: DELETE `TestReturn_GetTokenRedirectsWithStatus`, `TestReturn_PostTokenFromFormBody`, `TestReturn_PostAbortFromFormBody`, `TestReturn_TokenAndAbortIsAbort`, `TestReturn_TimeoutPassesBuyOrder`, `TestReturn_CommitErrorRedirectsError`, `TestReturn_CompleterErrorRedirectsError`, `TestReturn_NoReturnToOmitsHash`, `TestReturn_ErrorWithPaymentReportsItsStatus` and add (same file, helpers `mounted`, `invokeGet`, `invokePost`, `assertRedirect`, `paid` stay):

```go
func TestReturn_PassesOnlyIdentification(t *testing.T) {
	c := &fakeCompleter{result: paymentmanager.Payment{Id: "pay-1", Status: paymentmanager.StatusFailed}}
	r, fw := mounted(t, c)
	invokeGet(r, "?TBK_TOKEN=t&TBK_ORDEN_COMPRA=p1")
	if len(c.got) != 1 || c.got[0].PaymentId != "p1" || c.got[0].Token != "t" {
		t.Fatalf("completer got %+v", c.got)
	}
	if len(fw.requests) != 0 {
		t.Fatal("the return route must not call Webpay itself: payment_manager does it through Confirm")
	}
}

func TestReturn_TokenAlonePassesTheToken(t *testing.T) {
	c := &fakeCompleter{result: paid()}
	r, _ := mounted(t, c)
	ctx := invokeGet(r, "?token_ws=tok-1")
	if len(c.got) != 1 || c.got[0].Token != "tok-1" || c.got[0].PaymentId != "" {
		t.Fatalf("completer got %+v", c.got)
	}
	assertRedirect(t, ctx, page+"?payment=pay-1&status="+paymentmanager.StatusPaid+"#booking-7")
}

func TestReturn_PostReadsTheFormBody(t *testing.T) {
	c := &fakeCompleter{result: paid()}
	r, _ := mounted(t, c)
	invokePost(r, "TBK_TOKEN=tok-a&TBK_ORDEN_COMPRA=b1&TBK_ID_SESION=s1")
	if len(c.got) != 1 || c.got[0].PaymentId != "b1" || c.got[0].Token != "tok-a" {
		t.Fatalf("completer got %+v", c.got)
	}
}

func TestReturn_TimeoutPassesBuyOrder(t *testing.T) {
	c := &fakeCompleter{result: paid()}
	r, _ := mounted(t, c)
	invokePost(r, "TBK_ORDEN_COMPRA=b1&TBK_ID_SESION=s1")
	if len(c.got) != 1 || c.got[0].PaymentId != "b1" || c.got[0].Token != "" {
		t.Fatalf("completer got %+v", c.got)
	}
}

func TestReturn_TokenWsIsReplacedByAbortToken(t *testing.T) {
	c := &fakeCompleter{result: paid()}
	r, _ := mounted(t, c)
	invokeGet(r, "?token_ws=a&TBK_TOKEN=b&TBK_ORDEN_COMPRA=p1")
	if len(c.got) != 1 || c.got[0].Token != "b" {
		t.Fatalf("a return that carries TBK_TOKEN is an abort: its token wins, got %+v", c.got)
	}
}

func TestReturn_CompleterErrorWithoutPaymentRedirectsError(t *testing.T) {
	c := &fakeCompleter{err: paymentmanager.ErrNotFound}
	r, _ := mounted(t, c)
	ctx := invokeGet(r, "?token_ws=forged")
	assertRedirect(t, ctx, page+"?status="+transbank.StatusError)
}

func TestReturn_ErrorWithPaymentReportsItsStatus(t *testing.T) {
	c := &fakeCompleter{result: paymentmanager.Payment{Id: "pay-1", Status: paymentmanager.StatusPending}, err: paymentmanager.ErrGatewayFailed}
	r, _ := mounted(t, c)
	ctx := invokeGet(r, "?token_ws=tok-1")
	assertRedirect(t, ctx, page+"?payment=pay-1&status="+paymentmanager.StatusPending)
}
```

**3.4** `tests/webpay_test.go` (append; imports `paymentmanager`, `testing` are already there; if `webpay_test.go` lacks one of them add it):

```go
func TestConfirm_AuthorizedOnCommit(t *testing.T) {
	fw := newFake(t)
	fw.respond(200, `{"status":"AUTHORIZED","response_code":0,"buy_order":"p1","amount":25000}`)
	res, err := fw.gateway().Confirm("tok-1")
	if err != nil || res.Outcome != paymentmanager.OutcomeAuthorized || res.Amount != 25000 {
		t.Fatalf("Confirm = %+v, %v", res, err)
	}
}

func TestConfirm_CommitFailsThenStatusReportsPending(t *testing.T) {
	fw := newFake(t)
	fw.respondTo("PUT", 422, `{"error_message":"already committed or not committable"}`)
	fw.respondTo("GET", 200, `{"status":"INITIALIZED","buy_order":"p1","amount":25000}`)
	res, err := fw.gateway().Confirm("tok-1")
	if err != nil || res.Outcome != paymentmanager.OutcomePending {
		t.Fatalf("Confirm = %+v, %v; want pending", res, err)
	}
}
```
and in `tests/contract_test.go` (or where `var _ paymentmanager.Gateway = (*transbank.Webpay)(nil)` lives) keep that assertion: it now includes `Confirm`. Each new or renamed test has its row in `docs/TESTS.md`; remove the rows of the deleted ones.

## Stage 4 — documentation (Spanish, veltylabs)

`README.md` and `docs/ARCHITECTURE.md`: the return route only identifies the payment; `payment_manager` asks Transbank through `Confirm`; why (a forged return cannot change a payment). Update every `docs/diagrams/*.md` that shows the return deciding an outcome.

## Delivery check

`gotest` passes on every destination; every grep of this plan is empty; the pull request is open. A tree that does not compile or whose tests did not run is a failed delivery.
