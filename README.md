# LIMS 第三方检测实验室管理系统

基于 Go + Vue 3 的第三方检测实验室综合管理系统，覆盖从委托登记到报告归档的全流程管理，满足 CNAS 认证合规要求。

> **当前开发进度：Phase 1-9 已完成，Phase 10（检验单集成）进行中**（项目骨架 → 全流程 16 节点贯通 → Univer Sheet 检验单编辑）

---

## 系统架构

```
┌─────────────────────────────────────────────────────────────┐
│                        前端 (Vue 3)                          │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────────┐  │
│  │ Element  │ │  Vue     │ │  Pinia   │ │  Axios       │  │
│  │ Plus UI  │ │  Router  │ │  State   │ │  HTTP Client │  │
│  └──────────┘ └──────────┘ └──────────┘ └──────────────┘  │
└──────────────────────┬──────────────────────────────────────┘
                       │ HTTP / REST API
                       │ Bearer JWT Token
                       ▼
┌─────────────────────────────────────────────────────────────┐
│                      后端 (Go / Gin)                         │
│                                                             │
│  ┌────────────┐ ┌──────────────┐ ┌──────────────────┐     │
│  │           │ │              │ │                  │     │
│  │ JWT Auth  │ │ RBAC 权限    │ │ 审计日志 (GORM   │     │
│  │ 中间件     │ │ 中间件       │ │ Plugin)          │     │
│  │           │ │              │ │                  │     │
│  └────────────┘ └──────────────┘ └──────────────────┘     │
│                                                             │
│  ┌────────────┐ ┌──────────────┐ ┌──────────────────┐     │
│  │ Handler    │ │ Service      │ │ GORM ORM         │     │
│  │ (Controller)│ │ (Business)   │ │ (Data Access)    │     │
│  └────────────┘ └──────────────┘ └──────────────────┘     │
│                                                             │
│  ┌────────────────────────────────────────────────────┐     │
│  │   工作流引擎 (Phase 5 已完成)                      │     │
│  │   16节点状态机 / 流转 / 驳回 / 历史追溯            │     │
│  └────────────────────────────────────────────────────┘     │
└──────────────────────┬──────────────────────────────────────┘
                       │
          ┌────────────┼────────────┐
          ▼            ▼            ▼
   ┌──────────┐ ┌──────────┐ ┌──────────┐
   │PostgreSQL │ │  MinIO   │ │  Redis   │
   │ (数据)    │ │ (文件)   │ │ (待集成) │
   └──────────┘ └──────────┘ └──────────┘
```

### 业务流程（16 节点全流程）

```
业务室(1 任务委托) → 技术室(2 合同评审) → 质控室(3 质控任务) → 现场室(4 采样调度)
    → 现场室(5 现场采样) → 样品室(6 样品接收) → 实验室(7-10 任务分配/数据录入/数据复核/数据审核)
    → 报告室(11 报告编制) → 实验室(12 报告复核) → 质控室(13 报告审核) → 技术室(14 报告签发)
    → 业务室(15 报告发放) → 报告室(16 项目归档)
```

| 步骤 | 节点 | 责任部门 | 说明 | 产出文档 |
|---|---|---|---|---|
| 1 | 任务委托 | 业务室 | 录入客户/项目/样品/检测项目，启动流程 | D1 委托任务单 |
| 2 | 合同评审 | 技术室 | 评审合同/协议 | D2 检测合同/协议 |
| 3 | 质控任务 | 质控室 | 检测任务中加入质控方案 | D3 委托检测方案 |
| 4 | 采样调度 | 现场室 | 安排采样团队/点位/设备 | — |
| 5 | 现场采样 | 现场室 | 现场采样 | D4 现场采样记录及设备校准记录 |
| 6 | 样品接收 | 样品室 | 接收样品、编码登记 | D5 样品接收记录 |
| 7 | 任务分配 | 实验室 | 分配检测项目给实验员 | — |
| 8 | 数据录入 | 实验室 | 录入检测数据 | D6 实验原始记录 |
| 9 | 数据复核 | 实验室 | 复核数据 | D7 实验原始记录 |
| 10 | 数据审核 | 实验室 | 审核数据 | D8 实验原始记录 |
| 11 | 报告编制 | 报告室 | 编制报告 | D9 报告 + 实验原始记录 |
| 12 | 报告复核 | 实验室 | 复核报告 | D10 报告 + 实验原始记录 |
| 13 | 报告审核 | 质控室 | 审核报告 | D11 报告 + 实验原始记录 |
| 14 | 报告签发 | 技术室 | 签发盖章 | D12 报告 + 原始记录、D15 报告审核签发单 |
| 15 | 报告发放 | 业务室 | 打印发放给客户 | D13 报告发放记录 |
| 16 | 项目归档 | 报告室 | 归档 D1-D13 全部文档 | D14 以上所有文档 |

