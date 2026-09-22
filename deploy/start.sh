#!/usr/bin/env bash
# =============================================================================
# LIMS 第三方检测实验室管理系统 — 一键启动脚本
# =============================================================================
# 用法:
#   ./start.sh dev                 开发模式 (go run + npm run dev)
#   ./start.sh ubuntu              构建 Ubuntu 部署包 → pkg/lims-deploy.tar.gz
#   ./start.sh stop                停止本地 LIMS 服务
#   ./start.sh --help              帮助
# =============================================================================
# 前置条件:
#   开发模式: Go 1.21+, Node.js 18+, PostgreSQL 已就绪并配置好
#   Ubuntu 构建: Go 1.21+, Node.js 18+ (会在 Mac 上交叉编译)
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
PKG_DIR="$PROJECT_ROOT/pkg"

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
                        后端: go run ./cmd/server (端口 8080)
                        前端: npm run dev (端口 3000)

  ./start.sh ubuntu     构建 Ubuntu 部署包
                        交叉编译 Linux 后端 + 构建前端
                        打包至 pkg/lims-deploy.tar.gz

  ./start.sh stop       停止所有 LIMS 本地服务进程

  ./start.sh --help     显示此帮助

前置条件:
  1. PostgreSQL 已安装并运行（数据库需提前创建）
  2. 后端数据库配置: backend/config/config.yaml
  3. Go 1.21+, Node.js 18+
EOF
    exit 0
}

# ── 检查前置条件 ──────────────────────────────────
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

# ── PID 管理 ──────────────────────────────────────
save_pid() { echo "$1" >> "$PID_FILE"; }

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

    wait
}

