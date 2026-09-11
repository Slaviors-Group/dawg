# build-bundle.ps1 — Assembles the self-contained monolithic DAWG desktop bundle (Strategy A)
[CmdletBinding()]
param(
    [switch]$SkipDownload,
    [switch]$SkipTauri,
    [string]$MitmVersion = "12.2.3",
    [string]$NodeVersion = "22.14.0"
)

$ErrorActionPreference = "Stop"

Write-Host "========================================================" -ForegroundColor Cyan
Write-Host "  DAWG Desktop Monolithic Assembly Pipeline (Strategy A)" -ForegroundColor Cyan
Write-Host "========================================================" -ForegroundColor Cyan

$desktopDir = $PSScriptRoot
$repoRoot = Split-Path -Parent $desktopDir
$engineDir = Join-Path $repoRoot "engine"
$schemaSourceDir = Join-Path $repoRoot "schema"
$extensionSourceDir = Join-Path $repoRoot "extension"
$cacheDir = Join-Path $desktopDir ".cache"
$resourcesDir = Join-Path $desktopDir "src-tauri\resources"

# Destination directories inside resources
$binariesDir = Join-Path $resourcesDir "binaries"
$mitmTargetDir = Join-Path $binariesDir "mitmdump"
$nodeTargetDir = Join-Path $binariesDir "node"
$scriptsTargetDir = Join-Path $resourcesDir "scripts"
$schemaTargetDir = Join-Path $resourcesDir "schema"
$browsersTargetDir = Join-Path $resourcesDir "browsers"
$extensionTargetDir = Join-Path $resourcesDir "extension"

# Ensure clean directory structure
@($cacheDir, $resourcesDir, $binariesDir, $mitmTargetDir, $nodeTargetDir, $scriptsTargetDir, $schemaTargetDir, $browsersTargetDir) | ForEach-Object {
    if (-Not (Test-Path $_)) {
        New-Item -ItemType Directory -Force -Path $_ | Out-Null
    }
}

# Do not package stale Unix executables when this checkout has previously been
# used to assemble a Linux bundle.
@(
    (Join-Path $binariesDir "dawg"),
    (Join-Path $mitmTargetDir "mitmdump"),
    (Join-Path $nodeTargetDir "node")
) | ForEach-Object {
    if (Test-Path $_ -PathType Leaf) {
        Remove-Item -Force $_
    }
}

# -------------------------------------------------------------------
# 1. Compile DAWG Engine Binary
# -------------------------------------------------------------------
Write-Host "`n[1/7] Compiling DAWG Go Engine..." -ForegroundColor Yellow
Set-Location -Path $engineDir
$engineTargetExe = Join-Path $binariesDir "dawg.exe"
go build -trimpath -ldflags "-s -w" -o $engineTargetExe ./cmd/dawg
if ($LASTEXITCODE -ne 0) {
    Write-Error "Go engine build failed!"
}
Write-Host "  -> Engine built at $engineTargetExe" -ForegroundColor Green

# -------------------------------------------------------------------
# 2. Stage Standalone mitmdump
# -------------------------------------------------------------------
Write-Host "`n[2/7] Staging Standalone mitmdump (v$MitmVersion)..." -ForegroundColor Yellow
$mitmExe = Join-Path $mitmTargetDir "mitmdump.exe"

if (-Not (Test-Path $mitmExe) -or -Not $SkipDownload) {
    # Check if local mitmdump exists on machine first to avoid unnecessary download
    $localMitmdump = Get-Command mitmdump.exe -ErrorAction SilentlyContinue
    $mitmZip = Join-Path $cacheDir "mitmproxy-$MitmVersion.zip"

    if (-Not (Test-Path $mitmZip) -and -Not $SkipDownload) {
        $mitmUrl = "https://downloads.mitmproxy.org/$MitmVersion/mitmproxy-$MitmVersion-windows-x86_64.zip"
        Write-Host "  Downloading $mitmUrl..." -ForegroundColor DarkGray
        try {
            Invoke-WebRequest -Uri $mitmUrl -OutFile $mitmZip -UseBasicParsing
        } catch {
            # Fallback to github releases
            $mitmUrl = "https://github.com/mitmproxy/mitmproxy/releases/download/v$MitmVersion/mitmproxy-$MitmVersion-windows-x86_64.zip"
            Write-Host "  Retrying via GitHub Releases: $mitmUrl..." -ForegroundColor DarkGray
            Invoke-WebRequest -Uri $mitmUrl -OutFile $mitmZip -UseBasicParsing
        }
    }

    if (Test-Path $mitmZip) {
        Write-Host "  Extracting mitmdump from cache..." -ForegroundColor DarkGray
        $tempExtract = Join-Path $cacheDir "mitm_extract"
        if (Test-Path $tempExtract) { Remove-Item -Recurse -Force $tempExtract }
        Expand-Archive -Path $mitmZip -DestinationPath $tempExtract -Force
        Copy-Item -Path (Join-Path $tempExtract "mitmdump.exe") -Destination $mitmExe -Force
        Remove-Item -Recurse -Force $tempExtract
    } elseif ($localMitmdump) {
        Write-Host "  Copying local mitmdump binary from $($localMitmdump.Source)..." -ForegroundColor DarkGray
        Copy-Item -Path $localMitmdump.Source -Destination $mitmExe -Force
    }
}

