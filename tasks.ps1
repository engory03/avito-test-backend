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
    gotestsum --format testname
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
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
    Write-Host "  up       - Start project with dependencies (docker compose up -d --build)"
    Write-Host "  down     - Stop Docker Compose stack"
    Write-Host "  seed     - Fill DB via docker compose"
    Write-Host "  dev      - Start Air hot reload (with Swagger hot reload)"
    Write-Host "  build    - Build Go app into ./bin/app (Swagger updated first)"
    Write-Host "  run      - Run Go app (no reload)"
    Write-Host "  swag     - Generate Swagger docs"
    Write-Host "  test     - Run tests with gotestsum"
    Write-Host "  help     - Show this help"
}

switch ($Command.ToLower()) {
    "up"    { Up }
    "down"  { Down }
    "seed"  { Seed }
    "dev"   { Dev }
    "build" { Build }
    "run"   { Run }
    "swag"  { Swagger }
    "test"  { Test-Project }
    "help"  { Help }
    default {
        Write-Host "Unknown command: $Command" -ForegroundColor Red
        Help
    }
}
