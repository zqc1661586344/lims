# LIMS 第三方检测实验室管理系统

基于 Go + Vue 3 的第三方检测实验室综合管理系统，覆盖从委托登记到报告归档的全流程管理，满足 CNAS 认证合规要求。

> **当前开发进度：Phase 1-8 已完成**（项目骨架 → 报告工作流闭环，全流程16节点贯通）

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

### 业务流程（完整16节点）

```
业务室    ┌──────────┐     ┌──────────────┐     ┌───────────┐
          │ 任务委托  │────▶│  报告打印发放  │◀────│  报告签发  │
          └────┬─────┘     └──────────────┘     └────▲──────┘
               │                                      │
技术室    ┌────▼─────┐     ┌──────────────┐     ┌────┴──────┐
          │ 合同评审  │     │  报告编制     │     │  报告审核  │
          └────┬─────┘     └──────▲───────┘     └────▲──────┘
               │                  │                    │
报告室    ┌────▼─────┐     ┌──────┴───────┐     ┌────┴──────┐
          │ 质控任务  │     │  原始记录录入 │     │  报告复核  │
          └────┬─────┘     └──────▲───────┘     └───────────┘
               │                  │
现场室    ┌────▼─────┐     ┌──────┴───────┐
          │ 采样调度  │     │  现场采样     │
          └────┬─────┘     └──────┬───────┘
               │                  │
样品室    ┌────▼──────────────────▼───────┐
          │      样品接收                  │
          └──────────────┬────────────────┘
                         │
实验室    ┌──────────────▼────────────────┐
          │      任务分配 / 数据录入       │
          └──────────────┬────────────────┘
                         │
          ┌──────────────▼────┐     ┌──────┴──────┐
          │   数据复核        │────▶│  数据审核    │
          └───────────────────┘     └─────────────┘
```

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
| 图标 | Element Plus Icons | 图标库 |

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
│   ├── cmd/server/main.go               # 应用入口
│   ├── config/
│   │   ├── config.yaml                  # 默认配置
│   │   └── config.dev.yaml              # 开发环境配置
│   ├── internal/
│   │   ├── config/config.go             # Viper 配置加载
│   │   ├── handler/
│   │   │   ├── auth_handler.go          # 登录/登出/刷新
│   │   │   ├── base_data_handler.go     # 基础数据 CRUD（Phase 4）
│   │   │   ├── workflow_handler.go      # 工作流 HTTP 接口（Phase 5）
│   │   │   └── system/                  # RBAC 控制器
│   │   │       ├── user_handler.go
│   │   │       ├── dept_handler.go
│   │   │       ├── role_handler.go
│   │   │       └── permission_handler.go
│   │   ├── middleware/
│   │   │   ├── auth.go                  # JWT 认证中间件
│   │   │   ├── permission.go            # RBAC 权限中间件
│   │   │   ├── logger.go                # 请求日志
│   │   │   ├── recovery.go              # 异常恢复
│   │   │   ├── audit.go                 # GORM 审计插件
│   │   │   └── gorm_context.go          # Gin→GORM 上下文传播
│   │   ├── model/
│   │   │   ├── user.go                  # 用户模型
│   │   │   ├── dept.go                  # 部门模型
│   │   │   ├── role.go                  # 角色模型
│   │   │   ├── permission.go            # 权限模型
│   │   │   ├── audit_log.go             # 审计日志模型
│   │   │   ├── base_data.go             # 基础数据模型（Phase 4）
│   │   │   └── workflow.go              # 流程实例/任务模型（Phase 5）
│   │   ├── service/
│   │   │   └── workflow_service.go      # 工作流服务封装（Phase 5）
│   │   ├── workflow/                    # 自研状态机引擎（Phase 5）
│   │   │   ├── engine.go               # 核心引擎（启动/审批/驳回）
│   │   │   ├── definition.go           # 16节点流程定义
│   │   │   ├── state.go                # 节点与部门代码常量
│   │   │   ├── transition.go           # 流转规则逻辑
│   │   │   └── errors.go               # 流程错误定义
│   │   ├── router/
│   │   │   ├── router.go               # 主路由 + 自动迁移
│   │   │   └── routes/
│   │   │       ├── system_routes.go     # RBAC 路由注册
│   │   │       ├── base_data_routes.go  # 基础数据路由（Phase 4）
│   │   │       └── workflow_routes.go   # 工作流路由（Phase 5）
│   │   ├── seed/seed.go                 # 初始数据填充
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
│   │   │   ├── request.ts              # Axios 实例 + 拦截器
│   │   │   ├── auth.ts                 # 认证 API
│   │   │   └── system/                 # RBAC API
│   │   │       ├── user.ts
│   │   │       ├── dept.ts
│   │   │       └── role.ts
│   │   ├── components/                  # 通用业务组件
│   │   │   ├── ApprovalDialog.vue      # 审批弹窗
│   │   │   ├── TaskList.vue            # 待办任务列表
│   │   │   ├── ProcessTimeline.vue     # 流程时间线
│   │   │   └── FileUpload.vue          # 文件上传
│   │   ├── layouts/
│   │   │   ├── MainLayout.vue          # 主布局（侧边栏+顶栏）
│   │   │   └── Sidebar.vue             # 侧边导航
│   │   ├── router/index.ts             # 路由配置
│   │   ├── store/                      # Pinia 状态
│   │   ├── views/
│   │   │   ├── login/index.vue         # 登录页
│   │   │   ├── dashboard/index.vue     # 工作台
│   │   │   └── system/                 # RBAC 管理页
│   │   │       ├── user/index.vue
│   │   │       ├── dept/index.vue
│   │   │       └── role/index.vue
│   │   ├── App.vue
│   │   └── main.ts                     # 应用入口
│   ├── nginx/lims.conf                 # Nginx 配置
│   ├── Dockerfile
│   ├── vite.config.ts
│   └── package.json
├── deploy/
│   ├── docker-compose.yml              # 服务编排
│   ├── init-db/init.sql                # 数据库初始化
│   └── nginx/lims.conf                 # 生产 Nginx 配置
├── docs/
│   └── LIMS第三方实验室管理系统 —— 详细技术设计方案（正式版）.md
├── Makefile                            # 常用命令入口
└── .gitignore
```

---

## 开发阶段完成情况

| 阶段 | 状态 | 内容 |
|------|------|------|
| **Phase 1** 项目骨架 | ✅ **已完成** | Go/Gin 后端骨架、Vue 3 前端骨架、Docker Compose 基础设施、健康检查 |
| **Phase 2** RBAC 权限 | ✅ **已完成** | 用户/部门/角色/权限 CRUD、JWT 认证、密码加密、部门种子数据 |
| **Phase 3** 中间件与审计 | ✅ **已完成** | GORM 审计插件（自动记录数据变更）、Gin↔GORM 上下文传播、审批/待办/时间线/上传通用组件 |
| **Phase 4** 基础数据 | ✅ **已完成** | 检测项目/标准/设备/试剂 CRUD、Excel 导入导出 |
| **Phase 5** 流程引擎 | ✅ **已完成** | 16 节点状态机引擎、流转/驳回规则 |
| **Phase 6-7** 业务流程 | ✅ **已完成** | 16 个业务节点前后端实现、全流程贯通 |
| **Phase 8** 报告工作流 | ✅ **已完成** | 报告编制/复核/审核节点、工作流闭环 |

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
# 方式一：使用 Makefile
make dev-backend

# 方式二：直接运行
cd backend && go run ./cmd/server
```