if (-Not (Test-Path $mitmExe)) {
    Write-Error "Failed to stage mitmdump.exe!"
}
Write-Host "  -> mitmdump staged at $mitmExe" -ForegroundColor Green

# -------------------------------------------------------------------
# 3. Stage Portable Node.js Runtime
# -------------------------------------------------------------------
Write-Host "`n[3/7] Staging Portable Node.js (v$NodeVersion)..." -ForegroundColor Yellow
$nodeExe = Join-Path $nodeTargetDir "node.exe"

if (-Not (Test-Path $nodeExe) -or -Not $SkipDownload) {
    $localNode = Get-Command node.exe -ErrorAction SilentlyContinue
    $nodeZip = Join-Path $cacheDir "node-v$NodeVersion-win-x64.zip"

    if (-Not (Test-Path $nodeZip) -and -Not $SkipDownload) {
        $nodeUrl = "https://nodejs.org/dist/v$NodeVersion/node-v$NodeVersion-win-x64.zip"
        Write-Host "  Downloading $nodeUrl..." -ForegroundColor DarkGray
        Invoke-WebRequest -Uri $nodeUrl -OutFile $nodeZip -UseBasicParsing
    }

    if (Test-Path $nodeZip) {
        Write-Host "  Extracting node.exe from cache..." -ForegroundColor DarkGray
        $tempExtract = Join-Path $cacheDir "node_extract"
        if (Test-Path $tempExtract) { Remove-Item -Recurse -Force $tempExtract }
        Expand-Archive -Path $nodeZip -DestinationPath $tempExtract -Force
        $extractedNode = Get-ChildItem -Path $tempExtract -Recurse -Filter "node.exe" | Select-Object -First 1
        if ($extractedNode) {
            Copy-Item -Path $extractedNode.FullName -Destination $nodeExe -Force
        }
        Remove-Item -Recurse -Force $tempExtract
    } elseif ($localNode) {
        Write-Host "  Copying local node.exe binary from $($localNode.Source)..." -ForegroundColor DarkGray
        Copy-Item -Path $localNode.Source -Destination $nodeExe -Force
    }
}

if (-Not (Test-Path $nodeExe)) {
    Write-Error "Failed to stage node.exe!"
}
Write-Host "  -> Node.js staged at $nodeExe" -ForegroundColor Green

# -------------------------------------------------------------------
# 4. Provision Playwright & Node Modules for replay
# -------------------------------------------------------------------
Write-Host "`n[4/7] Provisioning Playwright replay dependencies..." -ForegroundColor Yellow
$engineNodeModules = Join-Path $engineDir "node_modules"
$resourcesNodeModules = Join-Path $resourcesDir "node_modules"
$enginePackageJson = Join-Path $engineDir "package.json"
$enginePackageLock = Join-Path $engineDir "package-lock.json"
$resourcesPackageJson = Join-Path $resourcesDir "package.json"
$resourcesPackageLock = Join-Path $resourcesDir "package-lock.json"

if (Test-Path $engineNodeModules) {
    Write-Host "  Copying engine node_modules to bundle resources..." -ForegroundColor DarkGray
    if (Test-Path $resourcesNodeModules) {
        Remove-Item -Recurse -Force $resourcesNodeModules
    }
    Copy-Item -Path $engineNodeModules -Destination $resourcesNodeModules -Recurse -Force
    Copy-Item -Path $enginePackageJson -Destination $resourcesPackageJson -Force
    Copy-Item -Path $enginePackageLock -Destination $resourcesPackageLock -Force
} elseif ($SkipDownload) {
    $playwrightCli = Join-Path $resourcesNodeModules "playwright\cli.js"
    $lockIsCurrent = (Test-Path $resourcesPackageLock) -and ((Get-FileHash $enginePackageLock).Hash -eq (Get-FileHash $resourcesPackageLock).Hash)
    if (-Not (Test-Path $playwrightCli) -or -Not $lockIsCurrent) {
        Write-Error "Cached Node dependencies are missing or stale; rerun without -SkipDownload."
    }
    npm --prefix $resourcesDir ls --omit=dev | Out-Null
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Cached Node dependency validation failed; rerun without -SkipDownload."
    }
} else {
    Write-Host "  Installing npm dependencies in bundle resources..." -ForegroundColor DarkGray
    Copy-Item -Path $enginePackageJson -Destination $resourcesPackageJson -Force
    Copy-Item -Path $enginePackageLock -Destination $resourcesPackageLock -Force
    Set-Location -Path $resourcesDir
    npm ci --omit=dev
    if ($LASTEXITCODE -ne 0) {
        Write-Error "npm dependency installation failed!"
    }
}
Write-Host "  -> Node modules staged at $resourcesNodeModules" -ForegroundColor Green

