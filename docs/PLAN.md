---
PLAN: "feat!: adapt to payment_manager v0.3.0 (commit inside the return route, OnlineResult, Provider)"
TAG: v0.2.0
EXECUTOR: jules
REVIEWER: none
---

> Read AGENTS.md at the repo root before any change.

<!-- Parked. Dispatch only after payment_manager v0.3.0 (its docs/PLAN_GATEWAY.md) is published; then rename this
file to docs/PLAN.md, write the real tag in "Pinned versions" and check the "APIs used" table against it. -->

# Plan — Webpay adapter for the processor-agnostic `Gateway`

## Why

`payment_manager` v0.3.0 removed `Gateway.Commit` and `OnlineReturn`: each adapter verifies what its processor
sends and calls `CompleteOnlinePayment(OnlineResult)` with a verified result. For Webpay, "verify" means calling
its commit (PUT) inside the return route. The adapter also owns its return URL now (`Deps.ReturnURL` is gone).

Language: identifiers and error strings in English; **comments and docs in Spanish**.

## Delivery rules

- **Commit after every stage.** Never run `git reset`, `git restore`, `git checkout -- <file>` or `git stash`.
- Never pause to ask permission or to report progress. When the Delivery check passes: push and open the PR. Stuck
  for 30 minutes: open the PR anyway, listing what fails.
- A PR that does not compile, or whose tests could not run, is a failed delivery.
- Never edit `model_orm.go` by hand.
- The code blocks below are the implementation: copy them.

## Tools and pinned versions

```bash
go install webtyp.com/devflow/cmd/gotest@latest
export GOPRIVATE='webtyp.com/*,github.com/veltylabs/*' GOFLAGS=-mod=mod
go get github.com/veltylabs/payment_manager@v0.3.0 && go mod tidy   # ← tag written by the dispatcher
```

## APIs used (`github.com/veltylabs/payment_manager` v0.3.0)

| Symbol | Shape |
|---|---|
| `Gateway` | `interface{ Provider() string; Start(GatewayStart) (GatewayStarted, error); Status(token string) (GatewayResult, error); Refund(token string, amount int64) (GatewayRefund, error) }` |
| `GatewayStart` | `struct{ PaymentId string; Amount int64; Currency string; Description string; Payer Payer }` |
| `GatewayStarted` | `struct{ Token, RedirectURL, TokenField string }` |
| `OnlineResult` | `struct{ PaymentId string; Result GatewayResult }` |
| `GatewayResult` | unchanged from v0.2.0 (`Outcome, Amount, AuthorizationCode, PaymentTypeCode, Installments, CardLast4`) |
| `OutcomePending/Authorized/Rejected/Cancelled/Expired` | string constants |
| `(*Module).CompleteOnlinePayment` | `func (m *Module) CompleteOnlinePayment(r OnlineResult) (Payment, error)` |

If a shape differs in the published tag, stop and name it in the PR; never adapt by guessing.

## Stage 1 — Config owns the return URL (`config.go`)

- `Config` gains `ReturnURL string // URL absoluta y pública de ReturnPath (https://<sitio>/payments/webpay/return)`.
- `Open(raw string)` → `Open(raw, returnURL string)`: same parsing; passes `ReturnURL: returnURL` to `New`.
- `New`: `cfg.ReturnURL == ""` is allowed (tests that never call `Start`); `Start` with an empty `ReturnURL` returns
  the new error `ErrNoReturnURL = fmt.Err("transbank: Config.ReturnURL is required to start a payment")` (add it to
  `errors.go` next to the others).

**Commit.**

## Stage 2 — Gateway (`webpay.go`)

```go
// ProviderName es el nombre que payment_manager guarda en cada pago de este adaptador.
const ProviderName = "transbank"

// TokenField es el campo con que el navegador envía el token a Webpay (POST).
const TokenField = "token_ws"

func (w *Webpay) Provider() string { return ProviderName }

// Start crea la transacción; el navegador hace POST a RedirectURL con TokenField.
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
```

- `Commit` becomes private: `func (w *Webpay) commit(token string) (paymentmanager.GatewayResult, string, error)`
  returning the result **and the buy order**; `transaction` returns `(paymentmanager.GatewayResult, string, error)`
  with `tr.BuyOrder` as the second value; `Status` discards it (`res, _, err := w.transaction(methodGet, token)`).
- `MaxSessionID` stays (the session id equals the payment id, already bounded by `MaxBuyOrder`).

**Commit.**

## Stage 3 — Return route (`return.go`)

```go
// Completer es lo que la ruta necesita de payment_manager (lo satisface *paymentmanager.Module).
type Completer interface {
	CompleteOnlinePayment(r paymentmanager.OnlineResult) (paymentmanager.Payment, error)
}
```

Inside the handler, replace the block that builds `ret paymentmanager.OnlineReturn` and calls the completer with:

