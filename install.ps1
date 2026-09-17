$ErrorActionPreference = "Stop"

$repo = "uaigasp/ccswitch"
$installDir = "$env:LOCALAPPDATA\ccswitch"
$exePath = "$installDir\ccswitch.exe"

Write-Host "bajando ccswitch..."
New-Item -ItemType Directory -Force -Path $installDir | Out-Null

$release = Invoke-RestMethod -Uri "https://api.github.com/repos/$repo/releases/latest"
$asset = $release.assets | Where-Object { $_.name -eq "ccswitch.exe" }

if (-not $asset) {
    throw "no se encontro ccswitch.exe en el ultimo release"
}

Invoke-WebRequest -Uri $asset.browser_download_url -OutFile $exePath

Write-Host "agregando $installDir al PATH del usuario..."
$currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($currentPath -notlike "*$installDir*") {
    [Environment]::SetEnvironmentVariable("Path", "$currentPath;$installDir", "User")
}

Write-Host ""
Write-Host "listo. abri una terminal nueva y segui estos pasos:"
Write-Host "  1. logueate con la primera cuenta en claude code"
Write-Host "  2. ccswitch add --alias principal --email vos@ejemplo.com"
Write-Host "  3. logueate con la segunda cuenta"
Write-Host "  4. ccswitch add --alias secundaria --email otra@ejemplo.com"
Write-Host "  5. ccswitch service install   (correr como administrador)"
