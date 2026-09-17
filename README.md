# ccswitch

Rotación automática de cuentas de Claude Code en Windows. Monitorea el uso real de cada cuenta (ventanas de 5h/7d), y cuando la cuenta activa se acerca al límite, swapea las credenciales OAuth por las de la cuenta con más margen — sin matar la sesión de terminal en curso.

No es una TUI ni un dashboard: es un binario chico (Go, sin CGO, un solo `.exe`) más un Windows Service que corre el ciclo de decisión en segundo plano.

## Arquitectura

```
cmd/ccswitch/
  main.go        dispatcher de subcomandos, sin frameworks de CLI
  commands.go    add / list / status / switch / auto
  service.go     wrapper de github.com/kardianos/service

internal/
  accounts/      store de cuentas (alias, email, OAuthData) -> accounts.json
  usage/         cliente HTTP del endpoint de uso de Anthropic
  swap/          lectura/escritura atómica de ~/.claude/.credentials.json
  autoswitch/    decisión pura: threshold + cooldown + mejor candidato
  config/        config.json con defaults
  serviceloop/   orquesta usage + autoswitch + swap en un solo ciclo
```

`autoswitch.Evaluate` no toca red ni disco — recibe los porcentajes ya
resueltos y devuelve una `Decision`. Eso es lo que permite testear la
lógica de rotación con tablas de casos sin mockear nada.

## El endpoint de uso

Claude Code expone el estado de rate limit de la cuenta autenticada vía:

```
GET https://api.anthropic.com/api/oauth/usage
Authorization: Bearer <access_token>
anthropic-beta: oauth-2025-04-20
```

La respuesta trae `five_hour.utilization` y `seven_day.utilization` (0-100)
más sus `resets_at`. `internal/usage.Client.Fetch` es el único punto de
contacto con ese endpoint; el resto del código nunca ve un token salvo
para reenviarlo tal cual en el header.

## Cómo swapea sin romper el `.credentials.json`

`.credentials.json` tiene más claves que solo `claudeAiOauth` (por ejemplo
`mcpOAuth`, si tenés MCP servers logueados). `swap.WriteActive` no
sobreescribe el archivo entero: lo parsea a `map[string]json.RawMessage`,
reemplaza únicamente la clave `claudeAiOauth`, y escribe con el patrón
tmp-file + `os.Rename` para que un crash a mitad de escritura no deje el
archivo corrupto.

Antes de pisar la cuenta activa con la del destino, tanto `runSwitch` como
el loop del servicio leen el token *en vivo* de esa cuenta y lo guardan de
vuelta en `accounts.json` — Claude Code refresca el access token por su
cuenta, así que el snapshot tomado en `ccswitch add` se volvería viejo sin
este paso.

## El servicio de Windows

`ccswitch service install` registra un Windows Service real vía
`kardianos/service` (no una tarea programada, no un proceso colgado a
mano). Dos detalles que importan si lo tocás:

- El servicio corre por default como **LocalSystem**, cuyo `%APPDATA%` y
  home no son los del usuario interactivo. Por eso `runServiceInstall`
  resuelve `appDataDir()` y `swap.DefaultCredentialsPath()` en el momento
  del *install* (corriendo como vos) y los hornea en
  `service.Config.Arguments` como `service run --dir X --cred Y`. El loop
  nunca vuelve a resolverlos en runtime.
- Todo el logging va a `%APPDATA%\ccswitch\ccswitch.log`, nunca a stdout.
  Si el archivo de log no se puede abrir, cae al logger nativo de Windows
  (`service.Logger`) en vez de morir en silencio.

Loop de decisión (cada `intervalSeconds`, default 60s):

1. Fetch de uso de la cuenta activa.
2. Si está por debajo del `thresholdPercent` o todavía en `cooldownSeconds`
   desde el último switch, corta ahí — no gasta requests de más
   consultando al resto de las cuentas.
3. Si hay que rotar, fetch del resto de las cuentas y `autoswitch.Evaluate`
   elige la de mayor margen (nunca rota a una cuenta igual o peor).
4. Guarda el token viejo, escribe el nuevo, persiste `accounts.json`.

## Instalar

```powershell
irm https://raw.githubusercontent.com/uaigasp/ccswitch/main/install.ps1 | iex
```

O manual: `.exe` desde [Releases](https://github.com/uaigasp/ccswitch/releases),
o compilar con `go build ./cmd/ccswitch` (Go 1.23+, sin dependencias fuera
de `kardianos/service`).

## Uso

```
ccswitch add --alias principal --email vos@ejemplo.com
ccswitch list
ccswitch status
ccswitch switch secundaria
ccswitch auto --dry-run          # corre un ciclo de decisión sin tocar nada, útil para debug
ccswitch service install         # requiere admin
ccswitch service status
ccswitch service uninstall
```

## Configuración

`%APPDATA%\ccswitch\config.json`:

```json
{
  "thresholdPercent": 90,
  "intervalSeconds": 60,
  "cooldownSeconds": 300
}
```

Un archivo parcial (por ejemplo `{"thresholdPercent": 80}`) no deja los
campos faltantes en cero: `config.Load` unmarshalea sobre los defaults.

## Logs

`%APPDATA%\ccswitch\ccswitch.log`

## Tests

```
go test ./...
```

Sin mocks de HTTP con librerías externas — `internal/usage` se testea con
`httptest.Server`, y `internal/serviceloop` con un `usageFetcher` fake
inyectado, sin tocar red ni el filesystem real fuera de `t.TempDir()`.
