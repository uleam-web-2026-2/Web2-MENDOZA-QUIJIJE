$ErrorActionPreference = "Stop"
$servidor = Split-Path $PSScriptRoot -Parent
$rutaEnv = Join-Path $servidor ".env"
if (Test-Path -LiteralPath $rutaEnv) {
    Write-Host "Ya existe servidor/.env; se conserva su configuracion."
    exit 0
}
$bytes = New-Object byte[] 32
$rng = [System.Security.Cryptography.RandomNumberGenerator]::Create()
$rng.GetBytes($bytes)
$claveBD = ([BitConverter]::ToString($bytes)).Replace("-", "").ToLowerInvariant()
$rng.GetBytes($bytes)
$secretoJWT = ([BitConverter]::ToString($bytes)).Replace("-", "").ToLowerInvariant()
$rng.Dispose()
$contenido = @"
PORT=8080
POSTGRES_PASSWORD=$claveBD
DATABASE_URL=postgres://eventosports:$claveBD@localhost:5433/eventosports?sslmode=disable
JWT_SECRET=$secretoJWT
DEMO_SEED=true
"@
[IO.File]::WriteAllText($rutaEnv, $contenido, (New-Object System.Text.UTF8Encoding($false)))
Write-Host "Configuracion local creada con claves aleatorias. No se mostraron ni se suben a Git."
