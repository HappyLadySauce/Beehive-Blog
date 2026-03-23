# 设置环境变量脚本
# 从 .env 文件读取配置并设置到当前 PowerShell 会话

$envFile = Join-Path $PSScriptRoot ".env"

if (-not (Test-Path $envFile)) {
    Write-Error "环境变量文件不存在: $envFile"
    exit 1
}

Write-Host "正在加载环境变量..." -ForegroundColor Cyan

Get-Content $envFile | ForEach-Object {
    $line = $_.Trim()
    
    # 跳过空行和注释
    if ([string]::IsNullOrWhiteSpace($line) -or $line.StartsWith("#")) {
        return
    }
    
    # 解析 KEY=VALUE 格式
    if ($line -match "^([^=]+)=(.*)$") {
        $key = $matches[1].Trim()
        $value = $matches[2].Trim()
        
        # 移除可能的引号
        if ($value.StartsWith('"') -and $value.EndsWith('"')) {
            $value = $value.Substring(1, $value.Length - 2)
        }
        if ($value.StartsWith("'") -and $value.EndsWith("'")) {
            $value = $value.Substring(1, $value.Length - 2)
        }
        
        # 设置环境变量
        [Environment]::SetEnvironmentVariable($key, $value, "Process")
        Write-Host "  ✓ $key = $value" -ForegroundColor Green
    }
}

Write-Host "`n环境变量加载完成!" -ForegroundColor Cyan
Write-Host "当前会话已设置以下环境变量:" -ForegroundColor Gray
Get-ChildItem Env: | Where-Object { 
    $_.Name -in @("GEO_LITE2_LICENSE_KEY", "IPIPToken", "BaiduAK") 
} | ForEach-Object {
    Write-Host "  $($_.Name) = $($_.Value)" -ForegroundColor DarkGray
}
