param(
    [Parameter(Position=0)]
    [string]$Command = "help"
)

function Invoke-DockerCompose {
    param(
        [Parameter(ValueFromRemainingArguments = $true)]
        [string[]]$ComposeArgs
    )
    docker compose version *>$null
    if ($LASTEXITCODE -eq 0) {
        docker compose @ComposeArgs
    }
    else {
        docker-compose @ComposeArgs
    }
}

function Up {
    Write-Host "Starting Docker Compose stack..." -ForegroundColor Cyan
    Invoke-DockerCompose up -d --build
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
}

function Down {
    Write-Host "Stopping Docker Compose stack..." -ForegroundColor Cyan
    Invoke-DockerCompose down
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
}

function Seed {
    Write-Host "Seeding database..." -ForegroundColor Cyan
    Invoke-DockerCompose exec app ./seeder
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
    Write-Host "Database seeded" -ForegroundColor Green
}

function Test-Project {
    Write-Host "Running tests with gotestsum..." -ForegroundColor Cyan
    gotestsum --format testname -- -count=1 -cover ./...
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
}

function LoadTest {
    if (-not (Get-Command k6 -ErrorAction SilentlyContinue)) {
        Write-Host "Install k6: https://grafana.com/docs/k6/latest/set-up/install-k6/" -ForegroundColor Red
        exit 1
    }
    $repoRoot = if ([string]::IsNullOrEmpty($PSScriptRoot)) { (Get-Location).Path } else { $PSScriptRoot }
    $base = if ($env:BASE_URL) { $env:BASE_URL } else { "http://127.0.0.1:8080" }
    $rate = if ($env:LOADTEST_RATE) { $env:LOADTEST_RATE } else { "50" }
    $dur = if ($env:LOADTEST_DURATION) { $env:LOADTEST_DURATION } else { "90s" }
    Write-Host "k6 slots/list load test: BASE_URL=$base RATE=$rate DURATION=$dur" -ForegroundColor Cyan
    Push-Location $repoRoot
    try {
        k6 run loadtest/k6/slots_list.js -e "BASE_URL=$base" -e "LOADTEST_RATE=$rate" -e "LOADTEST_DURATION=$dur"
        if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
    }
    finally {
        Pop-Location
    }
}

function LoadTest-Smoke {
    if (-not (Get-Command k6 -ErrorAction SilentlyContinue)) {
        Write-Host "Install k6: https://grafana.com/docs/k6/latest/set-up/install-k6/" -ForegroundColor Red
        exit 1
    }
    $repoRoot = if ([string]::IsNullOrEmpty($PSScriptRoot)) { (Get-Location).Path } else { $PSScriptRoot }
    $base = if ($env:BASE_URL) { $env:BASE_URL } else { "http://127.0.0.1:8080" }
    Push-Location $repoRoot
    try {
        k6 run loadtest/k6/slots_list.js -e "BASE_URL=$base" -e "LOADTEST_PROFILE=smoke"
        if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
    }
    finally {
        Pop-Location
    }
}

function LoadTest-Docker {
    $repoRoot = if ([string]::IsNullOrEmpty($PSScriptRoot)) { (Get-Location).Path } else { $PSScriptRoot }
    $rate = if ($env:LOADTEST_RATE) { $env:LOADTEST_RATE } else { "50" }
    $dur = if ($env:LOADTEST_DURATION) { $env:LOADTEST_DURATION } else { "90s" }
    Write-Host "k6 (Docker) against host.docker.internal:8080 ..." -ForegroundColor Cyan
    docker run --rm --add-host=host.docker.internal:host-gateway `
        -v "${repoRoot}:/t" -w /t `
        -e BASE_URL=http://host.docker.internal:8080 `
        -e LOADTEST_RATE=$rate `
        -e LOADTEST_DURATION=$dur `
        grafana/k6:0.57.0 run loadtest/k6/slots_list.js
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
}

function Test-E2E {
    $repoRoot = if ([string]::IsNullOrEmpty($PSScriptRoot)) { (Get-Location).Path } else { $PSScriptRoot }
    $base = if ($env:E2E_BASE_URL) { $env:E2E_BASE_URL } else { "http://127.0.0.1:8080" }
    Write-Host "Running E2E tests against $base ..." -ForegroundColor Cyan
    $prevE2E = $env:E2E_BASE_URL
    Push-Location $repoRoot
    try {
        $env:E2E_BASE_URL = $base
        gotestsum --format testname -- -count=1 -v ./e2e/
        if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
    }
    finally {
        Pop-Location
        if ([string]::IsNullOrEmpty($prevE2E)) { Remove-Item Env:E2E_BASE_URL -ErrorAction SilentlyContinue }
        else { $env:E2E_BASE_URL = $prevE2E }
    }
}

