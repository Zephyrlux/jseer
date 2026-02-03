#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
cd "$ROOT_DIR"

echo "== 启动说明 =="
echo "1) 默认读取 configs/config.yaml，可用 .env 覆盖"
echo "2) sqlite 必须带 _fk=1（示例: file:jseer.db?_fk=1）"
echo "3) 端口：登录服1863 / 网关5000 / 资源32400 / 登录IP32401 / GM3001 / GM前端5173"
echo ""

check_port() {
  local port=$1
  if command -v lsof >/dev/null 2>&1; then
    if lsof -nP -iTCP:"$port" -sTCP:LISTEN >/dev/null 2>&1; then
      echo "⚠ 端口 $port 已被占用，请先关闭占用进程。"
      lsof -nP -iTCP:"$port" -sTCP:LISTEN || true
      return 1
    fi
  fi
  return 0
}

check_port 1863 || exit 1
check_port 5000 || exit 1
check_port 32400 || exit 1
check_port 32401 || exit 1
check_port 3001 || exit 1
check_port "${GM_WEB_PORT:-5173}" || exit 1

if [[ -f .env ]]; then
  # 自动导出 .env 里的变量，供子进程读取
  set -a
  # shellcheck disable=SC1091
  source .env
  set +a
fi

echo "== 当前环境变量预览 =="
echo "DATABASE_DRIVER=${DATABASE_DRIVER:-}"
echo "DATABASE_DSN=${DATABASE_DSN:-}"
echo ""

export JSEER_CONFIG=${JSEER_CONFIG:-"configs/config.yaml"}

pids=()

start_service() {
  local name=$1
  local cmd=$2
  echo "[start] $name"
  bash -c "$cmd" &
  pids+=("$!")
}

start_service "loginserver" "go run ./cmd/loginserver"
start_service "gateway" "go run ./cmd/gateway"
start_service "ressrv" "go run ./cmd/ressrv"
start_service "gmserver" "go run ./cmd/gmserver"
start_service "gm-web" "cd gm-web && npm run dev -- --host 0.0.0.0 --port ${GM_WEB_PORT:-5173}"

echo "\nAll services started. PIDs: ${pids[*]}"
echo "资源地址: http://localhost:32400/index.html"
echo "登录IP:  http://localhost:32401/ip.txt"
echo "GM 地址:  http://localhost:3001"
echo "GM 前端: http://localhost:${GM_WEB_PORT:-5173}"

cleanup() {
  echo "\nStopping services..."
  for pid in "${pids[@]}"; do
    kill "$pid" >/dev/null 2>&1 || true
  done
}

trap cleanup INT TERM EXIT
wait
