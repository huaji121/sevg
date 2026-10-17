$ErrorActionPreference = 'Stop'
$goPath = $null
$cmd = Get-Command go -ErrorAction SilentlyContinue
if ($cmd) {
    $goPath = $cmd.Source
} else {
    foreach ($p in @('C:\Program Files\Go\bin\go.exe', "$env:USERPROFILE\go\bin\go.exe", 'E:\my_apps\Go\bin\go.exe')) {
        if (Test-Path $p) { $goPath = $p; break }
    }
}
if (-not $goPath) { throw 'go was not found. Install Go 1.27+ or add it to PATH.' }
& $goPath build -trimpath -o sevg.exe .
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
Write-Host "built sevg.exe"
