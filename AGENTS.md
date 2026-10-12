# AGENTS.md — veltylabs/transbank

Notas de trabajo para agentes IA. Diseño: [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md);
flujos: [docs/diagrams/flows.md](docs/diagrams/flows.md).

## Misión

Adaptador concreto de **Transbank Webpay Plus (API REST v1.2)** para el puerto
`paymentmanager.Gateway` de [`github.com/veltylabs/payment_manager`](https://github.com/veltylabs/payment_manager).
Implementa el contrato; nunca lo redefine. Si al contrato le falta algo, la corrección es un
plan contra `payment_manager`, nunca un tipo local que lo duplique.

También es dueño de la **ruta HTTP de retorno** de Webpay: un módulo de dominio no puede
registrar rutas, un adaptador sí. Solo lo importa la raíz de composición de una app
(`config/`), nunca un módulo `veltylabs/modules/*`.

## Reglas del ecosistema webtyp (sin excepción)

Aunque hoy corre solo en el servidor, este repo sigue las mismas reglas que el resto del
ecosistema, para que compile en TinyGo/wasm y no arrastre la stdlib:

- **Prohibido** `net/http`, `encoding/json`, `strings`, `strconv`, `errors`, `fmt` (stdlib),
  `net/url`, `time` (stdlib) en código que no sea de prueba. Reemplazos:
  - HTTP saliente → `webtyp.com/fetch` (`fetch.Post(...).BaseURL(...).Header(...).Send(...)`).
  - JSON → `webtyp.com/json` sobre modelos `model.Definition` generados con `ormc`
    (`go install webtyp.com/ormc/cmd/ormc@latest && ormc`; **nunca** editar `model_orm.go`).
  - Texto, números, errores → `webtyp.com/fmt` (`fmt.Err`, `fmt.Convert`, …).
  - Ruta de retorno → `webtyp.com/router` (`router.Router`, `router.Context`,
    `router.QueryParam`).
- **Sin `map[K]V`** en ningún archivo, pruebas incluidas: slices con escaneo lineal o
  `fmt.KeyValue`.
- **Sin `reflect`.**
- **Valores de enumeración y literales repetidos solo en constantes exportadas** (rutas de la
  API, nombres de cabecera, estados de Webpay, nombres de parámetro).
- **Sin secretos en el código ni en los logs.** El código de comercio y la llave llegan por
  configuración (`EnvWebpay`); las credenciales públicas de integración de Transbank sí son
  constantes exportadas (son públicas y documentadas por Transbank).
- Documentación y comentarios en español.

## Pruebas

- **Todas en `tests/`** (paquete `tests`, solo API exportada). Ningún `_test.go` fuera de
  `tests/`.
- Ejecutor `gotest` (`go install webtyp.com/devflow/cmd/gotest@latest`), nunca `go test`.
- Webpay se simula con un servidor HTTP falso dentro de `tests/` (`net/http/httptest` está
  permitido **solo en pruebas** con `//go:build !wasm`). La ruta de retorno se prueba con
  `webtyp.com/router/mock`.
- **TDD:** cada rama de [docs/diagrams/flows.md](docs/diagrams/flows.md) tiene su prueba, escrita
  en rojo antes que el código. **Cobertura 100%** del paquete.
- Pruebas contra el ambiente real de integración de Transbank: solo con el build tag
  `transbank_integration`, nunca en el `gotest` por defecto, y nunca contra producción.

## Publicación

Plan en `docs/PLAN.md`, despacho con `codejob`; el agente nunca ejecuta `codejob` ni `gopush`.
