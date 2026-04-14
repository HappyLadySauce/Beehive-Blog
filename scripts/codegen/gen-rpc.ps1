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
