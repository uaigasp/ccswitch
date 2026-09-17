# ccswitch

Cambia entre varias cuentas de Claude Code en Windows, solo, antes de que se llene el limite de uso (5 horas / 7 dias).

## Como funciona

Guarda el token de cada cuenta que agregues. Cuando el uso de la cuenta activa llega al 90%, revisa las otras cuentas guardadas y cambia a la que tenga mas margen libre. El cambio es instantaneo, no hace falta reiniciar la sesion de terminal donde esta corriendo Claude Code.

## Instalar

Opcion rapida (Windows, PowerShell):

```powershell
irm https://raw.githubusercontent.com/uaigasp/ccswitch/main/install.ps1 | iex
```

Opcion manual: bajar `ccswitch.exe` de [Releases](https://github.com/uaigasp/ccswitch/releases) o compilar con `go build ./cmd/ccswitch`.

## Uso

```
ccswitch add --alias principal --email vos@ejemplo.com    agrega la cuenta logueada ahora
ccswitch list                                              lista las cuentas y su uso
ccswitch status                                             muestra la cuenta activa
ccswitch switch secundaria                                   cambia de cuenta a mano
ccswitch auto --once --dry-run                                simula un ciclo de auto-switch sin tocar archivos reales
ccswitch service install                                      instala el auto-switch (correr como admin)
ccswitch service status                                       dice si esta corriendo
ccswitch service uninstall                                    lo saca
```

## Configuracion

Editar `%APPDATA%\ccswitch\config.json`:

```json
{
  "thresholdPercent": 90,
  "intervalSeconds": 60,
  "cooldownSeconds": 300
}
```

## Logs

`%APPDATA%\ccswitch\ccswitch.log`
