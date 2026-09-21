#!/usr/bin/env bash
# =============================================================================
# LIMS 第三方检测实验室管理系统 — 一键启动脚本
# =============================================================================
# 用法:
#   ./start.sh dev       开发模式 (go run + npm run dev)
#   ./start.sh deploy    部署模式 (编译后运行)
#   ./start.sh stop      停止服务
#   ./start.sh --help    帮助
# =============================================================================
# 前置条件:
#   1. PostgreSQL 已安装并运行，建好数据库 (默认: lims)
#   2. 后端配置: backend/config/config.yaml (或 config.dev.yaml)
#   3. 前端配置: frontend/.env (VITE_API_BASE_URL)
#   4. Go 1.21+, Node.js 18+
# =============================================================================

set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PID_FILE="/tmp/lims-pids"

log_info()  { echo -e "${BLUE}[INFO]${NC}  $*"; }
log_ok()    { echo -e "${GREEN}[OK]${NC}    $*"; }
log_warn()  { echo -e "${YELLOW}[WARN]${NC}  $*"; }
log_error() { echo -e "${RED}[ERROR]${NC} $*"; }
log_step()  { echo -e "\n${CYAN}══════════════════════════════════════════════════${NC}"; echo -e "${CYAN}  $*${NC}"; echo -e "${CYAN}══════════════════════════════════════════════════${NC}\n"; }

show_help() {
    cat << 'EOF'
LIMS 第三方检测实验室管理系统 — 一键启动脚本

用法:
  ./start.sh dev        开发模式 — 本地热重载
                        后端: go run ./cmd/server
                        前端: npm run dev (端口 3000)

  ./start.sh deploy     部署模式 — 编译后运行
                        先编译前后端，再启动服务
                        前端: npm run build → nginx 或静态文件

  ./start.sh stop       停止所有 LIMS 服务进程

  ./start.sh --help     显示此帮助

前置条件:
  1. PostgreSQL 已安装并运行（数据库需提前创建）
  2. 后端数据库配置: backend/config/config.yaml
  3. Go 1.21+, Node.js 18+
EOF
    exit 0
}

# 检查前置条件
check_prerequisites() {
    local missing=0

    if ! command -v go &>/dev/null; then
        log_error "Go 未安装 → https://go.dev/dl/"
        missing=1
    fi

    if ! command -v node &>/dev/null; then
        log_error "Node.js 未安装 → https://nodejs.org/"
        missing=1
    fi

    if ! command -v npm &>/dev/null; then
        log_error "npm 未安装"
        missing=1
    fi

    if [ "$missing" -eq 1 ]; then
        exit 1
    fi

    log_ok "Go $(go version | awk '{print $3}' | sed 's/go//') ✓  Node.js $(node -v) ✓"
}

# 保存 PID
save_pid() {
    echo "$1" >> "$PID_FILE"
}

# 读取 PID
read_pids() {
    if [ -f "$PID_FILE" ]; then
        cat "$PID_FILE"
    fi
}

# =============================================================================
# 停止服务
# =============================================================================
stop_services() {
    log_info "正在停止 LIMS 服务..."

    if [ ! -f "$PID_FILE" ]; then
        log_warn "未找到运行中的 LIMS 服务 (PID 文件不存在)"
        return
    fi

    while IFS= read -r pid; do
        if [ -n "$pid" ] && kill -0 "$pid" 2>/dev/null; then
            log_info "停止进程 PID: $pid"
            kill "$pid" 2>/dev/null || true
        fi
    done < "$PID_FILE"

    rm -f "$PID_FILE"
    log_ok "所有 LIMS 服务已停止"
}

# =============================================================================
# 开发模式 — Mac / Linux 本地热重载
# =============================================================================
run_dev() {
    echo -e "${CYAN}"
    echo "  ╔══════════════════════════════════════════╗"
    echo "  ║    LIMS 开发模式                          ║"
    echo "  ║    Go 后端 + Vue 前端 (热重载)            ║"
    echo "  ╚══════════════════════════════════════════╝"
    echo -e "${NC}"

    check_prerequisites

    # 先停掉旧进程
    stop_services 2>/dev/null || true
    : > "$PID_FILE"

    trap 'stop_services; exit' INT TERM EXIT

    # ── 启动后端 ──────────────────────────────────
    log_step "[1/2] 启动 Go 后端 (端口 8080)"
    cd "$PROJECT_ROOT/backend"

    go run ./cmd/server &
    BACKEND_PID=$!
    save_pid "$BACKEND_PID"
    log_info "后端进程 PID: $BACKEND_PID"

    sleep 3
    if kill -0 "$BACKEND_PID" 2>/dev/null; then
        log_ok "后端已启动 → http://localhost:8080/api/health"
    else
        log_error "后端启动失败，请检查数据库连接等配置"
        log_error "配置位置: backend/config/config.yaml"
        exit 1
    fi

    # ── 启动前端 ──────────────────────────────────
    log_step "[2/2] 启动 Vue 前端 (端口 3000)"
    cd "$PROJECT_ROOT/frontend"

    if [ ! -d node_modules ]; then
        log_info "正在安装前端依赖..."
        npm install
        log_ok "前端依赖安装完成"
    fi

    npm run dev &
    FRONTEND_PID=$!
    save_pid "$FRONTEND_PID"
    log_info "前端进程 PID: $FRONTEND_PID"

    sleep 3

    # ── 完成 ──────────────────────────────────────
    echo ""
    echo -e "${GREEN}══════════════════════════════════════════════════${NC}"
    echo -e "${GREEN}  LIMS 启动成功!${NC}"
    echo -e "${GREEN}  ┌─────────────────────────────────────────┐${NC}"
    echo -e "${GREEN}  │  前端:  http://localhost:3000            │${NC}"
    echo -e "${GREEN}  │  后端:  http://localhost:8080/api/health │${NC}"
    echo -e "${GREEN}  └─────────────────────────────────────────┘${NC}"
    echo -e "${GREEN}══════════════════════════════════════════════════${NC}"
    echo ""
    log_info "按 Ctrl+C 停止所有服务"

    # 保持前台运行
    wait
}

