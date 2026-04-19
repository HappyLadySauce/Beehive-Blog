param(
    [Parameter(Mandatory = $true)]
    [string]$ProtoFile,

    [Parameter(Mandatory = $true)]
    [string]$OutputDir
)

if (-not (Get-Command protoc -ErrorAction SilentlyContinue)) {
    Write-Error "protoc is required before running RPC generation."
    exit 1
}

goctl rpc protoc $ProtoFile --go_out=. --go-grpc_out=. --zrpc_out=$OutputDir

$repoRoot = Resolve-Path (Join-Path $PSScriptRoot "../..")
$descriptorDir = Join-Path $repoRoot "services/gateway/etc/protos"
if (-not (Test-Path -LiteralPath $descriptorDir)) {
    New-Item -ItemType Directory -Path $descriptorDir | Out-Null
}

$service = [System.IO.Path]::GetFileNameWithoutExtension($ProtoFile)
$descriptorPath = Join-Path $descriptorDir "$service.protoset"
protoc --descriptor_set_out=$descriptorPath --include_imports $ProtoFile
