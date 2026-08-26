# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

LIMS (Laboratory Information Management System) — a third-party testing lab management system built with Go + Vue 3. The system covers the full testing lifecycle from委托 (commission) through 报告归档 (report archival), designed for CNAS certification compliance.

**Current status**: Phase 1-8 complete (nodes 1-13 of 16). Phase 9 (nodes 14-16: report signing, printing, archiving) is next.

## Commands

```bash
# Infrastructure (PostgreSQL + MinIO)
make docker-up              # Start PostgreSQL (5432) + MinIO (9000/9001)
make docker-down            # Stop all containers

# Backend
make dev-backend            # Run backend with hot reload (localhost:8080)
cd backend && go build ./...  # Compile check
cd backend && go test ./...   # Run tests

# Frontend
make dev-frontend           # Run frontend dev server (localhost:3000)
cd frontend && npx vue-tsc --noEmit  # TypeScript type check
cd frontend && npm run build         # Production build

# Full build
make build                  # Build both backend and frontend
```

## Architecture

### Backend (Go / Gin / GORM)

```
cmd/server/main.go          # Entry: config → logger → DB → seed → router → serve
internal/
├── config/                 # Viper config loading (config.yaml + config.dev.yaml)
├── middleware/             # auth.go (JWT), permission.go (RBAC), audit.go (GORM plugin),
│                         # logger.go, recovery.go, gorm_context.go (Gin→GORM context)
├── model/                  # GORM models — all tables auto-migrated on startup
│   ├── user/dept/role/permission.go  # RBAC models
│   ├── workflow.go                     # ProcessInstance, ProcessTask
│   ├── business_models.go            # All 16 node business models in one file
│   ├── base_data.go                  # TestItem, TestStandard, Equipment, Reagent
│   └── audit_log.go                  # AuditLog (auto-recorded via GORM plugin)
├── handler/               # HTTP handlers (controllers)
│   ├── system/            # RBAC handlers (user/dept/role/permission)
│   ├── auth_handler.go    # login/logout/refresh/profile
│   ├── base_data_handler.go
│   ├── workflow_handler.go  # Workflow instance/task HTTP endpoints
│   └── business_handler.go  # All business node handlers (Phase 6-8)
├── service/               # Business logic
│   ├── workflow_service.go  # Wraps workflow engine for handler use
│   └── business_service.go  # Shared business helpers
├── workflow/              # Custom state machine engine (Phase 5, system core)
│   ├── engine.go          # Engine: StartInstance, ApproveTask, RejectTask, GetPendingTasks
│   ├── definition.go      # 16-node workflow definition (NodeDefinition slice)
│   ├── state.go           # Node code + dept code constants
│   ├── transition.go      # Transition rule logic
│   └── errors.go          # Workflow error types
├── router/
│   ├── router.go          # Main router setup + autoMigrate()
│   └── routes/            # Route registration per module
│       ├── system_routes.go
│       ├── base_data_routes.go
│       ├── workflow_routes.go
│       └── business_routes.go  # All business node routes (Phase 6-8)
├── seed/seed.go           # Initial data: 7 departments + admin user
└── utils/                 # jwt.go, password.go (bcrypt), response.go (unified JSON response)
```

### Frontend (Vue 3 / Vite / Element Plus / TypeScript)

```
src/
├── api/                   # Axios wrappers
│   ├── request.ts         # Axios instance with interceptors (JWT token, /api proxy)
│   ├── auth.ts, system.ts, baseData.ts, workflow.ts, business.ts
├── components/            # Shared components
│   ├── TaskList.vue       # Pending tasks list (used across all nodes)
│   ├── ProcessTimeline.vue # Workflow progress timeline
│   ├── ApprovalDialog.vue # Approve/reject dialog (shared by all nodes)
│   └── FileUpload.vue     # MinIO file upload
├── layouts/
│   ├── MainLayout.vue     # Sidebar + top bar + content area
│   └── Sidebar.vue        # Navigation menu (add new menu items here)
├── router/index.ts        # All routes (add new routes here)
├── store/                 # Pinia stores
├── views/
│   ├── login/             # Login page
│   ├── dashboard/         # Dashboard
│   ├── system/            # RBAC pages (user/dept/role)
│   ├── baseData/          # Base data pages (items/standards/equipment/reagents)
│   └── business/          # Business node pages (one dir per node)
│       ├── taskOrder/     # Node 1: 任务委托
│       ├── contractReview/ # Node 2: 合同评审
│       ├── ...            # Nodes 3-10
│       ├── reportPrepare/ # Node 11: 报告编制
│       ├── reportReview/  # Node 12: 报告复核
│       ├── reportAudit/   # Node 13: 报告审核
│       └── pendingTasks/  # Cross-node pending task view
```