> 详细流转说明见 [docs/business-workflow.md](docs/business-workflow.md)，流程图见 [docs/process.md](docs/process.md)。

---

## 技术栈

### 后端

| 组件 | 技术 | 用途 |
|------|------|------|
| 语言 | Go 1.21+ | 主开发语言 |
| 框架 | Gin v1.9 | HTTP 路由框架 |
| ORM | GORM v2 + pgx v5 | 数据库操作与迁移 |
| 数据库 | PostgreSQL 15+ | 关系型数据库 |
| 认证 | golang-jwt v5 + bcrypt | JWT 身份认证 |
| 配置 | Viper | 配置管理 |
| 日志 | Zap | 结构化日志 |
| 权限 | 自定义 RBAC | 基于用户/角色/权限的访问控制 |
| 审计 | GORM Plugin | 自动记录数据变更审计日志 |

### 前端

| 组件 | 技术 | 用途 |
|------|------|------|
| 框架 | Vue 3 + Vite 5 | 前端 SPA 框架 |
| UI 组件 | Element Plus 2.7 | 企业级 UI 组件库 |
| 状态管理 | Pinia 2.1 | 状态管理 |
| 路由 | Vue Router 4 | 前端路由 |
| 语言 | TypeScript 5.4 | 类型安全 |
| HTTP | Axios 1.7 | HTTP 客户端 |
| 日期 | Day.js | 日期处理 |
| 图标 | Element Plus Icons | 图标库 |
| 表格编辑 | Univer Sheet v0.5 | Excel 式检验单编辑、公式计算、模板管理 |

### 基础设施

| 组件 | 版本 | 用途 |
|------|------|------|
| Docker Compose | 3.8 | 服务编排 |
| PostgreSQL | 15 Alpine | 主数据库 |
| MinIO | Latest | 文件存储 (S3 兼容) |
| Nginx | Alpine | 前端静态资源 / 反向代理 |

---

## 项目目录结构