后端服务默认监听 `localhost:8080`。

> 首次启动会自动执行数据库迁移（创建表结构）和种子数据填充（7大部门 + admin 用户）。

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

# 前端类型检查
cd frontend && npx vue-tsc --noEmit
```

---

## API 概览

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

---

## 权限系统设计

### 7 大预设部门

| 部门 | 编码 | 流程职责 |
|------|------|---------|
| 业务室 | `dept_business` | 任务委托、报告打印发放 |
| 技术室 | `dept_tech` | 合同评审、报告签发 |
| 报告室 | `dept_report` | 质控任务、报告编制 |
| 现场室 | `dept_field` | 采样调度、现场采样 |
| 样品室 | `dept_sample` | 样品接收 |
| 实验室 | `dept_lab` | 任务分配、数据录入/复核/审核、报告复核 |
| 质控室 | `dept_qc` | 报告审核 |

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

```
Phase 1 (项目骨架)
   │
Phase 2 (RBAC 权限)
   │
Phase 3 (中间件 + 审计日志)
   │
   ├──────────────────┐
   │                  │
Phase 4 (基础数据)  Phase 5 (流程引擎)   ← 可并行开发
   │                  │
   └────────┬─────────┘
            │
       Phase 6 (业务流程前半段 节点 1-8)
            │
       Phase 7 (实验室检测流程 节点 7-10)
            │
       Phase 8 (报告工作流 节点 11-13)  ✅ 当前完成
            │
       Phase 9 (报告签发/打印/归档 节点 14-16)
```

---

## License

MIT