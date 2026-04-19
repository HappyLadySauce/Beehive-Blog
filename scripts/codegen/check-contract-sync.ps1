param(
    [string]$GatewayApi = "api/gateway.api",
    [string]$GatewayRoutes = "services/gateway/internal/handler/routes.go",
    [string]$ProtoDir = "proto",
    [string]$GatewayConfig = "services/gateway/etc/gateway.yaml"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

function Get-ApiRoutes {
    param([string]$Path)

    if (-not (Test-Path -LiteralPath $Path)) {
        throw "gateway api file not found: $Path"
    }

    $content = Get-Content -LiteralPath $Path -Raw
    $regex = [regex]'(?im)^\s*(get|post|put|delete|patch|head|options)\s+(\S+)'
    $routes = @{}
    foreach ($m in $regex.Matches($content)) {
        $method = $m.Groups[1].Value.ToUpperInvariant()
        $routePath = $m.Groups[2].Value.Trim()
        $key = "$method $routePath"
        $routes[$key] = $true
    }
    return $routes
}

function Get-GoRoutes {
    param([string]$Path)

    if (-not (Test-Path -LiteralPath $Path)) {
        throw "gateway routes file not found: $Path"
    }

    $content = Get-Content -LiteralPath $Path -Raw
    $regex = [regex]'Method:\s+http\.Method([A-Za-z]+),\s*Path:\s+"([^"]+)"'
    $routes = @{}
    foreach ($m in $regex.Matches($content)) {
        $method = $m.Groups[1].Value.ToUpperInvariant()
        $routePath = $m.Groups[2].Value.Trim()
        $key = "$method $routePath"
        $routes[$key] = $true
    }
    return $routes
}

function Get-UpstreamRoutesAndIssues {
    param([string]$ConfigPath)

    $result = @{
        Routes = @{}
        Issues = New-Object System.Collections.Generic.List[string]
    }

    if (-not (Test-Path -LiteralPath $ConfigPath)) {
        $result.Issues.Add("gateway config not found: $ConfigPath")
        return $result
    }

    $configDir = Split-Path -Parent $ConfigPath
    $rawMain = Get-Content -LiteralPath $ConfigPath -Raw
    $upstreamFiles = @()
    foreach ($line in ($rawMain -split "`r?`n")) {
        if ($line -match '^\s*-\s+(.+\.ya?ml)\s*$') {
            $upstreamFiles += $matches[1].Trim()
        }
    }

    foreach ($upstreamFile in $upstreamFiles) {
        $path = $upstreamFile
        if (-not [System.IO.Path]::IsPathRooted($path)) {
            $path = Join-Path $configDir $upstreamFile
        }
        if (-not (Test-Path -LiteralPath $path)) {
            $result.Issues.Add("upstream config not found: $path")
            continue
        }

        $lines = Get-Content -LiteralPath $path
        for ($i = 0; $i -lt $lines.Count; $i++) {
            if ($lines[$i] -match '^\s*-\s+Method:\s*(\w+)\s*$') {
                $method = $matches[1].Trim().ToUpperInvariant()
                if ($i + 1 -lt $lines.Count -and $lines[$i + 1] -match '^\s*Path:\s*(\S+)\s*$') {
                    $routePath = $matches[1].Trim()
                    $result.Routes["$method $routePath"] = $true
                }
            }

            if ($lines[$i] -match '^\s*-\s+(.+\.(pb|protoset))\s*$') {
                $protoSetPath = $matches[1].Trim()
                if (-not [System.IO.Path]::IsPathRooted($protoSetPath)) {
                    $protoSetPath = Join-Path (Split-Path -Parent $path) $protoSetPath
                }
                if (-not (Test-Path -LiteralPath $protoSetPath)) {
                    $result.Issues.Add("missing protoset: $protoSetPath")
                    continue
                }

                $service = [System.IO.Path]::GetFileNameWithoutExtension($protoSetPath)
                $protoPath = Join-Path $ProtoDir "$service.proto"
                if (-not (Test-Path -LiteralPath $protoPath)) {
                    $result.Issues.Add("missing proto for protoset: $protoPath")
                    continue
                }

                $protoMtime = (Get-Item -LiteralPath $protoPath).LastWriteTimeUtc
                $protoSetMtime = (Get-Item -LiteralPath $protoSetPath).LastWriteTimeUtc
                if ($protoSetMtime -lt $protoMtime) {
                    $result.Issues.Add("stale protoset: $protoSetPath (older than $protoPath)")
                }
            }
        }
    }

    return $result
}

function Get-RpcProtoIssues {
    param([string]$Dir)

    $issues = New-Object System.Collections.Generic.List[string]

    if (-not (Test-Path -LiteralPath $Dir)) {
        $issues.Add("proto directory not found: $Dir")
        return $issues
    }

    $protoFiles = Get-ChildItem -LiteralPath $Dir -Filter *.proto -File
    foreach ($proto in $protoFiles) {
        $service = [System.IO.Path]::GetFileNameWithoutExtension($proto.Name)
        $serviceRoot = Join-Path "services" $service
        $pbDir = Join-Path $serviceRoot "pb"

        if (-not (Test-Path -LiteralPath $pbDir)) {
            $issues.Add("missing generated pb directory: $pbDir (from $($proto.FullName))")
            continue
        }

        $pbFile = Join-Path $pbDir "$service.pb.go"
        $grpcFile = Join-Path $pbDir "$service`_grpc.pb.go"

        foreach ($f in @($pbFile, $grpcFile)) {
            if (-not (Test-Path -LiteralPath $f)) {
                $issues.Add("missing generated file: $f (from $($proto.FullName))")
                continue
            }

            $protoMtime = (Get-Item -LiteralPath $proto.FullName).LastWriteTimeUtc
            $generatedMtime = (Get-Item -LiteralPath $f).LastWriteTimeUtc
            if ($generatedMtime -lt $protoMtime) {
                $issues.Add("stale generated file: $f (older than $($proto.FullName))")
            }
        }

        $zrpcClientFile = Join-Path (Join-Path $serviceRoot $service) "$service.go"
        if (-not (Test-Path -LiteralPath $zrpcClientFile)) {
            $issues.Add("missing zrpc client file: $zrpcClientFile (from $($proto.FullName))")
        }
    }

    return $issues
}

try {
    $apiRoutes = Get-ApiRoutes -Path $GatewayApi
    $goRoutes = Get-GoRoutes -Path $GatewayRoutes
    $upstream = Get-UpstreamRoutesAndIssues -ConfigPath $GatewayConfig

    $onlyInApi = @($apiRoutes.Keys | Where-Object { -not $goRoutes.ContainsKey($_) } | Sort-Object)
    $onlyInGo = @($goRoutes.Keys | Where-Object { -not $apiRoutes.ContainsKey($_) } | Sort-Object)
    $rpcIssues = @(Get-RpcProtoIssues -Dir $ProtoDir)
    $routeOverlap = @($apiRoutes.Keys | Where-Object { $upstream.Routes.ContainsKey($_) } | Sort-Object)
    $upstreamIssues = @($upstream.Issues)

    if ($onlyInApi.Count -eq 0 -and $onlyInGo.Count -eq 0 -and $rpcIssues.Count -eq 0 -and $routeOverlap.Count -eq 0 -and $upstreamIssues.Count -eq 0) {
        Write-Host "contract sync check passed." -ForegroundColor Green
        exit 0
    }

    Write-Host "contract sync check failed." -ForegroundColor Red

    if ($onlyInApi.Count -gt 0) {
        Write-Host ""
        Write-Host "routes declared in api but missing in routes.go:" -ForegroundColor Yellow
        foreach ($item in $onlyInApi) {
            Write-Host "  - $item"
        }
    }

    if ($onlyInGo.Count -gt 0) {
        Write-Host ""
        Write-Host "routes declared in routes.go but missing in api:" -ForegroundColor Yellow
        foreach ($item in $onlyInGo) {
            Write-Host "  - $item"
        }
    }

    if ($rpcIssues.Count -gt 0) {
        Write-Host ""
        Write-Host "rpc/proto sync issues:" -ForegroundColor Yellow
        foreach ($issue in $rpcIssues) {
            Write-Host "  - $issue"
        }
    }

    if ($routeOverlap.Count -gt 0) {
        Write-Host ""
        Write-Host "routes declared in both api and upstream mappings:" -ForegroundColor Yellow
        foreach ($item in $routeOverlap) {
            Write-Host "  - $item"
        }
    }

    if ($upstreamIssues.Count -gt 0) {
        Write-Host ""
        Write-Host "gateway upstream sync issues:" -ForegroundColor Yellow
        foreach ($issue in $upstreamIssues) {
            Write-Host "  - $issue"
        }
    }

    exit 1
}
catch {
    Write-Error $_
    exit 1
}
