# Arquitectura de `transbank`

`transbank` implementa `paymentmanager.Gateway` contra Webpay Plus REST v1.2 y no guarda
estado: la verdad del pago vive en `payment_manager`, que confirma cada transacción con el
servidor de Webpay (nunca confía en lo que trae el navegador). El flujo completo, de la
reserva al pago confirmado, está en el diagrama de secuencia de
[payment_manager](https://github.com/veltylabs/payment_manager/blob/main/docs/diagrams/flows.md).

## Piezas

```
payment_manager ──Gateway──▶ Webpay (Start/Status/Refund) ──call──▶ API Webpay
       ▲                                                           ▲
       └─Completer── MountReturn ◀── navegador (token_ws / …)      │
                          │                                        │
                          └────── commit ──────────────────────────┘
```

| Operación | API | Método |
|---|---|---|
| `Start` | `/rswebpaytransaction/api/webpay/v1.2/transactions` | `POST` |
| `commit` (interno) | `…/transactions/{token}` | `PUT` |
| `Status` | `…/transactions/{token}` | `GET` |
| `Refund` | `…/transactions/{token}/refunds` | `POST` |

## Traducción de estados

| Webpay `status` | `response_code` | `GatewayResult.Outcome` |
|---|---|---|
| `AUTHORIZED` | `0` | `authorized` |
| `AUTHORIZED` | otro | `rejected` |
| `INITIALIZED` | — | `pending` |
| `FAILED`, `REVERSED`, `NULLIFIED`, `PARTIALLY_NULLIFIED` | — | `rejected` |
| otro | — | error `transbank: unknown status` |

| Webpay refund `type` | `response_code` | `GatewayRefund` |
|---|---|---|
| `REVERSED` | — | `Kind` reversed, `Amount` = monto pedido |
| `NULLIFIED` | `0` | `Kind` nullified, `Amount` = `nullified_amount`, `AuthorizationCode` |
| `NULLIFIED` | otro | error `transbank: refund rejected` |
| otro | — | error `transbank: unknown refund type` |

## Ruta de retorno

`MountReturn` registra `GET` y `POST` en `ReturnPath`, ambos públicos: Webpay no trae sesión.
La ruta de retorno solo identifica el pago (obteniendo el ID del pago y el token del query o
del cuerpo); `payment_manager` luego pregunta a Transbank a través de `Confirm` cuál es su
estado real. Esto asegura que un retorno falsificado desde el navegador no puede cambiar
un pago.

- `TBK_TOKEN` (con o sin `token_ws`) ⇒ el token que se pasa es `TBK_TOKEN` (abandono).
- Solo `token_ws` ⇒ el token que se pasa es `token_ws`.
- Nada ⇒ 303 a `page?status=ERROR`, sin llamar a `payment_manager`.

Siempre responde 303, nunca 5xx: un error que igual identifica el pago lleva a la pantalla
del pago con su estado; si no lo identifica, a `status=ERROR` sin detalles.

## Moneda

Solo CLP (entero, sin decimales). El plan pedía también USD como `monto/100`, pero
`Commit`, `Status` y `Refund` del puerto no reciben la moneda, así que un monto USD no se
podría devolver a unidades mínimas; además Webpay cobra USD con un código de comercio aparte.
Cobrar en USD requiere antes extender el puerto en `payment_manager`.

## Seguridad

- La llave secreta solo viaja en la cabecera `Tbk-Api-Key-Secret`; ningún error, log ni
  redirección la incluye (`TestOpen_ErrorsNeverContainKey`).
- `buy_order` (26), `session_id` (61) y `return_url` (256) se validan antes de llamar a la API.
- La ruta pública solo acepta tokens: no cambia nada que `payment_manager` no haya confirmado
  con Webpay.

## Restricciones

Ver [AGENTS.md](../AGENTS.md): sin `map`, sin stdlib de texto/JSON/HTTP en código del paquete
(`webtyp.com/fmt`, `webtyp.com/json`, `webtyp.com/fetch`), pruebas solo en `tests/`.
