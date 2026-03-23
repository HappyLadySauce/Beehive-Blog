# 基于脚本所在目录解析路径，保证从任意工作目录执行都能找到 api 文件
$ScriptDir = $PSScriptRoot
$ApiFile = Join-Path $ScriptDir "..\..\..\api\beehive-m-gateway.api"
$OutDir = Join-Path $ScriptDir "..\..\.."

if (-not (Test-Path $ApiFile)) {
    Write-Error "api 文件不存在: $ApiFile"
    exit 1
}

Push-Location $OutDir
try {
    goctl api go -api $ApiFile --style=goZero -dir .\services\gateway 
} finally {
    Pop-Location
}