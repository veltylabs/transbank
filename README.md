# transbank
<img src="docs/img/badges.svg">

Adaptador de Transbank **Webpay Plus** (API REST v1.2) para el puerto `Gateway` de
[payment_manager](https://github.com/veltylabs/payment_manager). Crea, confirma, consulta y
devuelve transacciones, y monta la ruta pública por la que Webpay devuelve el navegador.
Diseño: [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md); flujos y ramas probadas:
[docs/diagrams/flows.md](docs/diagrams/flows.md).

Solo servidor: el navegador nunca ve la llave secreta ni llama a Webpay.

## Configuración

Una variable de entorno, `TRANSBANK_WEBPAY`:

```
webpay://<código de comercio>:<llave secreta>@integration|production
```

Ambiente de integración con las credenciales públicas de Transbank:

```
TRANSBANK_WEBPAY=webpay://597055555532:579B532A7440BB0C9079DED94D31EA1615BACEB56610332264630D42D0A36B1C@integration
```

`Open` nunca incluye la llave en sus errores. Sin `Open` se puede usar
`transbank.New(transbank.Config{...})`; `TimeoutMs` 0 usa 30 s.

## Cableado en una app

```go
wp, err := transbank.Open(env.Get(transbank.EnvWebpay), publicURL+transbank.ReturnPath)
pm, err := paymentmanager.New(db, paymentmanager.Deps{Gateway: wp, Charges: []paymentmanager.ChargeSource{{Module: "appointment_booking", Reader: booking}}})
wp.MountReturn(r, pm, "/")
```

Tras el pago, la ruta redirige (303) a `"/?payment=<id>&status=<estado>#<return_to>"`, o a
`"/?status=ERROR"` si el retorno no identifica un pago.

## Moneda

Solo CLP. Otra moneda devuelve `ErrCurrency` sin llamar a Webpay (ver
[ARCHITECTURE](docs/ARCHITECTURE.md#moneda)).

## Pruebas

```bash
gotest                                                   # unitarias, servidor Webpay falso
go test -tags transbank_integration ./tests/             # contra el ambiente de integración real
```

## Archivos

| Archivo | Qué contiene |
|---|---|
| `config.go` | `Open`, `New`, `Config`, ambientes y credenciales de integración |
| `api.go` | `call`: un pedido HTTP a Webpay con `fetch`, errores de red/HTTP/JSON |
| `amount.go` | traducción de montos |
| `webpay.go` | `Start`, `Commit`, `Status`, `Refund` (el `Gateway`) |
| `return.go` | `MountReturn`: la ruta de retorno del navegador |
| `model.go` | JSON de la API (`model_orm.go` lo genera `ormc`) |
| `errors.go` | errores exportados |
