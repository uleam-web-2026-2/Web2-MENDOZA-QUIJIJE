param(
    [string]$ApiBaseUrl = "http://localhost:8080"
)

$ErrorActionPreference = "Stop"
$servidor = Split-Path $PSScriptRoot -Parent
$raiz = Split-Path $servidor -Parent
if (!(Test-Path -LiteralPath (Join-Path $servidor ".env")) -and !$env:TEST_DATABASE_URL) {
    throw "Falta configurar la base. Ejecute preparar-entorno.ps1 y arranque PostgreSQL."
}
Push-Location $raiz
try {
    Write-Host "1/5 Validando contrato..."
    & npx.cmd --yes @redocly/cli lint docs/openapi.yaml --format=stylish
    if ($LASTEXITCODE -ne 0) { throw "Fallo el contrato" }
    Set-Location $servidor
    Write-Host "2/5 Ejecutando pruebas, incluida PostgreSQL..."
    & go test ./... -count=1
    if ($LASTEXITCODE -ne 0) { throw "Fallaron las pruebas" }
    Write-Host "3/5 Revisando Go..."
    & go vet ./...
    if ($LASTEXITCODE -ne 0) { throw "Fallo go vet" }
    Write-Host "4/5 Compilando..."
    & go build ./...
    if ($LASTEXITCODE -ne 0) { throw "Fallo la compilacion" }
    Set-Location (Join-Path $raiz "docs/hito2_bruno")
    Write-Host "5/5 Ejecutando peticiones Bruno (la API debe estar encendida)..."
    & npx.cmd --yes @usebruno/cli@4.3.0 run --env Local --env-var "baseUrl=$ApiBaseUrl" --bail
    if ($LASTEXITCODE -ne 0) { throw "Fallo la coleccion de Bruno" }
    Write-Host "Hito 2: contrato, pruebas, compilacion y peticiones correctos."
} finally {
    Pop-Location
}