# =============================================================================
# Ubuntu 构建 — 在 Mac 上交叉编译并打包
# =============================================================================
build_ubuntu() {
    echo -e "${CYAN}"
    echo "  ╔══════════════════════════════════════════╗"
    echo "  ║    LIMS Ubuntu 构建                       ║"
    echo "  ║    交叉编译后端 + 构建前端 → 打包          ║"
    echo "  ╚══════════════════════════════════════════╝"
    echo -e "${NC}"

    check_prerequisites

    # 清理 & 创建 pkg 目录
    log_step "[0/4] 清理构建目录"
    rm -rf "$PKG_DIR"
    mkdir -p "$PKG_DIR/frontend/dist"
    mkdir -p "$PKG_DIR/backend/config"
    mkdir -p "$PKG_DIR/deploy/nginx"
    log_ok "构建目录已准备: $PKG_DIR"

    # ── 编译后端 (Linux amd64) ────────────────────
    log_step "[1/4] 交叉编译后端 (Linux amd64)"
    cd "$PROJECT_ROOT/backend"

    GOOS=linux GOARCH=amd64 go build -o "$PKG_DIR/lims-server-linux" ./cmd/server
    log_ok "后端编译完成: $(ls -lh $PKG_DIR/lims-server-linux | awk '{print $5}')"

    # ── 复制配置文件 ──────────────────────────────
    log_step "[2/4] 复制配置文件"
    cp "$PROJECT_ROOT/backend/config/config.yaml" "$PKG_DIR/backend/config/"
    cp "$PROJECT_ROOT/backend/config/config.dev.yaml" "$PKG_DIR/backend/config/"
    if [ -f "$PROJECT_ROOT/backend/config/config.prod.yaml" ]; then
        cp "$PROJECT_ROOT/backend/config/config.prod.yaml" "$PKG_DIR/backend/config/"
    fi
    log_ok "配置文件已复制"

    # ── 构建前端 ──────────────────────────────────
    log_step "[3/4] 构建前端"
    cd "$PROJECT_ROOT/frontend"

    if [ ! -d node_modules ]; then
        log_info "正在安装前端依赖..."
        npm install
        log_ok "前端依赖安装完成"
    fi
    npm run build
    cp -r dist/* "$PKG_DIR/frontend/dist/"
    log_ok "前端构建完成: $(ls -lh $PKG_DIR/frontend/dist/index.html | awk '{print $5}')"

    # ── 复制 Nginx 配置并打包 ─────────────────────
    log_step "[4/4] 打包部署文件"
    cp "$PROJECT_ROOT/deploy/nginx/lims.conf" "$PKG_DIR/deploy/nginx/"

    # 生成 Ubuntu 部署说明文件
    cat > "$PKG_DIR/README.txt" << 'DEPLOYEOF'
=========================================
LIMS Ubuntu 部署指南
=========================================

【前置条件】
  1. PostgreSQL 已安装并运行，创建好数据库
     sudo apt install postgresql
     sudo -u postgres psql -c "CREATE USER lims WITH PASSWORD 'yourpass';"
     sudo -u postgres psql -c "CREATE DATABASE lims OWNER lims;"

  2. 配置后端数据库连接
     vi backend/config/config.dev.yaml
     修改 database 下的 user/password/dbname

【部署步骤】
  # 1. 解压
  tar xzf lims-deploy.tar.gz
  cd lims-deploy

  # 2. 启动后端 (后台运行)
  nohup ./lims-server-linux > lims.log 2>&1 &
  curl http://localhost:8080/api/health  # 验证

  # 3. 配置 nginx 前端
  sudo apt install nginx
  sudo mkdir -p /var/www/lims
  sudo cp -r frontend/dist/* /var/www/lims/

  # 4. 配置 nginx 反向代理
  sudo cp deploy/nginx/lims.conf /etc/nginx/sites-available/lims
  sudo ln -sf /etc/nginx/sites-available/lims /etc/nginx/sites-enabled/
  sudo rm -f /etc/nginx/sites-enabled/default
  sudo nginx -t && sudo systemctl restart nginx

  # 5. 访问 http://服务器IP
DEPLOYEOF

    cd "$PKG_DIR"
    tar czf "$PKG_DIR/lims-deploy.tar.gz" \
        lims-server-linux \
        backend/ \
        frontend/ \
        deploy/ \
        README.txt

    log_ok "打包完成: $PKG_DIR/lims-deploy.tar.gz"
    log_info "压缩包大小: $(ls -lh $PKG_DIR/lims-deploy.tar.gz | awk '{print $5}')"

    # ── 完成 ──────────────────────────────────────
    echo ""
    echo -e "${GREEN}══════════════════════════════════════════════════${NC}"
    echo -e "${GREEN}  Ubuntu 部署包构建完成!${NC}"
    echo -e "${GREEN}══════════════════════════════════════════════════${NC}"
    echo ""
    echo "  pkg/ 目录内容:"
    echo "    ├── lims-server-linux         $(ls -lh $PKG_DIR/lims-server-linux | awk '{print $5}')"
    echo "    ├── frontend/dist/            前端静态文件"
    echo "    ├── backend/config/           后端配置文件"
    echo "    ├── deploy/nginx/lims.conf    Nginx 配置"
    echo "    └── lims-deploy.tar.gz        $(ls -lh $PKG_DIR/lims-deploy.tar.gz | awk '{print $5}')"
    echo ""
    echo "  传到 Ubuntu 服务器:"
    echo "    scp $PKG_DIR/lims-deploy.tar.gz root@你的服务器IP:/opt/"
    echo ""
    echo "  部署指南见: pkg/README.txt"
    echo ""
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
        ubuntu)
            build_ubuntu
            ;;
        stop)
            stop_services
            ;;
        auto)
            log_info "请指定运行模式:"
            log_info "  ./start.sh dev        开发模式 (go run + npm run dev)"
            log_info "  ./start.sh ubuntu     构建 Ubuntu 部署包"
            log_info "  ./start.sh stop       停止服务"
            exit 1
            ;;
        *)
            log_error "未知模式: $1"
            log_error "用法: ./start.sh {dev|ubuntu|stop|--help}"
            exit 1
            ;;
    esac
}

main "$@"