param(
    [Parameter(Position=0)]
    [string]$Command = "help"
)

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
    Write-Host "  dev      - Start Air hot reload (with Swagger hot reload)"
    Write-Host "  build    - Build Go app into ./bin/app (Swagger updated first)"
    Write-Host "  run      - Run Go app (no reload)"
    Write-Host "  help     - Show this help"
}

switch ($Command.ToLower()) {
    "dev"   { Dev }
    "build" { Build }
    "run"   { Run }
    "swag"  { Swagger }
    "help"  { Help }
    default {
        Write-Host "Unknown command: $Command" -ForegroundColor Red
        Help
    }
}
