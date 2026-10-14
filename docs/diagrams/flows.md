# Flujos de `transbank`

Cada rombo lleva un id. La tabla de trazabilidad de [`docs/PLAN.md`](../PLAN.md) §9 asigna a
cada rama (`O1-sí`, `O1-no`, …) su prueba. Una rama sin prueba es un defecto.

## 1. Configuración (`Open`)

Formato de `TRANSBANK_WEBPAY`: `webpay://<código de comercio>:<llave secreta>@<ambiente>`,
con ambiente `integration` o `production`.

```mermaid
flowchart TD
    A[Open raw] --> O1{O1: ¿empieza con webpay://?}
    O1 -- no --> E1[ErrConfigScheme]
    O1 -- sí --> O2{O2: ¿tiene código:llave@ambiente?}
    O2 -- no --> E2[ErrConfigFormat]
    O2 -- sí --> O3{O3: ¿código y llave no vacíos?}
    O3 -- no --> E3[ErrConfigCredentials]
    O3 -- sí --> O4{O4: ¿ambiente?}
    O4 -- integration --> I[BaseURL = IntegrationBaseURL]
    O4 -- production --> P[BaseURL = ProductionBaseURL]
    O4 -- otro --> E4[ErrConfigEnvironment]
    I --> G[New Config]
    P --> G
```

## 2. Llamada a la API (común a Start, Commit, Status, Refund)

```mermaid
flowchart TD
    A[call método, ruta, cuerpo] --> H[fetch con Tbk-Api-Key-Id,<br>Tbk-Api-Key-Secret, JSON, Timeout]
    H --> C1{C1: ¿error de red?}
    C1 -- sí --> E1[error envuelto: transbank: network]
    C1 -- no --> C2{C2: ¿HTTP 2xx?}
    C2 -- no --> C3{C3: ¿cuerpo con error_message?}
    C3 -- sí --> E2[error: transbank: HTTP N: mensaje]
    C3 -- no --> E3[error: transbank: HTTP N]
    C2 -- sí --> C4{C4: ¿JSON decodifica?}
    C4 -- no --> E4[error: transbank: bad response]
    C4 -- sí --> OK[respuesta tipada]
```

## 3. Traducir moneda y monto

```mermaid
flowchart TD
    A[monto en unidades mínimas, moneda] --> M1{M1: ¿moneda?}
    M1 -- CLP --> CLP[entero tal cual]
    M1 -- USD --> E[ErrCurrency]
    M1 -- otra --> E
```

Solo CLP. `Commit`, `Status` y `Refund` del `Gateway` no reciben la moneda, así que un monto
USD que sale como decimal no se podría devolver a unidades mínimas; además Webpay cobra USD con
un código de comercio aparte. Cobrar en USD requiere primero extender el puerto en
`payment_manager`.

## 4. Traducir el resultado de Commit/Status (`GatewayResult.Outcome`)

```mermaid
flowchart TD
    A[status, response_code] --> S1{S1: ¿status?}
    S1 -- AUTHORIZED --> S2{S2: ¿response_code == 0?}
    S2 -- sí --> AUTH[authorized]
    S2 -- no --> REJ1[rejected]
    S1 -- INITIALIZED --> PEN[pending]
    S1 -- "FAILED, REVERSED,<br>NULLIFIED, PARTIALLY_NULLIFIED" --> REJ2[rejected]
    S1 -- otro --> E[error: transbank: unknown status]
```

## 5. Traducir la devolución

```mermaid
flowchart TD
    A[respuesta de refunds] --> F1{F1: ¿type?}
    F1 -- REVERSED --> REV[Kind REVERSED,<br>Amount = monto pedido]
    F1 -- NULLIFIED --> F2{F2: ¿response_code == 0?}
    F2 -- sí --> NUL[Kind NULLIFIED,<br>Amount = nullified_amount,<br>AuthorizationCode]
    F2 -- no --> E1[error: transbank: refund rejected]
    F1 -- otro --> E2[error: transbank: unknown refund type]
```

## 6. Ruta de retorno (`MountReturn`)

Webpay vuelve con GET (producción) o POST de formulario (abandono en integración).

```mermaid
flowchart TD
    A[GET o POST ReturnPath] --> R1{R1: ¿método POST?}
    R1 -- sí --> PB[leer parámetros del cuerpo<br>application/x-www-form-urlencoded]
    R1 -- no --> PQ[leer parámetros de la query]
    PB --> R2
    PQ --> R2{R2: ¿qué parámetros?}
    R2 -- "token_ws sin TBK_TOKEN" --> N[OnlineReturn Token]
    R2 -- TBK_TOKEN --> AB[OnlineReturn AbortToken]
    R2 -- "solo TBK_ORDEN_COMPRA" --> TO[OnlineReturn BuyOrder]
    R2 -- ninguno --> BAD[303 a Page?status=ERROR]
    N --> CP[Completer.CompleteOnlinePayment]
    AB --> CP
    TO --> CP
    CP --> R3{R3: ¿devolvió pago con id?}
    R3 -- sí --> OK[303 a Page?payment=id&status=ESTADO#return_to]
    R3 -- no --> ERR[303 a Page?status=ERROR]
```

Un error con pago (p. ej. `ErrAmountMismatch` junto al pago `FAILED`) cuenta como `R3-sí`: se
informa el estado del pago.
