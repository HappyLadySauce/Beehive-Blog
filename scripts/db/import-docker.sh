#!/usr/bin/env bash
# 与 import-docker.ps1 等价：将仓库内 SQL 导入 Docker 中的 PostgreSQL。
# 用法见文末 --help。

set -euo pipefail

SQL_FILE="db/init.sql"
SQL_DIR="db"
CONTAINER="beehive-postgres"
DB_USER="Beehive-Blog"
DB_NAME="Beehive-Blog"
RESET_SCHEMA=0

fail() {
  echo "错误: $*" >&2
  exit 1
}

usage() {
  cat <<'EOF'
用法: import-docker.sh [选项]

将项目 db/ 目录复制进容器并在其中执行指定 SQL 文件（默认 db/init.sql）。

选项:
  --sql-file PATH     相对仓库根目录的 SQL 文件路径 (默认: db/init.sql)
  --sql-dir DIR       相对仓库根目录、需复制进容器的目录 (默认: db)
  --container NAME    Docker 容器名 (默认: beehive-postgres)
  --db-user USER      PostgreSQL 用户 (默认: Beehive-Blog)
  --db-name NAME      数据库名 (默认: Beehive-Blog)
  --reset-schema      导入前 DROP/CREATE public schema
  -h, --help          显示本说明

示例:
  ./scripts/db/import-docker.sh
  ./scripts/db/import-docker.sh --reset-schema
EOF
}

while [[ $# -gt 0 ]]; do
  case "${1:-}" in
    --sql-file)
      [[ $# -ge 2 ]] || fail "--sql-file 需要参数"
      SQL_FILE="$2"
      shift 2
      ;;
    --sql-dir)
      [[ $# -ge 2 ]] || fail "--sql-dir 需要参数"
      SQL_DIR="$2"
      shift 2
      ;;
    --container)
      [[ $# -ge 2 ]] || fail "--container 需要参数"
      CONTAINER="$2"
      shift 2
      ;;
    --db-user)
      [[ $# -ge 2 ]] || fail "--db-user 需要参数"
      DB_USER="$2"
      shift 2
      ;;
    --db-name)
      [[ $# -ge 2 ]] || fail "--db-name 需要参数"
      DB_NAME="$2"
      shift 2
      ;;
    --reset-schema)
      RESET_SCHEMA=1
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      fail "未知参数: $1（使用 --help 查看用法）"
      ;;
  esac
done

echo "==> 检查 Docker 命令..."
command -v docker >/dev/null 2>&1 || fail "未检测到 docker 命令，请先安装并启动 Docker。"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

resolved_sql_path="$(cd "$ROOT" && realpath "$SQL_FILE")"
resolved_sql_dir="$(cd "$ROOT" && realpath "$SQL_DIR")"

echo "==> 检查 SQL 文件..."
[[ -f "$resolved_sql_path" ]] || fail "SQL 文件不存在: $resolved_sql_path"
[[ -d "$resolved_sql_dir" ]] || fail "SQL 目录不存在: $resolved_sql_dir"

echo "==> 检查容器状态..."
running="$(docker inspect -f '{{.State.Running}}' "$CONTAINER" 2>/dev/null || true)"
if [[ "$running" != "true" ]]; then
  fail "容器未运行或不存在: $CONTAINER"
fi

echo "==> 开始导入数据库..."
echo "    容器: $CONTAINER"
echo "    用户: $DB_USER"
echo "    数据库: $DB_NAME"
echo "    SQL: $resolved_sql_path"
echo "    SQL目录: $resolved_sql_dir"

target_dir="/tmp/beehive-db-import"
echo "==> 复制 SQL 目录到容器临时目录..."
docker exec "$CONTAINER" sh -c "rm -rf ${target_dir} && mkdir -p ${target_dir}" \
  || fail "容器内临时目录初始化失败。"

docker cp "${resolved_sql_dir}/." "${CONTAINER}:${target_dir}" \
  || fail "复制 SQL 文件到容器失败。"

sql_basename="$(basename "$resolved_sql_path")"

if [[ "$RESET_SCHEMA" -eq 1 ]]; then
  echo "==> 重建 public schema..."
  docker exec -i "$CONTAINER" psql -v ON_ERROR_STOP=1 -U "$DB_USER" -d "$DB_NAME" \
    -c "DROP SCHEMA IF EXISTS public CASCADE; CREATE SCHEMA public;" \
    || fail "重建 schema 失败，请检查数据库权限。"
fi

docker exec -i "$CONTAINER" psql -v ON_ERROR_STOP=1 -U "$DB_USER" -d "$DB_NAME" \
  -f "${target_dir}/${sql_basename}" \
  || fail "导入失败，请检查 SQL 内容或数据库连接参数。"

echo "==> 导入完成。"