```
/Users/zz/LanguagePath/go/lims/
├── backend/                              # Go 后端项目
│   ├── cmd/server/main.go               # 应用入口：config → logger → DB → seed → router → serve
│   ├── config/
│   │   ├── config.yaml                  # 默认配置
│   │   └── config.dev.yaml              # 开发环境配置
│   ├── internal/
│   │   ├── config/config.go             # Viper 配置加载
│   │   ├── middleware/
│   │   │   ├── auth.go                  # JWT 认证中间件
│   │   │   ├── permission.go            # RBAC 权限中间件
│   │   │   ├── logger.go                # 请求日志
│   │   │   ├── recovery.go              # 异常恢复
│   │   │   ├── audit.go                 # GORM 审计插件
│   │   │   └── gorm_context.go          # Gin→GORM 上下文传播
│   │   ├── model/
│   │   │   ├── user.go / dept.go / role.go / permission.go   # RBAC 模型
│   │   │   ├── audit_log.go             # 审计日志模型
│   │   │   ├── base_data.go             # 基础数据模型（检测项目/标准/设备/试剂）
│   │   │   ├── workflow.go              # 流程实例/任务模型
│   │   │   ├── business_models.go       # 16 节点业务模型
│   │   │   └── lab_sheet.go             # 🆕 LabSheet + LabSheetTemplate 模型（Univer Sheet 存储）
│   │   ├── handler/
│   │   │   ├── auth_handler.go          # 登录/登出/刷新/profile
│   │   │   ├── base_data_handler.go     # 基础数据 CRUD
│   │   │   ├── workflow_handler.go      # 工作流 HTTP 接口
│   │   │   ├── lab_sheet_handler.go     # 🆕 检验单 CRUD + 模板管理
│   │   │   ├── {node}_handler.go        # 16 节点各自独立 handler
│   │   │   └── system/                  # RBAC 控制器（user/dept/role/permission）
│   │   ├── service/
│   │   │   ├── workflow_service.go      # 工作流服务封装
│   │   │   └── business_service.go      # 业务服务封装
│   │   ├── workflow/                    # 自研状态机引擎（系统核心）
│   │   │   ├── engine.go               # 核心引擎（启动/审批/驳回/待办）
│   │   │   ├── definition.go           # 16 节点流程定义
│   │   │   ├── state.go                # 节点与部门代码常量
│   │   │   ├── transition.go           # 流转规则逻辑
│   │   │   └── errors.go               # 流程错误定义
│   │   ├── router/
│   │   │   ├── router.go               # 主路由 + 自动迁移
│   │   │   └── routes/
│   │   │       ├── system_routes.go     # RBAC 路由注册
│   │   │       ├── base_data_routes.go  # 基础数据路由
│   │   │       ├── workflow_routes.go   # 工作流路由
│   │   │       ├── business_routes.go   # 16 业务节点路由
│   │   │       └── lab_sheet_routes.go  # 🆕 检验单 + 模板路由
│   │   ├── seed/seed.go                 # 初始数据：7 部门 + admin 用户
│   │   └── utils/
│   │       ├── jwt.go                   # JWT 工具
│   │       ├── password.go              # bcrypt 密码
│   │       └── response.go              # 统一响应格式
│   ├── Dockerfile                       # 多阶段构建
│   ├── go.mod
│   └── go.sum
├── frontend/                            # Vue 3 前端项目
│   ├── src/
│   │   ├── api/
│   │   │   ├── request.ts              # Axios 实例 + 拦截器（JWT token、/api 代理）
│   │   │   ├── auth.ts / system.ts     # 认证 / RBAC API
│   │   │   ├── baseData.ts             # 基础数据 API
│   │   │   ├── workflow.ts             # 工作流 API
│   │   │   └── business.ts             # 16 业务节点 API（Phase 6-9）
│   │   ├── components/                  # 通用业务组件
│   │   │   ├── ApprovalDialog.vue      # 审批弹窗（所有节点共用）
│   │   │   ├── TaskList.vue            # 待办任务列表
│   │   │   ├── ProcessTimeline.vue     # 流程时间线
│   │   │   ├── FileUpload.vue          # 文件上传
│   │   │   └── LabSheetEditor.vue      # 🆕 通用 Univer Sheet 编辑器（支持 edit/readonly 模式）
│   │   ├── layouts/
│   │   │   ├── MainLayout.vue          # 主布局（侧边栏+顶栏）
│   │   │   └── Sidebar.vue             # 侧边导航（新菜单加这里）
│   │   ├── router/index.ts             # 路由配置（新路由加这里）
│   │   ├── store/                      # Pinia 状态（预留，当前为空目录）
│   │   ├── views/
│   │   │   ├── login/index.vue         # 登录页
│   │   │   ├── dashboard/index.vue     # 工作台
│   │   │   ├── system/                 # RBAC 管理页（user/dept/role）
│   │   │   ├── baseData/               # 基础数据页（items/standards/equipment/reagents）
│   │   │   ├── business/               # 16 业务节点页 + labSheetEditor/（检验单全屏编辑，供数据录入/复核/审核等节点复用）
│   │   │   └── App.vue / main.ts           # 组件与应用入口
│   ├── nginx/lims.conf                 # Nginx 配置
│   ├── Dockerfile
│   ├── vite.config.ts
│   └── package.json
├── deploy/
│   ├── docker-compose.yml              # 服务编排（PostgreSQL + MinIO）
│   ├── init-db/init.sql                # 数据库初始化
│   └── nginx/lims.conf                 # 生产 Nginx 配置
├── docs/
│   ├── process.md                      # 业务流程 Mermaid 图
│   ├── business-workflow.md            # 检测管理流转说明
│   ├── 16-node-flow-quickref.md        # 16 节点快速参考
│   ├── development-plan.md             # 开发计划
│   ├── univer-sheet-integration-plan.md # 🆕 Univer Sheet 集成方案
│   └── LIMS第三方实验室管理系统技术设计方案v1.md  # 技术设计方案
├── Makefile                            # 常用命令入口
└── .gitignore
```

