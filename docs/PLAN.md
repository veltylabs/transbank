---
PLAN: "feat: transbank v0.1.0 — Webpay Plus REST v1.2 como Gateway de payment_manager, con ruta de retorno"
TAG: v0.1.0
EXECUTOR: jules
REVIEWER: none
STATUS: review
SESSION: 9854697969735076693
PR: https://github.com/veltylabs/transbank/pull/1
REVIEWED_COMMIT: 8259e086c66be25305d862ab33d6c8044134ac25
---

> Este plan se despacha con el flujo CodeJob. Ver skill: agents-workflow.

> **Pieza T** de
> [`mjosefa-cms/docs/RESERVA_Y_PAGOS_MASTER_PLAN.md`](https://github.com/veltylabs/mjosefa-cms/blob/main/docs/RESERVA_Y_PAGOS_MASTER_PLAN.md)
> (RP-3, RP-13). **Requiere `github.com/veltylabs/payment_manager` v0.1.0 publicado** (pieza P):
> usa sus tipos `Gateway`, `GatewayStart`, `GatewayStarted`, `GatewayResult`, `GatewayRefund`,
> `OnlineReturn`, `Payment` y las constantes `Outcome*`, `RefundKind*`, `Currency*`,
> `QueryPayment`, `QueryStatus`. Lee primero [`AGENTS.md`](../AGENTS.md) y
> [`docs/diagrams/flows.md`](diagrams/flows.md).

# Plan — `transbank`: Webpay Plus para `payment_manager`

## 0. Contexto (leer primero)

Repo nuevo: solo existe el esqueleto de `gonew` (`transbank.go` con un tipo vacío), que **se
borra** en la etapa 0.

`payment_manager` cobra online a través de un puerto `Gateway`. Este repo lo implementa contra
la API REST v1.2 de Webpay Plus y además registra la ruta HTTP a la que Webpay devuelve al
cliente después de pagar.

Hechos de la API (documentación oficial de Transbank Developers):

| Dato | Valor |
|---|---|
| Base integración | `https://webpay3gint.transbank.cl` |
| Base producción | `https://webpay3g.transbank.cl` |
| Ruta | `/rswebpaytransaction/api/webpay/v1.2/transactions` |
| Cabeceras | `Tbk-Api-Key-Id: <código de comercio>`, `Tbk-Api-Key-Secret: <llave>`, `Content-Type: application/json` |
| Credenciales de integración (públicas) | código `597055555532`, llave `579B532A7440BB0C9079DED94D31EA1615BACEB56610332264630D42D0A36B1C` |
| Crear | `POST {ruta}` con `buy_order` (≤ 26), `session_id` (≤ 61), `amount`, `return_url` (≤ 256) → `token`, `url` |
| Confirmar | `PUT {ruta}/{token}` → `vci`, `amount`, `status`, `buy_order`, `session_id`, `card_detail.card_number`, `accounting_date`, `transaction_date`, `authorization_code`, `payment_type_code`, `response_code`, `installments_amount`, `installments_number`, `balance` |
| Estado | `GET {ruta}/{token}` → mismos campos; `status` ∈ `INITIALIZED`, `AUTHORIZED`, `REVERSED`, `FAILED`, `NULLIFIED`, `PARTIALLY_NULLIFIED`, `CAPTURED` |
| Devolver | `POST {ruta}/{token}/refunds` con `amount` → `type` (`REVERSED` o `NULLIFIED`); si `NULLIFIED` además `authorization_code`, `authorization_date`, `balance`, `nullified_amount`, `response_code` |
| Retorno normal | llega `token_ws` (aprobado o rechazado: hay que confirmar) |
| Abandono | llegan `TBK_TOKEN`, `TBK_ORDEN_COMPRA`, `TBK_ID_SESION` |
| Formulario vencido | llegan `TBK_ORDEN_COMPRA`, `TBK_ID_SESION` (sin token) |
| Error con "volver al sitio" | llegan `token_ws` **y** `TBK_TOKEN` (se trata como abandono) |
| Método del retorno | GET (API ≥ 1.1); el abandono en integración llega por POST de formulario |

## Design gate

1. **Prior art:** SDK oficiales de Transbank (`transbank-sdk-*`: `Transaction.create/commit/status/refund`);
   `veltylabs/mikrotik` como adaptador concreto de un puerto del ecosistema (`network.Gateway`)
   configurado por una variable de entorno con URL.
2. **Prueba del nombre:** `transbank.Open(raw)` "abre la configuración de Webpay";
   `transbank.New(cfg)`; `(*Webpay).MountReturn(r, completer, page)` "monta la ruta a la que vuelve
   Webpay"; `ReturnPath` "la ruta"; `EnvWebpay` "la variable de entorno".
3. **Libro de complejidad:** 1 tipo `Webpay` (implementa `paymentmanager.Gateway`), 1 `Config`,
   1 interfaz `Completer`, `Open`, `New`, `MountReturn`, constantes de API y de configuración,
   modelos JSON generados.
4. **Dónde vive:** adaptador concreto fuera de los módulos de dominio (RP-3).
5. **Qué borra:** `transbank.go` del esqueleto.

## 1. Reglas (de `AGENTS.md`, inline)

- Sin `net/http`, `encoding/json`, `strings`, `strconv`, `errors`, `fmt`/`time`/`net/url` de la
  stdlib en código que no sea de prueba. HTTP con `webtyp.com/fetch`; JSON con `webtyp.com/json`
  sobre modelos `ormc`; texto y errores con `webtyp.com/fmt`; la ruta con `webtyp.com/router`.
- Sin `map` en ningún archivo, pruebas incluidas. Sin `reflect`.
- Literales de API (rutas, cabeceras, estados, parámetros) solo en constantes exportadas.
- Nunca escribir la llave secreta en un log ni en un mensaje de error.
- **Todas las pruebas en `tests/`**; `net/http/httptest` solo en pruebas `//go:build !wasm`.
- Documentación y comentarios en español.

## 1.1 Método: diagramas → pruebas en rojo → código (TDD), 100% de cobertura

- Fuente de verdad: [`docs/diagrams/flows.md`](diagrams/flows.md). Cada rama tiene en §9 su
  prueba con nombre exacto.
- En cada etapa 1 a 5: escribir sus pruebas de §9 → `gotest` → **confirmar que fallan** →
  commit `test(<etapa>): pruebas en rojo` → implementar → verde → commit `feat(<etapa>): …`.
- **Cobertura 100%** del paquete `transbank` (incluido `model_orm.go`):

  ```bash
  go test -coverpkg=github.com/veltylabs/transbank -coverprofile=/tmp/cover.out ./tests/... \
    && go tool cover -func=/tmp/cover.out | tail -1
  # debe imprimir: total: (statements) 100.0%
  ```

  Una línea inalcanzable desde la API pública se borra; nunca se exporta algo para cubrirla.

## 2. Etapa 0 — esqueleto

- Prerrequisitos: `go install webtyp.com/devflow/cmd/gotest@latest` y
  `go install webtyp.com/ormc/cmd/ormc@latest`.
- Borrar `transbank.go`. `go get github.com/veltylabs/payment_manager@v0.1.0 webtyp.com/fetch
  webtyp.com/json webtyp.com/fmt webtyp.com/router webtyp.com/model`.

## 3. Etapa 1 — `config.go` (diagrama §1)

```go
const EnvWebpay = "TRANSBANK_WEBPAY" // webpay://<código>:<llave>@integration|production

const (
	EnvironmentIntegration = "integration"
	EnvironmentProduction  = "production"
	IntegrationBaseURL     = "https://webpay3gint.transbank.cl"
	ProductionBaseURL      = "https://webpay3g.transbank.cl"
	IntegrationCommerceCode = "597055555532"
	IntegrationAPIKey       = "579B532A7440BB0C9079DED94D31EA1615BACEB56610332264630D42D0A36B1C"
	DefaultTimeoutMs        = 30000
)

type Config struct {
	CommerceCode string
	APIKey       string
	BaseURL      string // una de las dos de arriba; otra solo para un servidor falso en pruebas
	TimeoutMs    int    // 0 = DefaultTimeoutMs
}

// Open interpreta el valor de EnvWebpay. No llama a Transbank.
func Open(raw string) (*Webpay, error)
func New(cfg Config) *Webpay
```

Errores exactos (`domainError`): `ErrConfigScheme` = `transbank: config must start with webpay://`,
`ErrConfigFormat` = `transbank: config must be webpay://<code>:<key>@<environment>`,
`ErrConfigCredentials` = `transbank: commerce code and key are required`,
`ErrConfigEnvironment` = `transbank: environment must be integration or production`.
El parseo es a mano con `webtyp.com/fmt` (sin `net/url`); el mensaje de error **nunca** incluye
la llave.

## 4. Etapa 2 — `model.go` (+ `model_orm.go`) y `api.go` (diagrama §2, §3)

Modelos `model.Definition` con los nombres JSON exactos de Webpay:
`create_request` (`buy_order`, `session_id`, `amount` (Float), `return_url`),
`create_response` (`token`, `url`), `card_detail` (`card_number`),
`transaction_response` (todos los campos de la tabla §0, con `card_detail` como
`model.Struct(&CardDetailModel)`), `refund_request` (`amount`), `refund_response` (`type`,
`authorization_code`, `authorization_date`, `balance`, `nullified_amount`, `response_code`),
`api_error` (`error_message`).

`api.go`: una función privada `call(method, path string, body model.Encodable, out model.Decodable) error`
que arma el pedido con `fetch` (cabeceras de §0, `ContentTypeJSON`, `Timeout`) y **espera la
respuesta con un canal** (los métodos del puerto son síncronos y solo se usan en el servidor).
Errores según diagrama §2 (textos exactos: `transbank: network: <err>`,
`transbank: HTTP <n>: <error_message>`, `transbank: HTTP <n>`, `transbank: bad response`).

Monto (diagrama §3): `CLP` → entero; cualquier otra (incluida `USD`) →
`ErrCurrency` = `transbank: unsupported currency`. De vuelta, `amount` de Webpay → entero
redondeado.

> **Desviación en la revisión (PR #1):** el plan pedía USD como `monto/100`, pero `Commit`,
> `Status` y `Refund` del `Gateway` no traen la moneda: el monto de vuelta no se podría
> convertir. Se cobra solo CLP; USD exige antes extender el puerto de `payment_manager`.

## 5. Etapa 3 — `webpay.go`: el `Gateway` (diagramas §4, §5)

```go
type Webpay struct{ /* cfg */ }

var _ paymentmanager.Gateway = (*Webpay)(nil)

func (w *Webpay) Start(req paymentmanager.GatewayStart) (paymentmanager.GatewayStarted, error)
func (w *Webpay) Commit(token string) (paymentmanager.GatewayResult, error)
func (w *Webpay) Status(token string) (paymentmanager.GatewayResult, error)
func (w *Webpay) Refund(token string, amount int64) (paymentmanager.GatewayRefund, error)
```

- `Start`: `buy_order = req.BuyOrder`, `session_id = req.SessionId`; si alguno excede su largo
  (26 / 61) o `return_url` excede 256 ⇒ `ErrFieldTooLong` = `transbank: field exceeds Webpay length`
  sin llamar a la API. Devuelve `Token = token`, `RedirectURL = url`.
- `Commit` (`PUT`) y `Status` (`GET`): traducir con el diagrama §4. `CardLast4` = últimos 4 de
  `card_detail.card_number`; `Installments = installments_number`.
- `Refund`: diagrama §5.

## 6. Etapa 4 — `return.go`: la ruta de retorno (diagrama §6)

```go
const ReturnPath = "/payments/webpay/return"

const (
	ParamToken     = "token_ws"
	ParamAbort     = "TBK_TOKEN"
	ParamBuyOrder  = "TBK_ORDEN_COMPRA"
	ParamSession   = "TBK_ID_SESION"
	StatusError    = "ERROR"
)

// Completer es lo que la ruta necesita de payment_manager (lo satisface *paymentmanager.Module).
type Completer interface {
	CompleteOnlinePayment(ret paymentmanager.OnlineReturn) (paymentmanager.Payment, error)
}

// MountReturn registra GET y POST en ReturnPath, ambos .Public() (Webpay no trae sesión;
// el token solo sirve con la llave secreta). page es la página de la app a la que se
// redirige, p. ej. "/".
func (w *Webpay) MountReturn(r router.Router, c Completer, page string)
```

- POST: parámetros del cuerpo `application/x-www-form-urlencoded` leídos con
  `router.QueryParam("?"+string(ctx.Body()), clave)`. GET: `router.QueryParam(ctx.Path(), clave)`.
- Redirección: `ctx.SetHeader("Location", destino)` y `ctx.WriteStatus(303)`, con
  `destino = page + "?" + paymentmanager.QueryPayment + "=" + id + "&" + paymentmanager.QueryStatus + "=" + estado + "#" + returnTo`
  (`#returnTo` solo si `ReturnTo != ""`), o `page + "?" + QueryStatus + "=ERROR"`.
- La ruta nunca responde 5xx al navegador: todo termina en 303.

## 7. Etapa 5 — `README.md` y `docs/ARCHITECTURE.md`

- README: qué hace; configuración (`TRANSBANK_WEBPAY`, ejemplo de integración con las
  credenciales públicas); cableado en una app:

  ```go
  wp, err := transbank.Open(env.Get(transbank.EnvWebpay))
  pm, err := paymentmanager.New(db, paymentmanager.Deps{
      Gateway: wp, ReturnURL: publicURL + transbank.ReturnPath, Charges: …,
  })
  wp.MountReturn(r, pm, "/")
  ```

- ARCHITECTURE: el diagrama de secuencia de `payment_manager` (enlace), tabla de traducción de
  estados, reglas de seguridad (llave fuera de logs, ruta pública solo con token).

## 8. Etapa 6 — prueba contra el ambiente real (opcional, con tag)

`tests/integration_test.go` con `//go:build transbank_integration`: `Open` con las credenciales
públicas de integración, `Start` de $1.000 y `Status` del token ⇒ `pending`. No corre en el
`gotest` por defecto.

## 9. Pruebas — trazabilidad diagrama → prueba (`tests/`)

Servidor falso en `tests/fake_webpay_test.go` (`//go:build !wasm`, `httptest`): registra cada
pedido (método, ruta, cabeceras, cuerpo) en un slice y responde lo que la prueba configure.
`Completer` falso en `tests/setup_test.go`.

| Rama | Prueba | Archivo | Etapa |
|---|---|---|---|
| `O1-no` | `TestOpen_WrongScheme` | `config_test.go` | 1 |
| `O2-no` | `TestOpen_BadFormat` | `config_test.go` | 1 |
| `O3-no` | `TestOpen_EmptyCredentials` | `config_test.go` | 1 |
| `O4-integration`, `O4-production` | `TestOpen_Environments` | `config_test.go` | 1 |
| `O4-otro` | `TestOpen_UnknownEnvironment` | `config_test.go` | 1 |
| la llave nunca aparece en un error | `TestOpen_ErrorsNeverContainKey` | `config_test.go` | 1 |
| `New` con `TimeoutMs` 0 usa el valor por defecto | `TestNew_DefaultTimeout` (se observa: un servidor falso que tarda más que un timeout corto configurado da error de red; con 0 no) | `config_test.go` | 1 |
| modelos: ida y vuelta | `TestModels_RoundTrip` | `models_test.go` | 2 |
| cabeceras y JSON enviados | `TestCall_SendsHeadersAndBody` | `api_test.go` | 2 |
| `C1-sí` | `TestCall_NetworkError` (servidor cerrado) | `api_test.go` | 2 |
| `C2-no/C3-sí` | `TestCall_HTTPErrorWithMessage` | `api_test.go` | 2 |
| `C2-no/C3-no` | `TestCall_HTTPErrorWithoutMessage` | `api_test.go` | 2 |
| `C4-no` | `TestCall_BadJSON` | `api_test.go` | 2 |
| `M1-CLP`, `M1-USD` (⇒ `ErrCurrency`), `M1-otra` | `TestAmount_CLP`, `TestAmount_USD`, `TestAmount_Unsupported` | `amount_test.go` | 2 |
| `Start` feliz | `TestStart_ReturnsTokenAndURL` (verifica cuerpo enviado) | `webpay_test.go` | 3 |
| `Start` largos | `TestStart_FieldTooLongSkipsAPI` (tres casos) | `webpay_test.go` | 3 |
| `S1-AUTHORIZED/S2-sí` | `TestCommit_Authorized` (código, tipo, cuotas, últimos 4, monto) | `webpay_test.go` | 3 |
| `S2-no` | `TestCommit_AuthorizedWithNonZeroCodeIsRejected` | `webpay_test.go` | 3 |
| `S1-INITIALIZED` | `TestStatus_InitializedIsPending` | `webpay_test.go` | 3 |
| `S1-FAILED…` | `TestStatus_TerminalStatesAreRejected` (los 4 estados) | `webpay_test.go` | 3 |
| `S1-otro` | `TestStatus_UnknownStatus` | `webpay_test.go` | 3 |
| `F1-REVERSED` | `TestRefund_Reversed` | `webpay_test.go` | 3 |
| `F1-NULLIFIED/F2-sí` | `TestRefund_Nullified` | `webpay_test.go` | 3 |
| `F2-no` | `TestRefund_NullifiedRejected` | `webpay_test.go` | 3 |
| `F1-otro` | `TestRefund_UnknownType` | `webpay_test.go` | 3 |
| rutas registradas públicas | `TestMountReturn_RegistersPublicGetAndPost` | `return_test.go` | 4 |
| `R1-no` + `R2-token` + `R3-sí` | `TestReturn_GetTokenRedirectsWithStatus` (Location exacto, 303) | `return_test.go` | 4 |
| `R1-sí` + `R2-TBK_TOKEN` | `TestReturn_PostAbortFromFormBody` | `return_test.go` | 4 |
| `token_ws` + `TBK_TOKEN` | `TestReturn_TokenAndAbortIsAbort` | `return_test.go` | 4 |
| `R2-solo orden` | `TestReturn_TimeoutPassesBuyOrder` | `return_test.go` | 4 |
| `R2-ninguno` | `TestReturn_NoParamsRedirectsError` | `return_test.go` | 4 |
| `R3-no` | `TestReturn_CompleterErrorRedirectsError` | `return_test.go` | 4 |
| error con pago | `TestReturn_ErrorWithPaymentReportsItsStatus` | `return_test.go` | 4 |
| sin `return_to` | `TestReturn_NoReturnToOmitsHash` | `return_test.go` | 4 |
| contrato | `TestWebpay_ImplementsGateway` (asignación a la interfaz) | `contract_test.go` | 3 |

## 10. Criterios de aceptación

- `gotest` en verde y badge de cobertura en 100%; el comando de §1.1 imprime `100.0%`.
- Cada prueba de §9 existe con su nombre exacto en `tests/`; no hay `_test.go` fuera de `tests/`
  (`find . -name "*_test.go" -not -path "./tests/*"` → vacío).
- `grep -rnE '"(net/http|encoding/json|strings|strconv|errors|net/url)"' --include=*.go . | grep -v "^./tests/"` → vacío.
- `grep -rn "map\[" --include=*.go .` → vacío.
- `grep -rn "579B532A" --include=*.go . | grep -v IntegrationAPIKey` → vacío (la llave pública
  solo en su constante).
- El historial del PR muestra un commit `test(...)` en rojo antes de cada `feat(...)`.
- Antes de abrir el PR: `git diff --stat main` lista todos los archivos de la tabla de etapas, y
  no se borra ningún archivo que este plan no ordene borrar.

## 11. Tabla de etapas

| Etapa | Archivos |
|---|---|
| 0 | borrar `transbank.go`; `go.mod`, `go.sum` |
| 1 | `config.go`, `errors.go`, `tests/config_test.go`, `tests/setup_test.go` |
| 2 | `model.go`, `model_orm.go`, `api.go`, `amount.go`, `tests/models_test.go`, `tests/api_test.go`, `tests/amount_test.go`, `tests/fake_webpay_test.go` |
| 3 | `webpay.go`, `tests/webpay_test.go`, `tests/contract_test.go` |
| 4 | `return.go`, `tests/return_test.go` |
| 5 | `README.md`, `docs/ARCHITECTURE.md` |
| 6 | `tests/integration_test.go` |