# -------------------------------------------------------------------
# 5. Provision Chromium
# -------------------------------------------------------------------
Write-Host "`n[5/7] Provisioning bundled Chromium..." -ForegroundColor Yellow
$playwrightCli = Join-Path $resourcesNodeModules "playwright\cli.js"
if (-Not (Test-Path $playwrightCli)) {
    Write-Error "Playwright CLI is missing at $playwrightCli"
}

$env:PLAYWRIGHT_BROWSERS_PATH = $browsersTargetDir
if (-Not $SkipDownload) {
    # A shared/dual-boot checkout may contain a Chromium build for another OS.
    # Recreate the directory so the NSIS installer contains Windows assets only.
    if (Test-Path $browsersTargetDir) {
        Remove-Item -Recurse -Force $browsersTargetDir
    }
    New-Item -ItemType Directory -Force -Path $browsersTargetDir | Out-Null
    & $nodeExe $playwrightCli install chromium --no-shell
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Playwright Chromium installation failed!"
    }
}

# DAWG passes the full Chromium executable path and does not record video, so
# neither the separate headless shell nor ffmpeg is required at runtime.
Get-ChildItem -Path $browsersTargetDir -Directory -ErrorAction SilentlyContinue |
    Where-Object { $_.Name -like "chromium_headless_shell-*" -or $_.Name -like "ffmpeg-*" } |
    Remove-Item -Recurse -Force

$chromiumExe = Get-ChildItem -Path $browsersTargetDir -Recurse -File -Filter "chrome.exe" -ErrorAction SilentlyContinue |
    Where-Object { $_.FullName -match '[\\/]chromium-[^\\/]+[\\/]chrome-win(64)?[\\/]chrome\.exe$' } |
    Select-Object -First 1
if (-Not $chromiumExe) {
    if ($SkipDownload) {
        Write-Error "Bundled Chromium is missing; rerun without -SkipDownload."
    }
    Write-Error "Playwright completed without staging a supported Chromium executable."
}
Write-Host "  -> Chromium staged at $($chromiumExe.FullName)" -ForegroundColor Green

# -------------------------------------------------------------------
# 6. Copy Engine Scripts, Schema & Browser Extension
# -------------------------------------------------------------------
Write-Host "`n[6/7] Copying engine scripts, JSON schema/policies, and browser extension..." -ForegroundColor Yellow
@($scriptsTargetDir, $schemaTargetDir, $extensionTargetDir) | ForEach-Object {
    if (Test-Path $_) {
        Remove-Item -Recurse -Force $_
    }
    New-Item -ItemType Directory -Force -Path $_ | Out-Null
}
Copy-Item -Path "$engineDir\scripts\replay-browser.cjs" -Destination $scriptsTargetDir -Force
Copy-Item -Path "$schemaSourceDir\*" -Destination $schemaTargetDir -Recurse -Force
Copy-Item -Path "$extensionSourceDir\*" -Destination $extensionTargetDir -Recurse -Force
$extensionTestsDir = Join-Path $extensionTargetDir "tests"
if (Test-Path $extensionTestsDir) {
    Remove-Item -Recurse -Force $extensionTestsDir
}
Write-Host "  -> Scripts, schema, and installable extension staged into $resourcesDir" -ForegroundColor Green

# -------------------------------------------------------------------
# 7. Verify Staged Bundle via DAWG Doctor
# -------------------------------------------------------------------
Write-Host "`n[7/7] Verifying Staged Monolithic Bundle with 'dawg doctor'..." -ForegroundColor Yellow
$env:DAWG_RESOURCES_DIR = $resourcesDir
& $engineTargetExe doctor
if ($LASTEXITCODE -ne 0) {
    Write-Error "DAWG doctor reported degraded status; refusing to package an incomplete bundle."
} else {
    Write-Host "  Staged bundle passed all health checks! 🎉" -ForegroundColor Green
}

# -------------------------------------------------------------------
# Final: Build Tauri Installer (if requested)
# -------------------------------------------------------------------
if (-Not $SkipTauri) {
    Write-Host "`nBuilding Tauri Desktop Distribution Package..." -ForegroundColor Cyan
    Set-Location -Path $desktopDir
    $generatedResourcesDir = Join-Path $desktopDir "src-tauri\target\release\resources"
    if (Test-Path $generatedResourcesDir) {
        Remove-Item -Recurse -Force $generatedResourcesDir
    }
    # Invoke the local tauri CLI binary directly rather than via 'npm run',
    # since npm's argument passthrough on Windows can mangle the
    # '--bundles nsis' flag (it gets silently dropped, leaving a bare
    # 'nsis' positional argument that the Rust CLI rejects).
    & ".\node_modules\.bin\tauri.cmd" build --bundles nsis
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Tauri packaging build failed!"
    }
    Write-Host "`n🎉 Monolithic DAWG Desktop installer created successfully!" -ForegroundColor Green
} else {
    Write-Host "`nBundle staging complete (Tauri build skipped by flag)." -ForegroundColor Cyan
    Write-Host "Run '.\node_modules\.bin\tauri.cmd build --bundles nsis' inside 'desktop/' to compile the Windows installer." -ForegroundColor DarkGray
}