```go
		var r paymentmanager.OnlineResult
		switch { // R2
		case abort != "" && order != "":
			r = paymentmanager.OnlineResult{PaymentId: order, Result: paymentmanager.GatewayResult{Outcome: paymentmanager.OutcomeCancelled}}
		case token != "":
			res, buyOrder, err := w.commit(token) // verificar = confirmar con Webpay
			if err != nil {
				redirect(ctx, page+"?"+paymentmanager.QueryStatus+"="+StatusError)
				return
			}
			r = paymentmanager.OnlineResult{PaymentId: buyOrder, Result: res}
		case order != "":
			r = paymentmanager.OnlineResult{PaymentId: order, Result: paymentmanager.GatewayResult{Outcome: paymentmanager.OutcomeExpired}} // venció el formulario
		default:
			redirect(ctx, page+"?"+paymentmanager.QueryStatus+"="+StatusError)
			return
		}
		p, _ := c.CompleteOnlinePayment(r)
```

(the rest — redirect with `payment`, `status`, `#ReturnTo` — is unchanged).

| Case | Webpay sends | Completer receives |
|---|---|---|
| R1 | `token_ws`, commit AUTHORIZED code 0 | `{PaymentId: <buy_order>, Outcome: authorized, Amount…}` |
| R2 | `token_ws`, commit fails (500) | nothing; redirect `?status=ERROR` |
| R3 | `TBK_TOKEN` + `TBK_ORDEN_COMPRA=p1` | `{PaymentId: p1, Outcome: cancelled}`; no call to Webpay |
| R4 | only `TBK_ORDEN_COMPRA=p1` | `{PaymentId: p1, Outcome: expired}` |
| R5 | nothing | redirect `?status=ERROR` |
| R6 | `TBK_TOKEN` without order | redirect `?status=ERROR` |

**Commit.**

## Stage 4 — Tests and docs

- `tests/setup_test.go`: `fakeCompleter.got []paymentmanager.OnlineResult`, method signature as in Stage 3.
- `tests/return_test.go`: `mounted` uses the fake Webpay (`newFake(t).gateway()`) so the token path can commit; the
  token cases respond with a `TransactionResponse` JSON that has `"status":"AUTHORIZED","response_code":0,"buy_order":"p1","amount":25000`.
  Cases R1–R6. `TestReturn_TokenAndAbortIsAbort` keeps its meaning (R3 wins).
- `tests/webpay_test.go`, `tests/contract_test.go`, `tests/integration_test.go`: every `Commit(` call becomes a
  return-route call or is deleted if it only tested the public `Commit`; every `GatewayStart{BuyOrder…, ReturnURL…}`
  becomes `GatewayStart{PaymentId: …}` with `Config.ReturnURL` set; `Open(x)` → `Open(x, "https://example.cl"+transbank.ReturnPath)`.
  New tests: `Provider() == "transbank"`; `Start` without `ReturnURL` → `ErrNoReturnURL`; `Start` returns `TokenField == "token_ws"`.
- `README.md`: wiring becomes

```go
wp, err := transbank.Open(env.Get(transbank.EnvWebpay), publicURL+transbank.ReturnPath)
pm, err := paymentmanager.New(db, paymentmanager.Deps{Gateway: wp, Charges: …})
wp.MountReturn(r, pm, "/")
```

  and `docs/ARCHITECTURE.md`: the commit happens in the return route; `docs/diagrams/flows.md`: return diagram with
  R1–R6. **Commit.**

## Stage 5 — Spanish (only what is still in English)

This repo is `veltylabs/*`: every `README.md`, `docs/*` file, diagram and code comment is in Spanish, written for a
reader with little technical background (code, identifiers and keys stay English). List the English ones with
`grep -rln "[a-z] the \| is \| and " --include=*.go --include=*.md . | grep -v PLAN` and read each hit; translate
**only** the English text, never rewrite what is already Spanish. Every exported symbol whose comment lacks a usage
example gets one line `// Ejemplo: …`. **Commit.**

## Deletions — each must grep empty

```bash
grep -rn "OnlineReturn\|req.BuyOrder\|req.SessionId\|req.ReturnURL\|func (w \*Webpay) Commit" --include=*.go --include=*.md --exclude=PLAN*.md .
```

## Delivery check

```bash
go mod tidy && gotest
```

## Stages

| # | What | Files |
|---|---|---|
| 1 | Return URL in config | `config.go`, `errors.go` |
| 2 | Gateway v3 | `webpay.go` |
| 3 | Return route verifies | `return.go` |
| 4 | Tests and docs | `tests/*`, `README.md`, `docs/ARCHITECTURE.md`, `docs/diagrams/flows.md` |
| 5 | Spanish docs, diagrams, comments + usage examples | every English `*.md` and `*.go` comment |
