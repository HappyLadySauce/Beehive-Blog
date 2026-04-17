param(
    [switch]$WithGateway,
    [switch]$SkipRpc,
    [switch]$SkipCheck
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

function Invoke-Step {
    param(
        [string]$Name,
        [scriptblock]$Action
    )

    Write-Host ""
    Write-Host "==> $Name" -ForegroundColor Cyan
    & $Action
}

if (-not (Get-Command goctl -ErrorAction SilentlyContinue)) {
    Write-Error "goctl is required."
    exit 1
}

$repoRoot = Resolve-Path (Join-Path $PSScriptRoot "../..")
$gatewayScript = Join-Path $PSScriptRoot "gen-gateway.ps1"
$rpcScript = Join-Path $PSScriptRoot "gen-rpc.ps1"
$checkScript = Join-Path $PSScriptRoot "check-contract-sync.ps1"
$protoDir = Join-Path $repoRoot "proto"

if ($WithGateway) {
    Invoke-Step -Name "Generate gateway from api/gateway.api" -Action {
        & $gatewayScript
        if ($LASTEXITCODE -ne 0) {
            throw "gateway generation failed"
        }
    }
}
else {
    Write-Host ""
    Write-Host "skip gateway generation by default (pass -WithGateway to enable)." -ForegroundColor Yellow
}

if (-not $SkipRpc) {
    if (-not (Get-Command protoc -ErrorAction SilentlyContinue)) {
        Write-Error "protoc is required for RPC generation."
        exit 1
    }

    $protoFiles = Get-ChildItem -LiteralPath $protoDir -Filter *.proto -File | Sort-Object Name
    if ($protoFiles.Count -eq 0) {
        Write-Host ""
        Write-Host "no proto files found, skip RPC generation." -ForegroundColor Yellow
    }
    else {
        foreach ($proto in $protoFiles) {
            $service = [System.IO.Path]::GetFileNameWithoutExtension($proto.Name)
            $outputDir = "services/$service"
            $protoPath = "proto/$($proto.Name)"
            Invoke-Step -Name "Generate RPC for $($proto.Name) -> $outputDir" -Action {
                & $rpcScript -ProtoFile $protoPath -OutputDir $outputDir
                if ($LASTEXITCODE -ne 0) {
                    throw "rpc generation failed for $($proto.Name)"
                }
            }
        }
    }
}

if (-not $SkipCheck) {
    Invoke-Step -Name "Run contract sync check" -Action {
        & $checkScript
        if ($LASTEXITCODE -ne 0) {
            throw "contract sync check failed"
        }
    }
}

Write-Host ""
Write-Host "sync-all completed." -ForegroundColor Green