> **前端空目录说明**：`frontend/src/views/dashboards/`、`frontend/src/views/report/`、`frontend/src/views/archive/`、`frontend/src/store/` 为**预留空目录**（开发过程中被 `business` 目录取代，当前无业务代码），可按需清理或后续使用。

---

## 开发阶段完成情况

| 阶段 | 状态 | 内容 |
|------|------|------|
| **Phase 1** 项目骨架 | ✅ **已完成** | Go/Gin 后端骨架、Vue 3 前端骨架、Docker Compose 基础设施、健康检查 |
| **Phase 2** RBAC 权限 | ✅ **已完成** | 用户/部门/角色/权限 CRUD、JWT 认证、密码加密、部门种子数据 |
| **Phase 3** 中间件与审计 | ✅ **已完成** | GORM 审计插件、Gin↔GORM 上下文传播、审批/待办/时间线/上传通用组件 |
| **Phase 4** 基础数据 | ✅ **已完成** | 检测项目/标准/设备/试剂 CRUD、Excel 导入导出 |
| **Phase 5** 流程引擎 | ✅ **已完成** | 16 节点状态机引擎、流转/驳回规则 |
| **Phase 6** 业务流程前半段 | ✅ **已完成** | 节点 1-8（委托→数据录入）前后端实现 |
| **Phase 7** 实验室检测流程 | ✅ **已完成** | 节点 7-10（任务分配/数据录入/复核/审核） |
| **Phase 8** 报告工作流 | ✅ **已完成** | 节点 11-13（报告编制/复核/审核） |
| **Phase 9** 报告签发/打印/归档 | ✅ **已完成** | 节点 14-16（报告签发/发放/归档），全流程 16 节点贯通 |
| **Phase 10** 检验单集成 | 🔶 **进行中** | Univer Sheet 集成、LabSheet/LabSheetTemplate 模型、检验单全屏编辑页、数据录入节点对接（复核/审核节点只读模式待接入） |

---

## 快速开始

### 环境要求

- Go 1.21+
- Node.js 18+ / npm 9+
- Docker & Docker Compose
- PostgreSQL 15+（本地开发时）

### 1. 启动基础设施（PostgreSQL + MinIO）

```bash
# 使用 Docker Compose 启动数据库和文件存储
make docker-up
# 或: docker compose -f deploy/docker-compose.yml up -d
```

这会启动：
- **PostgreSQL**: `localhost:5432`（用户: `lims`, 密码: `lims123`, 数据库: `lims`）
- **MinIO**: API `localhost:9000`, Console `localhost:9001`（用户: `lims`, 密码: `lims123`）

### 2. 启动后端服务

```bash
# 方式一：使用 Makefile（air 热重载）
make dev-backend

# 方式二：直接运行（代码改动需手动重启）
cd backend && go run ./cmd/server
```

后端服务默认监听 `localhost:8080`。

> 首次启动会自动执行数据库迁移（创建表结构）和种子数据填充（7 大部门 + admin 用户）。

### 3. 启动前端开发服务器

```bash
# 方式一：使用 Makefile
make dev-frontend

# 方式二：直接运行
cd frontend && npm run dev
```

前端开发服务器默认监听 `localhost:3000`，已配置 `/api` 代理到后端 `localhost:8080`。

### 4. 访问系统

- **前端页面**: http://localhost:3000
- **健康检查**: http://localhost:8080/api/health

### 默认登录账号

