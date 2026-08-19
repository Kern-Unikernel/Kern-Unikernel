# ============================================================
#  kern — Windows Installer (via WSL2)
#
#  If you get an execution policy error, first run:
#  Set-ExecutionPolicy Bypass -Scope Process -Force
# ============================================================

$ErrorActionPreference = "Stop"

function Write-Step($msg) {
    Write-Host ""
    Write-Host "==> $msg" -ForegroundColor Cyan
}
function Write-OK($msg)   { Write-Host "  [OK] $msg" -ForegroundColor Green }
function Write-Fail($msg) { Write-Host "  [ERROR] $msg" -ForegroundColor Red; exit 1 }

# ── 0. Administrator check ───────────────────────────────────────────────
Write-Step "Checking permissions"
$isAdmin = ([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]"Administrator")
if (-NOT $isAdmin) {
    Write-Fail "Run PowerShell as Administrator (right click -> Run as administrator)"
}
Write-OK "Running as administrator"

# ── 1. Windows version check ─────────────────────────────────────────────
Write-Step "Checking Windows version"
$build = [System.Environment]::OSVersion.Version.Build
if ($build -lt 19041) {
    Write-Fail "You need Windows 10 build 19041 or higher. Your build: $build"
}
Write-OK "Windows build $build is compatible with WSL2"

# ── 2. Enable WSL2 ───────────────────────────────────────────────────────
Write-Step "Enabling WSL2"
$wslState = (Get-WindowsOptionalFeature -Online -FeatureName Microsoft-Windows-Subsystem-Linux).State
$vmState  = (Get-WindowsOptionalFeature -Online -FeatureName VirtualMachinePlatform).State
$needReboot = $false

if ($wslState -ne "Enabled") {
    Enable-WindowsOptionalFeature -Online -FeatureName Microsoft-Windows-Subsystem-Linux -NoRestart | Out-Null
    $needReboot = $true
}
if ($vmState -ne "Enabled") {
    Enable-WindowsOptionalFeature -Online -FeatureName VirtualMachinePlatform -NoRestart | Out-Null
    $needReboot = $true
}

if ($needReboot) {
    Write-Host ""
    Write-Host "  REBOOT REQUIRED. Run this script again after restarting." -ForegroundColor Yellow
    Write-Host "  Press any key to reboot now..."
    $null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
    Restart-Computer -Force
    exit
}
Write-OK "WSL2 enabled"

# ── 3. WSL2 kernel update ────────────────────────────────────────────────
# wsl --update may emit "WSL is finishing an upgrade..." on stderr
# even when it succeeds. Lower ErrorActionPreference for this block only.
Write-Step "Updating WSL2 kernel"
$ErrorActionPreference = "Continue"
wsl --update 2>&1 | Out-Null
Start-Sleep -Seconds 3
wsl --set-default-version 2 2>&1 | Out-Null
$ErrorActionPreference = "Stop"
Write-OK "WSL2 kernel updated"

# ── 4. Install Ubuntu if no distro present ───────────────────────────────
Write-Step "Checking Linux distribution"
$distros = wsl --list --quiet 2>&1
$hasUbuntu = $distros | Where-Object { $_ -match "Ubuntu" }

if (-not $hasUbuntu) {
    Write-Host "  Installing Ubuntu 22.04 (this may take a few minutes)..."
    wsl --install -d Ubuntu-22.04 --no-launch 2>&1 | Out-Null
    Write-OK "Ubuntu 22.04 installed"
    Write-Host ""
    Write-Host "  Set up your Linux username and password when prompted." -ForegroundColor Yellow
    wsl -d Ubuntu-22.04
} else {
    Write-OK "Ubuntu already available"
}

# ── 5. Run setup-wsl.sh inside WSL ───────────────────────────────────────
Write-Step "Installing Go, QEMU, kraft and kern inside WSL2"

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$setupScript = Join-Path $scriptDir "setup-wsl.sh"

if (-not (Test-Path $setupScript)) {
    Write-Fail "setup-wsl.sh not found. Make sure both files are in the same folder."
}

$wslPath = wsl wslpath -a ($setupScript -replace '\\', '/')
wsl -d Ubuntu-22.04 -- bash $wslPath

# ── 6. Desktop shortcut ──────────────────────────────────────────────────
Write-Step "Creating desktop shortcut"
$shell = New-Object -ComObject WScript.Shell
$shortcut = $shell.CreateShortcut("$env:USERPROFILE\Desktop\kern (WSL2).lnk")
$shortcut.TargetPath = "wsl.exe"
$shortcut.Arguments = "-d Ubuntu-22.04"
$shortcut.Description = "WSL2 terminal with kern"
$shortcut.Save()
Write-OK "Shortcut created on desktop"

# ── 7. Summary ───────────────────────────────────────────────────────────
Write-Host ""
Write-Host "============================================" -ForegroundColor Green
Write-Host "  Installation complete" -ForegroundColor Green
Write-Host "============================================" -ForegroundColor Green
Write-Host ""
Write-Host "  Open 'kern (WSL2)' from your desktop and run: kern --help"
Write-Host ""
