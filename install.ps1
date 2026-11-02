# ==============================================================================
# CodeDNA — Universal Zero-Dependency Installer for Windows (PowerShell)
# "Turn Any Codebase Into a Living Architecture Map"
#
# Usage:
#   irm https://raw.githubusercontent.com/0xzahed/codedna/main/install.ps1 | iex
#   or: powershell -ExecutionPolicy Bypass -File install.ps1 [-Uninstall]
#
# Repository: https://github.com/0xzahed/codedna
# Author: Abu Zahed <zahed04x@gmail.com>
# ==============================================================================

[CmdletBinding()]
param(
    [switch]$Uninstall,
    [switch]$Force
)

$ErrorActionPreference = "Stop"

$RepoOwner = "0xzahed"
$RepoName = "codedna"
$GitHubRepo = "https://github.com/$RepoOwner/$RepoName"
$InstallDir = Join-Path $HOME ".codedna"
$BinDir = Join-Path $InstallDir "bin"
$WebDir = Join-Path $InstallDir "web\out"
$ExePath = Join-Path $BinDir "codedna.exe"

# Handle Uninstallation
if ($Uninstall) {
    Write-Host "[!] Uninstalling CodeDNA..." -ForegroundColor Yellow
    if (Test-Path $InstallDir) {
        Remove-Item -Recurse -Force $InstallDir
    }
    
    # Remove from User PATH
    $UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if ($UserPath -like "*$BinDir*") {
        $CleanPath = ($UserPath -split ';' | Where-Object { $_ -ne $BinDir }) -join ';'
        [Environment]::SetEnvironmentVariable("Path", $CleanPath, "User")
        Write-Host "[✓] Removed $BinDir from User PATH environment variable." -ForegroundColor Green
    }
    Write-Host "[✓] CodeDNA successfully uninstalled." -ForegroundColor Green
    exit 0
}

Write-Host @"
   ____          _      ____  _   _    _      ____     ___  
  / ___|___   __| | ___|  _ \| \ | |  / \    |___ \   / _ \ 
 | |   / _ \ / _| |/ _ \ | | |  \| | / _ \     __) | | | | |
 | |__| (_) | (_| |  __/ |_| | |\  |/ ___ \   / __/ _| |_| |
  \____\___/ \__,_|\___|____/|_| \_/_/   \_\ |_____(_)\___/ 

  CodeDNA — Universal Software Architecture Engine
  Native Windows Installation
====================================================================
"@ -ForegroundColor Cyan

# Detect Architecture
$Arch = "amd64"
if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") {
    $Arch = "arm64"
}
Write-Host "[CODEDNA] Target platform: windows/$Arch" -ForegroundColor Cyan

# Create directories
New-Item -ItemType Directory -Force -Path $BinDir | Out-Null
New-Item -ItemType Directory -Force -Path $WebDir | Out-Null

$Installed = $false