### The 16-Node Workflow (System Core)

The workflow engine (`backend/internal/workflow/`) is a custom state machine that drives the entire business flow. Each node maps to a department and has defined transitions:

```
Node 1  任务委托     业务室   →  Node 2  合同评审     技术室
Node 3  质控任务     报告室   →  Node 4  采样调度     现场室
Node 5  现场采样     现场室   →  Node 6  样品接收     样品室
Node 7  任务分配     实验室   →  Node 8  数据录入     实验室
Node 9  数据复核     实验室   →  Node 10 数据审核     实验室
Node 11 报告编制     报告室   →  Node 12 报告复核     实验室
Node 13 报告审核     质控室   →  Node 14 报告签发     技术室  (Phase 9)
Node 15 报告发放     业务室   →  Node 16 项目归档     报告室  (Phase 9)
```

Most nodes support rejection (退回上一节点). The engine handles:
- `StartInstance()` — creates ProcessInstance + first ProcessTask
- `ApproveTask()` — completes current task, creates next node's task
- `RejectTask()` — marks rejected, creates task at rejection target node
- `GetPendingTasks()` — queries by department/user

### Adding a New Business Node

Each business node follows the same pattern. To add a new node (e.g., Phase 9's node 14-16):

1. **Model** (`model/business_models.go`): Add struct with `TaskOrderID uint gorm:"uniqueIndex"` + business fields
2. **Handler** (`handler/business_handler.go`): Add handler with standard CRUD + `Approve`/`Reject` methods
   - Approve: bind request → build model → `Where("task_order_id = ?", req.TaskID).Assign(...).FirstOrCreate(&record)` → call `svc.ApproveTask()`
   - Reject: bind request → build model → `FirstOrCreate` → call `svc.RejectTask()`
3. **Route** (`router/routes/business_routes.go`): Register CRUD + approve/reject endpoints inside the `group {}` block
4. **Auto-migrate** (`router/router.go`): Add model to `autoMigrate()`
5. **Frontend API** (`api/business.ts`): Add interface + 7 API functions (list/get/create/update/delete/approve/reject)
6. **Frontend page** (`views/business/<nodeName>/index.vue`): Single-file page with list + approval dialog
7. **Frontend route** (`router/index.ts`): Add route entry
8. **Frontend menu** (`layouts/Sidebar.vue`): Add menu item

### Key Patterns

- **Business models**: All in `model/business_models.go`, each with `TaskOrderID` as unique index linking to the workflow
- **Handler pattern**: Each handler holds `svc *service.BusinessService` + `db *gorm.DB`; uses `parseUint(c.Param("id"))` and `middleware.GetUserID(c)`
- **Unified response**: `utils.Success(c, data)` / `utils.Fail(c, code, msg)` / `utils.FailWithMsg(c, msg)`
- **Context propagation**: `middleware.GormContextMiddleware` extracts user info from Gin context into GORM context for audit logging
- **Audit trail**: GORM plugin auto-records all CREATE/UPDATE/DELETE to `audit_logs` table (CNAS compliance)
- **Frontend business pages**: Each is a single-file Vue component (~190 lines) with table + search + approval dialog, following the pattern in `dataAudit/index.vue`

### Database

- PostgreSQL 15+, auto-migrated on first startup
- All JSONB fields store structured business data (test items, sample info, etc.)
- Connection: `localhost:5432`, user `lims`, password `lims123`, database `lims`

### Authentication & Authorization

- JWT tokens with bcrypt password hashing
- RBAC: User → Role → Permission (menu/button/api types)
- 7 predefined departments: `dept_business`, `dept_tech`, `dept_report`, `dept_field`, `dept_sample`, `dept_lab`, `dept_qc`
- Default admin: `admin` / `admin123`

### Infrastructure

- Docker Compose (`deploy/docker-compose.yml`): PostgreSQL + MinIO (file storage, S3-compatible)
- MinIO: API `localhost:9000`, Console `localhost:9001`, user `lims`, password `lims123`
- Production: Nginx reverse proxy serving frontend static + proxying `/api` to backend