function Test-Coverage {
    $repoRoot = if ([string]::IsNullOrEmpty($PSScriptRoot)) { (Get-Location).Path } else { $PSScriptRoot }
    $coverageOut = Join-Path $repoRoot "coverage.out"
    Write-Host "Running tests with coverage profile ($coverageOut)..." -ForegroundColor Cyan
    Push-Location $repoRoot
    try {
        # Space-separated flags: PowerShell can mishandle -flag=value for native go.exe (profile never written).
        go test ./... -count=1 -cover -coverprofile $coverageOut -covermode atomic
        if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
        Write-Host ""
        Write-Host "Per-file summary (last line = total):" -ForegroundColor Cyan
        go tool cover -func $coverageOut
        if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
    }
    finally {
        Pop-Location
    }
}

function Lint {
    if (-not (Get-Command golangci-lint -ErrorAction SilentlyContinue)) {
        Write-Host "Install golangci-lint: https://golangci-lint.run/welcome/install/" -ForegroundColor Red
        exit 1
    }
    $repoRoot = if ([string]::IsNullOrEmpty($PSScriptRoot)) { (Get-Location).Path } else { $PSScriptRoot }
    Write-Host "Running linter..." -ForegroundColor Cyan
    Push-Location $repoRoot
    try {
        golangci-lint run ./...
        if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
    }
    finally {
        Pop-Location
    }
}

function Lint-Fix {
    if (-not (Get-Command golangci-lint -ErrorAction SilentlyContinue)) {
        Write-Host "Install golangci-lint: https://golangci-lint.run/welcome/install/" -ForegroundColor Red
        exit 1
    }
    $repoRoot = if ([string]::IsNullOrEmpty($PSScriptRoot)) { (Get-Location).Path } else { $PSScriptRoot }
    Write-Host "Running linter with autofix..." -ForegroundColor Cyan
    Push-Location $repoRoot
    try {
        golangci-lint run --fix ./...
        if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
    }
    finally {
        Pop-Location
    }
}

function Swagger {
    Write-Host "Generating Swagger docs..." -ForegroundColor Cyan
    swag init --parseDependency --parseInternal -g ./cmd/app/main.go
    Write-Host "Swagger docs generated" -ForegroundColor Green
}

function Dev {
    Swagger
    Write-Host "Starting dev mode (Air hot reload)..." -ForegroundColor Cyan
    air
}

function Build {
    Swagger
    Write-Host "Building Go application..." -ForegroundColor Cyan
    if (-Not (Test-Path "bin")) {
        New-Item -ItemType Directory -Path "bin" | Out-Null
    }
    go build -o bin/app ./cmd/app
    Write-Host "Build complete: bin/app" -ForegroundColor Green
}

function Run {
    Swagger
    Write-Host "Running Go application..." -ForegroundColor Cyan
    go run ./cmd/app
}

function Help {
    Write-Host "Available commands:" -ForegroundColor Yellow
    Write-Host "  up             - Start project with dependencies (docker compose up -d --build)"
    Write-Host "  down           - Stop Docker Compose stack"
    Write-Host "  seed           - Fill DB via docker compose"
    Write-Host "  dev            - Start Air hot reload (with Swagger hot reload)"
    Write-Host "  build          - Build Go app into ./bin/app (Swagger updated first)"
    Write-Host "  run            - Run Go app (no reload)"
    Write-Host "  swag           - Generate Swagger docs"
    Write-Host "  test           - Run tests with gotestsum (with per-package -cover)"
    Write-Host "  test-e2e       - HTTP E2E (E2E_BASE_URL or http://127.0.0.1:8080)"
    Write-Host "  loadtest       - k6 load test (BASE_URL, LOADTEST_RATE, LOADTEST_DURATION)"
    Write-Host "  loadtest-smoke - Short k6 sanity run"
    Write-Host "  loadtest-docker - k6 via Docker + volume mount"
    Write-Host "  test-coverage  - Full merged coverage profile -> coverage.out + go tool cover -func"
    Write-Host "  lint           - Run golangci-lint (see .golangci.yaml)"
    Write-Host "  lint-fix       - Run golangci-lint with --fix"
    Write-Host "  help           - Show this help"
}

switch ($Command.ToLower()) {
    "up"    { Up }
    "down"  { Down }
    "seed"  { Seed }
    "dev"   { Dev }
    "build" { Build }
    "run"   { Run }
    "swag"  { Swagger }
    "test"           { Test-Project }
    "test-e2e"        { Test-E2E }
    "loadtest"        { LoadTest }
    "loadtest-smoke"  { LoadTest-Smoke }
    "loadtest-docker" { LoadTest-Docker }
    "test-coverage"   { Test-Coverage }
    "lint"            { Lint }
    "lint-fix"        { Lint-Fix }
    "help"            { Help }
    default {
        Write-Host "Unknown command: $Command" -ForegroundColor Red
        Help
    }
}
