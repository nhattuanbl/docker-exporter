# Docker Exporter Build Script for Windows
# Usage: .\build.ps1 [build|release|clean|version]

param(
    [string]$Command = "build",
    [string]$Version = ""
)

$ErrorActionPreference = "Stop"

# Get version from git tag or parameter
if (-not $Version) {
    $Version = git describe --tags --always 2>$null
    if (-not $Version) { $Version = "dev" }
}

$GitCommit = git rev-parse --short HEAD 2>$null
if (-not $GitCommit) { $GitCommit = "unknown" }

$BuildDate = (Get-Date -Format "yyyy-MM-ddTHH:mm:ssZ")
$GoVersion = (go version) -replace "go version ", ""

# LDFLAGS
$LDFlags = "-s -w " +
    "-X github.com/nhattuanbl/docker-exporter/internal/config.Version=$Version " +
    "-X github.com/nhattuanbl/docker-exporter/internal/config.GitCommit=$GitCommit " +
    "-X github.com/nhattuanbl/docker-exporter/internal/config.BuildDate=$BuildDate " +
    "-X `"github.com/nhattuanbl/docker-exporter/internal/config.GoVersion=$GoVersion`""

$Binary = "docker-exporter"

# Platforms for cross-compilation
$Platforms = @(
    @{GOOS="linux"; GOARCH="amd64"},
    @{GOOS="linux"; GOARCH="arm64"},
    @{GOOS="darwin"; GOARCH="amd64"},
    @{GOOS="darwin"; GOARCH="arm64"},
    @{GOOS="windows"; GOARCH="amd64"}
)

function Build-Current {
    Write-Host "Building $Binary $Version..." -ForegroundColor Green
    Write-Host "  Git Commit: $GitCommit"
    Write-Host "  Build Date: $BuildDate"
    go build -ldflags $LDFlags -o "$Binary.exe" ./cmd/docker-exporter
    if ($LASTEXITCODE -eq 0) {
        Write-Host "Done! Binary: $Binary.exe" -ForegroundColor Green
    } else {
        Write-Host "Build failed!" -ForegroundColor Red
        exit 1
    }
}

function Build-Release {
    Write-Host "========================================" -ForegroundColor Cyan
    Write-Host " Building Docker Exporter $Version" -ForegroundColor Cyan
    Write-Host "========================================" -ForegroundColor Cyan
    Write-Host "  Git Commit: $GitCommit"
    Write-Host "  Build Date: $BuildDate"
    Write-Host "  Go Version: $GoVersion"
    Write-Host ""

    # Create dist directory
    if (Test-Path "dist") {
        Remove-Item -Path "dist" -Recurse -Force
    }
    New-Item -ItemType Directory -Path "dist" | Out-Null

    foreach ($platform in $Platforms) {
        $os = $platform.GOOS
        $arch = $platform.GOARCH
        
        # Include version in filename: docker-exporter_v1.0.0_linux_amd64
        $output = "dist/${Binary}_${Version}_${os}_${arch}"
        
        if ($os -eq "windows") {
            $output = "$output.exe"
        }

        Write-Host "  Building $output..." -ForegroundColor Yellow

        $env:GOOS = $os
        $env:GOARCH = $arch
        $env:CGO_ENABLED = "0"
        
        go build -ldflags $LDFlags -o $output ./cmd/docker-exporter
        
        if ($LASTEXITCODE -ne 0) {
            Write-Host "    FAILED!" -ForegroundColor Red
        } else {
            $size = (Get-Item $output).Length / 1MB
            Write-Host "    OK ($([math]::Round($size, 2)) MB)" -ForegroundColor Green
        }
    }

    # Reset environment
    Remove-Item Env:GOOS -ErrorAction SilentlyContinue
    Remove-Item Env:GOARCH -ErrorAction SilentlyContinue
    Remove-Item Env:CGO_ENABLED -ErrorAction SilentlyContinue

    Write-Host ""
    Write-Host "========================================" -ForegroundColor Cyan
    Write-Host " Build Complete! Files in dist/" -ForegroundColor Cyan
    Write-Host "========================================" -ForegroundColor Cyan
    Get-ChildItem dist/ | Format-Table Name, @{N='Size (MB)';E={[math]::Round($_.Length/1MB, 2)}}
}

function Clean {
    Write-Host "Cleaning..." -ForegroundColor Yellow
    Remove-Item -Path "$Binary.exe" -ErrorAction SilentlyContinue
    Remove-Item -Path "dist" -Recurse -ErrorAction SilentlyContinue
    Write-Host "Done!" -ForegroundColor Green
}

function Show-Version {
    Write-Host "Version:    $Version"
    Write-Host "Git Commit: $GitCommit"
    Write-Host "Build Date: $BuildDate"
    Write-Host "Go Version: $GoVersion"
}

function Show-Help {
    Write-Host "Docker Exporter Build Script" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "Usage: .\build.ps1 [command] [-Version <version>]"
    Write-Host ""
    Write-Host "Commands:"
    Write-Host "  build      Build for current platform (default)"
    Write-Host "  release    Build for all platforms with version in filename"
    Write-Host "  clean      Clean build artifacts"
    Write-Host "  version    Show version info"
    Write-Host ""
    Write-Host "Examples:"
    Write-Host "  .\build.ps1 build"
    Write-Host "  .\build.ps1 release -Version v1.0.0"
    Write-Host "  .\build.ps1 release  # Uses git tag as version"
}

# Main
switch ($Command.ToLower()) {
    "build" { Build-Current }
    "release" { Build-Release }
    "clean" { Clean }
    "version" { Show-Version }
    "help" { Show-Help }
    default { Show-Help }
}