# =============================================================================
# 部署模式 — 编译后运行
# =============================================================================
run_deploy() {
    echo -e "${CYAN}"
    echo "  ╔══════════════════════════════════════════╗"
    echo "  ║    LIMS 部署模式                          ║"
    echo "  ║    编译后端 + 构建前端 → 运行             ║"
    echo "  ╚══════════════════════════════════════════╝"
    echo -e "${NC}"

    check_prerequisites

    # 先停掉旧进程
    stop_services 2>/dev/null || true
    : > "$PID_FILE"

    trap 'stop_services; exit' INT TERM EXIT

    # ── 编译后端 ──────────────────────────────────
    log_step "[1/4] 编译后端"
    cd "$PROJECT_ROOT/backend"
    go build -o lims-server ./cmd/server
    log_ok "后端编译完成: backend/lims-server"

    # ── 构建前端 ──────────────────────────────────
    log_step "[2/4] 构建前端"
    cd "$PROJECT_ROOT/frontend"
    if [ ! -d node_modules ]; then
        log_info "正在安装前端依赖..."
        npm install
        log_ok "前端依赖安装完成"
    fi
    npm run build
    log_ok "前端构建完成: frontend/dist/"

    # ── 启动后端 ──────────────────────────────────
    log_step "[3/4] 启动后端 (端口 8080)"
    cd "$PROJECT_ROOT/backend"
    ./lims-server &
    BACKEND_PID=$!
    save_pid "$BACKEND_PID"
    log_info "后端进程 PID: $BACKEND_PID"

    sleep 3
    if kill -0 "$BACKEND_PID" 2>/dev/null; then
        log_ok "后端已启动 → http://localhost:8080/api/health"
    else
        log_error "后端启动失败，请检查配置"
        exit 1
    fi

    # ── 启动前端 (内置 Node 服务器，仅开发/测试用) ─
    log_step "[4/4] 启动前端"
    cd "$PROJECT_ROOT/frontend"
    npm run preview &
    FRONTEND_PID=$!
    save_pid "$FRONTEND_PID"
    log_info "前端进程 PID: $FRONTEND_PID"

    sleep 2

    # ── 完成 ──────────────────────────────────────
    echo ""
    echo -e "${GREEN}══════════════════════════════════════════════════${NC}"
    echo -e "${GREEN}  LIMS 部署模式启动成功!${NC}"
    echo -e "${GREEN}  ┌─────────────────────────────────────────┐${NC}"
    echo -e "${GREEN}  │  前端:  http://localhost:4173            │${NC}"
    echo -e "${GREEN}  │  后端:  http://localhost:8080/api/health │${NC}"
    echo -e "${GREEN}  └─────────────────────────────────────────┘${NC}"
    echo -e "${GREEN}══════════════════════════════════════════════════${NC}"
    echo ""
    log_info "按 Ctrl+C 停止所有服务"
    log_info "生产环境建议用 nginx 代理前端静态文件 (frontend/dist/)"
    log_info "  nginx 配置参考: deploy/nginx/lims.conf"

    wait
}

# =============================================================================
# 主入口
# =============================================================================
main() {
    case "${1:-auto}" in
        --help|-h)
            show_help
            ;;
        dev)
            run_dev
            ;;
        deploy)
            run_deploy
            ;;
        stop)
            stop_services
            ;;
        auto)
            log_info "请指定运行模式: ./start.sh dev  或  ./start.sh deploy"
            log_info "  dev     开发模式 (go run + npm run dev)"
            log_info "  deploy  部署模式 (编译后运行)"
            exit 1
            ;;
        *)
            log_error "未知模式: $1"
            log_error "用法: ./start.sh {dev|deploy|stop|--help}"
            exit 1
            ;;
    esac
}

main "$@"