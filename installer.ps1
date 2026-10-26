# Jalin Desktop Application Installer & Uninstaller
param(
    [switch]$Install,
    [switch]$Uninstall
)

$appName = "Jalin"
$displayName = "Jalin - Network Intelligence & Speed Benchmark"
$version = "1.0.0"
$sourceDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$installDir = Join-Path $env:LOCALAPPDATA "Programs\Jalin"
$desktopLnk = Join-Path ([Environment]::GetFolderPath("Desktop")) "Jalin.lnk"
$startMenuDir = Join-Path ([Environment]::GetFolderPath("StartMenu")) "Programs"
$startMenuLnk = Join-Path $startMenuDir "Jalin.lnk"
$regKey = "HKCU:\Software\Microsoft\Windows\CurrentVersion\Uninstall\Jalin"

function Create-Shortcut {
    param($targetPath, $shortcutPath, $iconPath, $workDir)
    $wsh = New-Object -ComObject WScript.Shell
    $s = $wsh.CreateShortcut($shortcutPath)
    $s.TargetPath = $targetPath
    $s.WorkingDirectory = $workDir
    $s.IconLocation = "$iconPath,0"
    $s.Description = "Jalin Internet Speed Meter & Benchmark"
    $s.Save()
    [System.Runtime.InteropServices.Marshal]::ReleaseComObject($wsh) | Out-Null
}

if ($Uninstall) {
    Write-Host "[*] Stopping running Jalin instances..." -ForegroundColor Yellow
    Get-Process -Name jalin -ErrorAction SilentlyContinue | Stop-Process -Force -ErrorAction SilentlyContinue
    Start-Sleep -Milliseconds 800

    Write-Host "[*] Removing shortcuts..." -ForegroundColor Yellow
    if (Test-Path $desktopLnk) { Remove-Item $desktopLnk -Force -ErrorAction SilentlyContinue }
    if (Test-Path $startMenuLnk) { Remove-Item $startMenuLnk -Force -ErrorAction SilentlyContinue }

    Write-Host "[*] Removing registry uninstaller entry..." -ForegroundColor Yellow
    if (Test-Path $regKey) { Remove-Item $regKey -Recurse -Force -ErrorAction SilentlyContinue }

    Write-Host "[*] Removing program files..." -ForegroundColor Yellow
    if (Test-Path $installDir) { Remove-Item $installDir -Recurse -Force -ErrorAction SilentlyContinue }

    Write-Host "[OK] Jalin has been completely uninstalled from your system." -ForegroundColor Green
    exit 0
}

# Default action: Install
Write-Host "==========================================================" -ForegroundColor Cyan
Write-Host "         Installing Jalin v$version into Windows          " -ForegroundColor Cyan
Write-Host "==========================================================" -ForegroundColor Cyan

# 1. Stop active process if any
Get-Process -Name jalin -ErrorAction SilentlyContinue | Stop-Process -Force -ErrorAction SilentlyContinue

# 2. Check binary
$exeSource = Join-Path $sourceDir "jalin.exe"
if (-not (Test-Path $exeSource)) {
    Write-Host "[*] jalin.exe not found. Compiling binary..." -ForegroundColor Yellow
    Push-Location $sourceDir
    go build -trimpath -ldflags="-s -w -H=windowsgui" -o jalin.exe .
    Pop-Location
}

if (-not (Test-Path $exeSource)) {
    Write-Host "[ERROR] Could not find or build jalin.exe." -ForegroundColor Red
    exit 1
}

# 3. Create install directories
New-Item -ItemType Directory -Force -Path $installDir | Out-Null
$assetsTarget = Join-Path $installDir "assets"
New-Item -ItemType Directory -Force -Path $assetsTarget | Out-Null

# 4. Copy files
Write-Host "[*] Copying program files to $installDir..." -ForegroundColor Yellow
Copy-Item $exeSource -Destination (Join-Path $installDir "jalin.exe") -Force
$icoSource = Join-Path $sourceDir "assets\logo.ico"
if (Test-Path $icoSource) {
    Copy-Item $icoSource -Destination (Join-Path $assetsTarget "logo.ico") -Force
}
$pngSource = Join-Path $sourceDir "assets\icon.png"
if (Test-Path $pngSource) {
    Copy-Item $pngSource -Destination (Join-Path $assetsTarget "icon.png") -Force
}
Copy-Item $MyInvocation.MyCommand.Path -Destination (Join-Path $installDir "installer.ps1") -Force

# 5. Create shortcuts
$installedExe = Join-Path $installDir "jalin.exe"
$installedIco = Join-Path $assetsTarget "logo.ico"
if (-not (Test-Path $installedIco)) {
    $installedIco = $installedExe
}

Write-Host "[*] Creating Desktop and Start Menu shortcuts..." -ForegroundColor Yellow
Create-Shortcut -targetPath $installedExe -shortcutPath $desktopLnk -iconPath $installedIco -workDir $installDir
Create-Shortcut -targetPath $installedExe -shortcutPath $startMenuLnk -iconPath $installedIco -workDir $installDir

# 6. Register in Windows Settings / Control Panel (Add or Remove Programs)
Write-Host "[*] Registering application in Windows Apps..." -ForegroundColor Yellow
if (-not (Test-Path $regKey)) {
    New-Item -Path $regKey -Force | Out-Null
}
Set-ItemProperty -Path $regKey -Name "DisplayName" -Value $displayName -Force
Set-ItemProperty -Path $regKey -Name "DisplayVersion" -Value $version -Force
Set-ItemProperty -Path $regKey -Name "Publisher" -Value "Jalin Project" -Force
Set-ItemProperty -Path $regKey -Name "DisplayIcon" -Value "$installedIco,0" -Force
Set-ItemProperty -Path $regKey -Name "InstallLocation" -Value $installDir -Force
$uninstallerCmd = "powershell.exe -NoProfile -ExecutionPolicy Bypass -File `"$installDir\installer.ps1`" -Uninstall"
Set-ItemProperty -Path $regKey -Name "UninstallString" -Value $uninstallerCmd -Force
Set-ItemProperty -Path $regKey -Name "QuietUninstallString" -Value $uninstallerCmd -Force
Set-ItemProperty -Path $regKey -Name "NoModify" -Value 1 -Type DWord -Force
Set-ItemProperty -Path $regKey -Name "NoRepair" -Value 1 -Type DWord -Force

Write-Host "==========================================================" -ForegroundColor Green
Write-Host " [OK] Jalin v$version has been successfully installed!" -ForegroundColor Green
Write-Host " Installed Path: $installDir\jalin.exe" -ForegroundColor White
Write-Host " Desktop and Start Menu shortcuts have been created." -ForegroundColor White
Write-Host "==========================================================" -ForegroundColor Green