| 用户名 | 密码 | 角色 |
|--------|------|------|
| `admin` | `admin123` | 系统管理员（拥有所有权限） |

> 登录后请在系统中创建具体操作人员的账号并分配角色。

---

## Docker 完整部署

一键构建并启动所有服务：

```bash
# 构建镜像
make docker-build

# 启动所有服务（后端 + 前端 + PostgreSQL + MinIO）
make docker-up
```

访问 http://localhost （Nginx 反向代理到前端）。

停止服务：

```bash
make docker-down
```

---

## 常用命令

```bash
# 开发
make dev-backend       # 启动后端（热重载）
make dev-frontend      # 启动前端（热重载）

# 构建
make build-backend     # 编译后端
make build-frontend    # 构建前端
make build             # 构建所有

# Docker
make docker-up         # 启动所有容器
make docker-down       # 停止所有容器
make docker-build      # 构建 Docker 镜像

# 后端测试
cd backend && go test ./...

# 后端编译检查
cd backend && go build ./...

# 前端类型检查
cd frontend && npx vue-tsc --noEmit
```

---

## API 概览

> 除健康检查外，所有接口均需在请求头携带 `Authorization: Bearer <token>`。

### 健康检查

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/health` | 系统健康检查 |

### 认证

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/auth/login` | 用户登录（返回 JWT Token） |
| GET | `/api/auth/profile` | 获取当前用户信息 |
| POST | `/api/auth/logout` | 登出 |
| POST | `/api/auth/refresh` | 刷新 Token |

### 系统管理 (RBAC)

| 方法 | 路径 | 说明 |
|------|------|------|
| GET/POST | `/api/system/users` | 用户列表 / 创建用户 |
| GET/PUT/DELETE | `/api/system/users/:id` | 用户详情 / 更新 / 删除 |
| PUT | `/api/system/users/:id/roles` | 分配用户角色 |
| GET/POST | `/api/system/depts` | 部门列表 / 创建部门 |
| GET/PUT/DELETE | `/api/system/depts/:id` | 部门详情 / 更新 / 删除 |
| GET/POST | `/api/system/roles` | 角色列表 / 创建角色 |
| GET/PUT/DELETE | `/api/system/roles/:id` | 角色详情 / 更新 / 删除 |
| PUT | `/api/system/roles/:id/permissions` | 分配角色权限 |
| GET | `/api/system/permissions` | 权限树列表 |

### 基础数据 (Base Data)

| 资源 | 路径前缀 | 说明 |
|------|------|------|
| 检测项目 | `/api/base-data/items` | 项目 CRUD |
| 检测标准 | `/api/base-data/standards` | 标准 CRUD |
| 设备 | `/api/base-data/equipment` | 设备 CRUD |
| 试剂 | `/api/base-data/reagents` | 试剂 CRUD |