# ------------------------------------------------------------------------------
# Priority 1: Download Prebuilt Release Zip from GitHub
# ------------------------------------------------------------------------------
try {
    Write-Host "[CODEDNA] Checking latest GitHub release..." -ForegroundColor Cyan
    $ReleaseApi = "https://api.github.com/repos/$RepoOwner/$RepoName/releases/latest"
    $Release = Invoke-RestMethod -Uri $ReleaseApi -Method Get -Headers @{ "Accept" = "application/vnd.github.v3+json" }
    $TagName = $Release.tag_name

    if ($TagName) {
        Write-Host "[CODEDNA] Found release: $TagName" -ForegroundColor Cyan
        $ZipName = "codedna-$TagName-windows-$Arch.zip"
        $FallbackName = "codedna-windows-$Arch.zip"
        $DownloadUrl = "https://github.com/$RepoOwner/$RepoName/releases/download/$TagName/$ZipName"
        $TempZip = Join-Path $env:TEMP $ZipName

        Write-Host "[CODEDNA] Downloading $ZipName..." -ForegroundColor Cyan
        try {
            Invoke-WebRequest -Uri $DownloadUrl -OutFile $TempZip -UseBasicParsing
        } catch {
            $DownloadUrl = "https://github.com/$RepoOwner/$RepoName/releases/download/$TagName/$FallbackName"
            Invoke-WebRequest -Uri $DownloadUrl -OutFile $TempZip -UseBasicParsing
        }

        # SHA256 Verification
        try {
            $ChecksumUrl = "https://github.com/$RepoOwner/$RepoName/releases/download/$TagName/SHA256SUMS"
            $TempChecksum = Join-Path $env:TEMP "SHA256SUMS"
            Invoke-WebRequest -Uri $ChecksumUrl -OutFile $TempChecksum -UseBasicParsing
            $ChecksumContent = Get-Content $TempChecksum
            $MatchedLine = $ChecksumContent | Where-Object { $_ -match $ZipName -or $_ -match $FallbackName }
            if ($MatchedLine) {
                $ExpectedSha = ($MatchedLine -split '\s+')[0].Trim()
                $ActualSha = (Get-FileHash -Path $TempZip -Algorithm SHA256).Hash.ToLower()
                if ($ExpectedSha.ToLower() -eq $ActualSha) {
                    Write-Host "[✓] SHA256 checksum verified ($($ActualSha.Substring(0, 16))...)" -ForegroundColor Green
                } else {
                    Write-Warning "Checksum mismatch. Expected: $ExpectedSha, Got: $ActualSha"
                }
            }
        } catch {
            Write-Host "[!] Checksum verification skipped (manifest not available)." -ForegroundColor Yellow
        }

        # Extract
        Write-Host "[CODEDNA] Extracting archive..." -ForegroundColor Cyan
        $TempExtract = Join-Path $env:TEMP "codedna-extracted"
        if (Test-Path $TempExtract) { Remove-Item -Recurse -Force $TempExtract }
        Expand-Archive -Path $TempZip -DestinationPath $TempExtract -Force

        $FoundExe = Get-ChildItem -Path $TempExtract -Filter "codedna.exe" -Recurse | Select-Object -First 1
        if ($FoundExe) {
            Copy-Item -Path $FoundExe.FullName -Destination $ExePath -Force
            Write-Host "[✓] Installed codedna.exe to $ExePath" -ForegroundColor Green

            # Extract web assets if bundled
            $FoundWeb = Get-ChildItem -Path $TempExtract -Directory -Filter "out" -Recurse | Select-Object -First 1
            if ($FoundWeb) {
                Copy-Item -Path (Join-Path $FoundWeb.FullName "*") -Destination $WebDir -Recurse -Force
                Write-Host "[✓] Installed web studio assets to $WebDir" -ForegroundColor Green
            }
            $Installed = $true
        }
    }
} catch {
    Write-Host "[!] Prebuilt binary release download skipped ($($_.Exception.Message))" -ForegroundColor Yellow
}

# ------------------------------------------------------------------------------
# Priority 2: Build Natively from Source with Go (if Go is installed)
# ------------------------------------------------------------------------------
if (-not $Installed -and (Get-Command go -ErrorAction SilentlyContinue)) {
    Write-Host "[CODEDNA] Go toolchain detected. Compiling natively from source..." -ForegroundColor Cyan
    $SrcDir = Join-Path $InstallDir "src"
    New-Item -ItemType Directory -Force -Path $SrcDir | Out-Null

    if (Test-Path (Join-Path $SrcDir ".git")) {
        git -C $SrcDir pull --quiet origin main
    } elseif (Get-Command git -ErrorAction SilentlyContinue) {
        git clone --depth=1 --quiet "$GitHubRepo.git" $SrcDir
    } else {
        $SourceZip = Join-Path $env:TEMP "codedna-source.zip"
        Invoke-WebRequest -Uri "$GitHubRepo/archive/refs/heads/main.zip" -OutFile $SourceZip -UseBasicParsing
        Expand-Archive -Path $SourceZip -DestinationPath $SrcDir -Force
        $InnerDir = Get-ChildItem -Path $SrcDir -Directory | Select-Object -First 1
        Move-Item -Path (Join-Path $InnerDir.FullName "*") -Destination $SrcDir -Force
    }

    Push-Location $SrcDir
    try {
        go build -ldflags="-s -w" -o $ExePath ./cmd/codedna
        Write-Host "[✓] Compiled native binary: $ExePath" -ForegroundColor Green
        $Installed = $true
    } finally {
        Pop-Location
    }
}

# ------------------------------------------------------------------------------
# Configure PATH Environment Variable
# ------------------------------------------------------------------------------
if ($Installed) {
    $UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if ($UserPath -notlike "*$BinDir*") {
        $NewPath = if ($UserPath) { "$UserPath;$BinDir" } else { $BinDir }
        [Environment]::SetEnvironmentVariable("Path", $NewPath, "User")
        $env:Path = "$env:Path;$BinDir"
        Write-Host "[✓] Added $BinDir to User PATH." -ForegroundColor Green
    }

    Write-Host @"
====================================================================
[✓] CodeDNA successfully installed!
====================================================================

QUICK START:
  1. Open a new PowerShell terminal or restart your current terminal.
  2. Navigate into ANY code project directory:
     cd C:\path\to\my-project

  3. Start CodeDNA Architecture Studio:
     codedna serve

  4. Open your browser:
     http://127.0.0.1:8080

Tip: Run diagnostics with: codedna doctor
"@ -ForegroundColor Green
} else {
    Write-Error "CodeDNA installation could not be completed automatically. Please install Go (https://go.dev) or download release binaries from $GitHubRepo/releases."
}
