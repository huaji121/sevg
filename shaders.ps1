$ErrorActionPreference = 'Stop'
$glslang = Get-Command glslang -ErrorAction SilentlyContinue
if ($glslang) {
    $exe = $glslang.Source
} elseif (Test-Path '.tools\glslang\bin\glslang.exe') {
    $exe = (Resolve-Path '.tools\glslang\bin\glslang.exe').Path
} else {
    throw 'glslang.exe not found. Download glslang and put it in .tools/glslang/bin or add it to PATH.'
}
$dir = Join-Path $PSScriptRoot 'internal\render\shaders'
Get-ChildItem $dir -File | Where-Object { $_.Extension -in '.vert', '.frag' } | ForEach-Object {
    & $exe -V --target-env vulkan1.0 -o ($_.FullName + '.spv') $_.FullName
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
}
Write-Host "compiled shaders in $dir"