### 工作流 (Workflow)

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/workflow/tasks/pending` | 部门待办列表 |
| GET | `/api/workflow/tasks/pending/user` | 我的待办列表 |
| POST | `/api/workflow/tasks/:id/approve` | 审批通过（推进到下一节点） |
| POST | `/api/workflow/tasks/:id/reject` | 驳回（退回目标节点） |
| GET | `/api/workflow/instances/:id/history` | 流程历史 |
| GET | `/api/workflow/instances/:id` | 流程实例详情 |
| GET | `/api/workflow/nodes` | 节点定义列表 |

### 业务流程 (Business，16 节点)

每个节点一组 CRUD + Approve/Reject 接口，统一路径前缀 `/api/business/<资源名>`：

| 节点 | 资源名 | 说明 |
|------|--------|------|
| 1 任务委托 | `task-orders` | 含 `POST /:id/submit` 提交启动流程 |
| 2 合同评审 | `contract-reviews` | |
| 3 质控任务 | `qc-tasks` | |
| 4 采样调度 | `sampling-schedules` | |
| 5 现场采样 | `field-sampling` | |
| 6 样品接收 | `sample-receiving` | |
| 7 任务分配 | `task-assign` | |
| 8 数据录入 | `data-entry` | |
| 9 数据复核 | `data-review` | |
| 10 数据审核 | `data-audit` | |
| 11 报告编制 | `report-prepare` | 含报告编号/编制意见 |
| 12 报告复核 | `report-review` | |
| 13 报告审核 | `report-audit` | |
| 14 报告签发 | `report-sign` | 自动汇聚报告审核签发单 |
| 15 报告发放 | `report-print` | |
| 16 项目归档 | `project-archive` | 自动汇聚 D1-D13 归档清单 |

> 节点 2-16 均支持 `GET /:id/approve`（通过）与 `GET /:id/reject`（驳回）。任务委托（节点 1）是流程起点无审批；项目归档（节点 16）是终节点不可驳回。

### 检验单 (Lab Sheet / Univer Sheet)

检验单基于 Univer Sheet 实现 Excel 式编辑，每个检测项目可关联独立模板。

**模板管理：**

| 方法 | 路径 | 说明 |
|------|------|------|
| GET/POST | `/api/lab-sheets/templates` | 模板列表 / 创建模板 |
| GET/PUT/DELETE | `/api/lab-sheets/templates/:id` | 模板详情 / 更新 / 删除 |
| GET | `/api/lab-sheets/templates/by-item/:testItemID` | 按检测项目查询模板 |

**检验单实例：**

| 方法 | 路径 | 说明 |
|------|------|------|
| GET/POST | `/api/lab-sheets` | 检验单列表 / 创建（参数: task_order_id, test_item_id, node_code, sheet_data） |
| GET/PUT/DELETE | `/api/lab-sheets/:id` | 详情 / 更新（sheet_data, formula_results, status） / 删除 |

> `sheet_data` 为 Univer Sheet Workbook JSON（jsonb 字段），包含完整的表格结构、单元格数据和公式。

---

## 权限系统设计

### 7 大预设部门

| 部门 | 编码 | 流程职责 |
|------|------|---------|
| 业务室 | `dept_business` | 任务委托、报告发放 |
| 技术室 | `dept_tech` | 合同评审、报告签发 |
| 质控室 | `dept_qc` | 质控任务、报告审核 |
| 现场室 | `dept_field` | 采样调度、现场采样 |
| 样品室 | `dept_sample` | 样品接收 |
| 实验室 | `dept_lab` | 任务分配、数据录入/复核/审核、报告复核 |
| 报告室 | `dept_report` | 报告编制、项目归档 |

### RBAC 模型

```
用户 (User) ──── N:N ──── 角色 (Role) ──── N:N ──── 权限 (Permission)
  │                                                      │
  └── 部门 (Dept)                                ├── 菜单权限 (menu)
                                                 ├── 按钮权限 (button)
                                                 └── API 权限 (api)
```

---

## 审计日志

系统通过 GORM Plugin 自动记录所有数据变更操作：

- **CREATE**: 记录新增数据的完整内容
- **UPDATE**: 记录修改前后的数据对比（JSONB 格式）
- **DELETE**: 记录被删除的数据

审计日志包含操作人、操作时间、操作表、记录 ID 等完整信息，满足 CNAS 认证对数据可追溯性的要求。

---

## 后续开发路线

全流程 16 节点（Phase 1-9）已全部完成，检验单集成（Phase 10）进行中。

```
✅ Phase 1-9 全部完成（委托 → 报告签发/发放 → 项目归档）
🔶 Phase 10 检验单集成（进行中）
   ├── ✅ Univer Sheet 组件封装（LabSheetEditor.vue）
   ├── ✅ LabSheet / LabSheetTemplate 数据模型 + API
   ├── ✅ 数据录入节点 → 检验单全屏编辑页（/business/lab-sheet-editor）
   ├── 📝 数据复核节点 → 检验单只读模式 + 批注
   ├── 📝 数据审核节点 → 检验单只读模式 + 审核意见
   └── 📝 检验单模板管理页面

可扩展方向：
   ├── Redis 集成（缓存 / 分布式会话）
   ├── 检测报告电子签章对接
   ├── 客户委托自助下单 / 进度查询
   ├── 消息通知（待办提醒）
   └── 多实验室 / 多分支机构支持
```

---

## License

MIT