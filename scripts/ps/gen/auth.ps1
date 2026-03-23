# 基于脚本所在目录解析路径，保证从任意工作目录执行都能找到 proto 文件
$ScriptDir = $PSScriptRoot
$RootDir = (Resolve-Path (Join-Path $ScriptDir "..\..\..")).Path
$ProtoDir = Join-Path $RootDir "proto"
$RpcFile = Join-Path $ProtoDir "beehive-m-auth.proto"
$OutDir = Join-Path $RootDir "services\auth"
# pb 生成到 services\auth\pb：--go_out 为项目根，go_package 为 "services/auth/pb"
$OutPbBase = $RootDir

if (-not (Test-Path $RpcFile)) {
    Write-Error "rpc 文件不存在: $RpcFile"
    exit 1
}

# 确保输出目录存在，否则 Push-Location 和 goctl 会失败
New-Item -ItemType Directory -Force -Path $OutDir | Out-Null

# 在项目根目录执行，并用 -I proto 指定 proto_path，传相对路径 proto/beehive-m-auth.proto
Push-Location $RootDir
try {
    goctl rpc protoc proto/beehive-m-auth.proto -I proto --go_out=$OutPbBase --go-grpc_out=$OutPbBase --zrpc_out=$OutDir --style=goZero --client=true -m
} finally {
    Pop-Location
}