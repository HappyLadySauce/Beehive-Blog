param(
    [string]$SqlFile = "db/init.sql",
    [string]$SqlDir = "db",
    [string]$Container = "beehive-postgres",
    [string]$DbUser = "Beehive-Blog",
    [string]$DbName = "Beehive-Blog",
    [switch]$ResetSchema
)

$ErrorActionPreference = "Stop"

function Fail([string]$Message) {
    Write-Error $Message
    exit 1
}

Write-Host "==> 检查 Docker 命令..."
if (-not (Get-Command docker -ErrorAction SilentlyContinue)) {
    Fail "未检测到 docker 命令，请先安装并启动 Docker Desktop。"
}

Write-Host "==> 检查 SQL 文件..."
$resolvedSqlPath = Join-Path -Path $PSScriptRoot -ChildPath "..\..\$SqlFile"
$resolvedSqlPath = [System.IO.Path]::GetFullPath($resolvedSqlPath)
if (-not (Test-Path -LiteralPath $resolvedSqlPath)) {
    Fail "SQL 文件不存在: $resolvedSqlPath"
}
$resolvedSqlDir = Join-Path -Path $PSScriptRoot -ChildPath "..\..\$SqlDir"
$resolvedSqlDir = [System.IO.Path]::GetFullPath($resolvedSqlDir)
if (-not (Test-Path -LiteralPath $resolvedSqlDir)) {
    Fail "SQL 目录不存在: $resolvedSqlDir"
}

Write-Host "==> 检查容器状态..."
$running = docker inspect -f "{{.State.Running}}" $Container 2>$null
if ($LASTEXITCODE -ne 0 -or "$running".Trim() -ne "true") {
    Fail "容器未运行或不存在: $Container"
}

Write-Host "==> 开始导入数据库..."
Write-Host "    容器: $Container"
Write-Host "    用户: $DbUser"
Write-Host "    数据库: $DbName"
Write-Host "    SQL: $resolvedSqlPath"
Write-Host "    SQL目录: $resolvedSqlDir"

Write-Host "==> 复制 SQL 目录到容器临时目录..."
$targetDir = "/tmp/beehive-db-import"
docker exec $Container sh -c "rm -rf $targetDir && mkdir -p $targetDir"
if ($LASTEXITCODE -ne 0) {
    Fail "容器内临时目录初始化失败。"
}

docker cp "$resolvedSqlDir/." "${Container}:${targetDir}"
if ($LASTEXITCODE -ne 0) {
    Fail "复制 SQL 文件到容器失败。"
}

$sqlBaseName = [System.IO.Path]::GetFileName($resolvedSqlPath)
if ($ResetSchema) {
    Write-Host "==> 重建 public schema..."
    docker exec -i $Container psql -v ON_ERROR_STOP=1 -U $DbUser -d $DbName -c "DROP SCHEMA IF EXISTS public CASCADE; CREATE SCHEMA public;"
    if ($LASTEXITCODE -ne 0) {
        Fail "重建 schema 失败，请检查数据库权限。"
    }
}

docker exec -i $Container psql -v ON_ERROR_STOP=1 -U $DbUser -d $DbName -f "${targetDir}/${sqlBaseName}"
if ($LASTEXITCODE -ne 0) {
    Fail "导入失败，请检查 SQL 内容或数据库连接参数。"
}

Write-Host "==> 导入完成。"
