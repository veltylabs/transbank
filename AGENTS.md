# AGENTS.md — veltylabs/transbank

<!-- webtyp:base — copied from devskills skill agents-md; do not edit here -->
## Ecosystem rules — read before any change

`docs/PLAN.md` is the spec: no design judgment; anything unclear is a question in the PR, not a guess.

- Public API: typed (no `any`/`interface{}`), explicit, minimal; never export a symbol only a test uses.
- One intent, one path: the old path is deleted in the same change. No silent fallback, `TODO`, commented code or stub.
- Fail at compile time, else a loud diagnostic — never silently. Zero values deny.
- A defect in a dependency is reported, never wrapped or patched here.
- No string literals in logic: named constants. The same logic never lives in two repos: import the package the plan names; if you find an existing one the plan does not name that already does what you are about to write, say so in the PR instead of writing it.
- `cmd/*/main.go`: parse flags, inject, print/exit only. stdout = data, stderr = logs, exit codes are the contract.
- WASM code: `webtyp/fmt`, `webtyp/time`, `webtyp/json` instead of stdlib; no `map`; `dom.Element` embedded by value.
- `css.go`, `svg.go`, `js.go`, `html.go` carry `//go:build !wasm`; no `ssr.go`, no `front.go`.
- Build tags: `wasm`/`!wasm` only when the API does not exist on that platform; `browser`/`!browser` when the file belongs to another binary One main per folder, no destination tag: `web/browser/main.go` → `web/public/browser.wasm`, `web/server/main.go` → `web/server/server`, `web/edge/main.go` → `web/edge/edge.wasm`.
- Code and keys in English; UI text is `lang.Text`, translations live in `lang.json`. Comments and docs: English in `webtyp/*`; in `veltylabs/*` **every README, `docs/*` file, diagram and code comment is in Spanish**, written for a reader with little technical background, and each exported symbol's comment shows a short usage example.
- Tests cover behaviour and domain rules, never the look (CSS, colours, markup order, class strings). In `tests/` (external package), real stack with fakes only at the edge; run `gotest`.
- Docs (`README.md`, `docs/*`) updated with the API; permanent docs never link `docs/PLAN.md`.
- Before the PR: `git diff --stat main` lists every file of the plan's stages, and nothing the plan did not order is deleted.
<!-- /webtyp:base -->

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
