param(
    [string]$Repo = $(if ($env:BOLTZ_API_REPO) { $env:BOLTZ_API_REPO } else { "boltz-bio/boltz-compute-api-cli" }),
    [string]$Version = $(if ($env:BOLTZ_API_VERSION) { $env:BOLTZ_API_VERSION } else { "latest" }),
    [string]$InstallDir = $env:BOLTZ_API_INSTALL_DIR
)

$ErrorActionPreference = "Stop"

function Fail($Message) {
    Write-Error $Message
    exit 1
}

$arch = switch ([System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture) {
    "X64" { "amd64" }
    "X86" { "386" }
    "Arm64" { "arm64" }
    default { Fail "Unsupported CPU architecture: $([System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture)" }
}

if ($Version -eq "latest") {
    $releaseUrl = "https://api.github.com/repos/$Repo/releases/latest"
} else {
    if ($Version.StartsWith("v")) {
        $tag = $Version
    } else {
        $tag = "v$Version"
    }
    $releaseUrl = "https://api.github.com/repos/$Repo/releases/tags/$tag"
}

$release = Invoke-RestMethod -Uri $releaseUrl -Headers @{ Accept = "application/vnd.github+json" }
if (-not $release.tag_name) {
    Fail "Could not determine the boltz-api release tag"
}

$versionNumber = $release.tag_name -replace "^v", ""

$asset = $release.assets |
    Where-Object { $_.name -match "^boltz-api_.*_windows_${arch}\.zip$" } |
    Select-Object -First 1

if (-not $asset) {
    Fail "No boltz-api release asset found for windows/$arch in $($release.tag_name)"
}

if (-not $InstallDir) {
    $existing = Get-Command boltz-api -ErrorAction SilentlyContinue
    if ($existing -and $existing.Source) {
        $existingBinary = $existing.Source
        $InstallDir = Split-Path -Parent $existing.Source
    } elseif ($env:LOCALAPPDATA) {
        $existingBinary = $null
        $InstallDir = Join-Path $env:LOCALAPPDATA "Programs\Boltz\bin"
    } else {
        $existingBinary = $null
        $InstallDir = Join-Path $HOME ".local\bin"
    }
} else {
    $existingBinary = Join-Path $InstallDir "boltz-api.exe"
}

if ($existingBinary -and (Test-Path $existingBinary)) {
    $currentOutput = & $existingBinary --version 2>$null
    if ($currentOutput -match "([0-9]+\.[0-9][^ ]*)") {
        if ($Matches[1] -eq $versionNumber) {
            Write-Host "boltz-api $($release.tag_name) is already installed at $existingBinary"
            exit 0
        }
    }
}

$tmp = Join-Path ([System.IO.Path]::GetTempPath()) ("boltz-api-" + [System.Guid]::NewGuid())
New-Item -ItemType Directory -Path $tmp | Out-Null

try {
    $archive = Join-Path $tmp "boltz-api.zip"
    Invoke-WebRequest -Uri $asset.browser_download_url -OutFile $archive
    Expand-Archive -Path $archive -DestinationPath $tmp -Force

    $binary = Get-ChildItem -Path $tmp -Filter "boltz-api.exe" -Recurse | Select-Object -First 1
    if (-not $binary) {
        Fail "Downloaded archive did not contain boltz-api.exe"
    }

    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    Copy-Item -Path $binary.FullName -Destination (Join-Path $InstallDir "boltz-api.exe") -Force
} finally {
    Remove-Item -Path $tmp -Recurse -Force -ErrorAction SilentlyContinue
}

Write-Host "Installed boltz-api $($release.tag_name) to $(Join-Path $InstallDir "boltz-api.exe")"

$pathEntries = $env:Path -split ";"
if ($pathEntries -notcontains $InstallDir) {
    Write-Warning "Add $InstallDir to PATH to run boltz-api without the full path."
}
