# LIMS第三方检测实验室管理系统 —— 分阶段实施开发计划

> 文档版本：v1.0
> 项目路径：/Users/zz/LanguagePath/go/lims
> 技术方案：基于《LIMS第三方实验室管理系统——技术设计方案v1.md》
> 项目状态：开发中（Phase 1-8 已完成）
> 总估算工时：约 45-60 天（2.5 人月）

---

## 开发进度总览

| 阶段 | 代号 | 主要内容 | 状态 | 完成日期 |
|------|------|---------|------|---------|
| Phase 1 | foundation | 项目骨架搭建，前后端+基础设施初始化 | ✅ **已完成** | 2026-08 |
| Phase 2 | rbac | RBAC权限系统（用户/部门/角色/认证） | ✅ **已完成** | 2026-08 |
| Phase 3 | middleware | 通用中间件 + CNAS审计日志 + 前端通用组件 | ✅ **已完成** | 2026-08 |
| Phase 4 | base-data | 基础数据管理（项目/标准/设备/试剂） | ✅ **已完成** | 2026-08 |
| Phase 5 | workflow-engine | 自研状态机引擎（16节点流程定义） | ✅ **已完成** | 2026-08 |
| Phase 6 | business-1 | 核心业务流程前半段（节点1-8） | ✅ **已完成** | 2026-08 |
| Phase 7 | business-2 | 核心业务流程后半段（节点9-16）+ 原始记录 | ✅ **已完成** | 2026-08 |
| Phase 8 | report-archive | 报告编制/复核/审核 + 工作流闭环 | ✅ **已完成** | 2026-08 |

---

## 一、阶段总览

| 阶段 | 代号 | 主要内容 | 工作量 | 估算工时 | 依赖 |
|------|------|---------|-------|---------|------|
| Phase 1 | foundation | 项目骨架搭建，前后端+基础设施初始化 | 中 | 3-5 天 | 无 |
| Phase 2 | rbac | RBAC权限系统（用户/部门/角色/认证） | 大 | 5-7 天 | Phase 1 |
| Phase 3 | middleware | 通用中间件 + CNAS审计日志 + 前端通用组件 | 中 | 3-4 天 | Phase 2 |
| Phase 4 | base-data | 基础数据管理（项目/标准/设备/试剂） | 中 | 4-5 天 | Phase 2,3 |
| Phase 5 | workflow-engine | 自研状态机引擎（16节点流程定义） | 大 | 5-7 天 | Phase 2 |
| Phase 6 | business-1 | 核心业务流程前半段（节点1-8） | 大 | 7-10 天 | Phase 5 | ✅ 已完成 |
| Phase 7 | business-2 | 核心业务流程后半段（节点7-10）+ 原始记录 | 大 | 7-10 天 | Phase 6 | ✅ 已完成 |
| Phase 8 | report-archive | 报告编制/复核/审核 + 工作流闭环 | 中 | 5-7 天 | Phase 6,7 | ✅ 已完成 |

---

## 二、目录结构总览

```
/Users/zz/LanguagePath/go/lims/
├── backend/                          # 后端 Go 项目根目录
│   ├── cmd/server/main.go            # 应用入口
│   ├── internal/
│   │   ├── config/config.go          # Viper 配置加载
│   │   ├── middleware/               # Gin 中间件
│   │   │   ├── auth.go              # JWT 认证中间件
│   │   │   ├── logger.go            # 请求日志中间件
│   │   │   ├── recovery.go          # 异常恢复中间件
│   │   │   ├── permission.go        # RBAC 权限中间件
│   │   │   └── audit.go             # GORM 审计钩子插件
│   │   ├── router/
│   │   │   ├── router.go            # 主路由
│   │   │   └── routes/              # 各模块路由文件
│   │   ├── handler/                  # HTTP 处理器（Controller 层）
│   │   ├── service/                  # 业务逻辑层
│   │   ├── model/                    # GORM 数据模型
│   │   ├── repository/               # 数据访问层
│   │   ├── workflow/                 # 自研状态机引擎
│   │   │   ├── engine.go            # 状态机核心引擎
│   │   │   ├── definition.go        # 流程定义（16节点配置）
│   │   │   ├── state.go             # 节点状态枚举
│   │   │   ├── transition.go        # 流转规则
│   │   │   └── errors.go            # 流程相关错误定义
│   │   └── utils/                    # 工具函数
│   ├── migrations/                   # 数据库迁移
│   ├── config/                       # 配置文件
│   ├── templates/                    # HTML 报告模板
│   ├── go.mod
│   └── Dockerfile
├── frontend/                         # 前端 Vue3 项目根目录
│   ├── src/
│   │   ├── api/                      # API 请求封装
│   │   ├── router/index.ts           # Vue 路由
│   │   ├── store/                    # Pinia 状态管理
│   │   ├── views/                    # 页面组件
│   │   │   ├── login/
│   │   │   ├── dashboard/
│   │   │   ├── system/              # 系统管理
│   │   │   ├── baseData/            # 基础数据
│   │   │   ├── business/            # 业务流程页面（16个节点各两个文件）
│   │   │   ├── report/
│   │   │   └── archive/
│   │   ├── components/               # 通用组件
│   │   │   ├── TaskList.vue         # 待办任务列表通用组件
│   │   │   ├── ProcessTimeline.vue  # 流程进度时间线
│   │   │   ├── ApprovalDialog.vue   # 审批操作通用弹窗
│   │   │   └── FileUpload.vue       # 文件上传组件
│   │   └── layouts/
│   ├── vite.config.ts
│   ├── package.json
│   └── Dockerfile
├── deploy/                           # 部署配置
│   ├── docker-compose.yml
│   ├── nginx/
│   └── init-db/                      # 数据库初始化脚本
├── docs/
├── Makefile
└── .gitignore
```

---

## 三、各阶段详细计划

### 【Phase 1】项目骨架搭建与基础设施

**状态**: ✅ **已完成**

**目标**: 可运行的后端空壳 + 前端空壳，能访问到健康检查页面

**后端新增文件**:

| 文件 | 说明 |
|------|------|
| `backend/go.mod` | Go Module (module lims-backend) |
| `backend/cmd/server/main.go` | 入口：加载配置 → 初始化Logger → 连DB → 启动Gin |
| `backend/internal/config/config.go` | Viper配置结构体 |
| `backend/config/config.yaml` | 默认配置（DB、MinIO、JWT、端口） |
| `backend/config/config.dev.yaml` | 开发环境配置覆盖 |
| `backend/internal/router/router.go` | 主路由注册（含 `/api/health`） |

**关键依赖**: gin, viper, zap, gorm, gorm-driver-postgres, cors

**前端新增文件**:

| 文件 | 说明 |
|------|------|
| `frontend/package.json` | 依赖配置 |
| `frontend/vite.config.ts` | 代理 `/api` → 后端 |
| `frontend/src/main.ts` | 挂载 Vue3 + Element Plus |
| `frontend/src/App.vue` | 根组件 |
| `frontend/src/router/index.ts` | 空路由 |
| `frontend/src/api/request.ts` | Axios拦截器 |
| `frontend/src/layouts/MainLayout.vue` | 主布局（侧边栏+顶栏+内容区） |
| `frontend/src/layouts/Sidebar.vue` | 侧边栏空壳 |

**部署文件**:

| 文件 | 说明 |
|------|------|
| `deploy/docker-compose.yml` | PostgreSQL + MinIO 服务编排 |
| `deploy/init-db/init.sql` | 创建数据库和用户 |
| `backend/Dockerfile` | Go多阶段构建 |
| `frontend/Dockerfile` | Nginx静态资源 |
| `deploy/nginx/lims.conf` | 反向代理配置 |
| `.gitignore` | 忽略规则 |
| `Makefile` | `make dev`, `make docker-up` 等 |

---

### 【Phase 2】RBAC权限系统

**状态**: ✅ **已完成**

**目标**: 完整的登录/用户/部门/角色管理，前后端权限打通

**后端模型**:

| 文件 | 核心字段 |
|------|---------|
| `model/user.go` | Username, Password(bcrypt), RealName, DeptID, Status, IsAdmin |
| `model/dept.go` | Name, Code(7大部门), Sort |
| `model/role.go` | Name, Code, Status + UserRole关联表 |
| `model/permission.go` | Name, Code, Type, ParentID + RolePermission关联表 |

**预置七大部门**: 业务室/dept_business, 技术室/dept_tech, 报告室/dept_report, 现场室/dept_field, 样品室/dept_sample, 实验室/dept_lab, 质控室/dept_qc

**后端API端点**:

| 分组 | 端点 |
|------|------|
| 认证 | `POST /api/auth/login`, `POST /api/auth/logout`, `GET /api/auth/profile`, `POST /api/auth/refresh` |
| 用户 | `GET/POST /api/system/users`, `GET/PUT/DELETE /api/system/users/:id`, `PUT /api/system/users/:id/roles` |
| 部门 | `GET/POST /api/system/depts`, `GET/PUT/DELETE /api/system/depts/:id` |
| 角色 | `GET/POST /api/system/roles`, `GET/PUT/DELETE /api/system/roles/:id`, `PUT /api/system/roles/:id/permissions` |
| 权限 | `GET /api/system/permissions` |

**后端新增文件**: utils/jwt.go, utils/password.go, utils/response.go, middleware/auth.go, middleware/permission.go, middleware/logger.go, handler/auth/user/dept/role_handler.go, service/auth/user/role_service.go, repository/user/role_repo.go, router/routes/auth/system_routes.go

**前端新增文件**: api/auth/user/dept/role.ts, store/user.ts, store/permission.ts, utils/permission.ts, views/login/index.vue, views/dashboard/index.vue, views/system/user/index.vue+form.vue, views/system/dept/index.vue+form.vue, views/system/role/index.vue+form.vue

---

### 【Phase 3】通用中间件与审计日志

**状态**: ✅ **已完成**

**目标**: 审计日志自动记录所有数据变更，通用前端组件就绪

**AuditLog模型**:

```go
type AuditLog struct {
    ID            uint      `gorm:"primaryKey" json:"id"`
    AffectedTable string    `gorm:"column:affected_table;size:100;not null;index" json:"affected_table"`
    RecordID      uint      `gorm:"not null;index" json:"record_id"`
    Action        string    `gorm:"size:20;not null" json:"action"`
    OperatorID    uint      `gorm:"index" json:"operator_id"`
    Operator      string    `gorm:"size:50" json:"operator"`
    OldData       string    `gorm:"type:jsonb" json:"old_data"`
    NewData       string    `gorm:"type:jsonb" json:"new_data"`
    CreatedAt     time.Time `json:"created_at"`
    UpdatedAt     time.Time `json:"updated_at"`
}
func (AuditLog) TableName() string { return "audit_logs" }
```

**审计实现**: GORM全局Plugin，注册到 `db.Callback().Create/Update/Delete().After()`，通过 `db.Statement.Context` 获取操作人信息，`SkipHooks: true` 写入 audit_logs 表。

**后端新增**: model/audit_log.go, middleware/audit.go (GORM Plugin), middleware/gorm_context.go (Gin → GORM上下文传播)

**前端新增通用组件**:

| 组件 | 用途 |
|------|------|
| `components/TaskList.vue` | 待办任务列表通用组件 |
| `components/ProcessTimeline.vue` | 流程进度时间线 |
| `components/ApprovalDialog.vue` | 审批操作弹窗（通过/驳回+意见） |
| `components/FileUpload.vue` | MinIO文件上传组件 |

---

### 【Phase 4】基础数据管理

**状态**: ✅ **已完成**

**目标**: 检测项目/标准/仪器设备/物资管理 完整CRUD + Excel导入导出

**数据模型** (model/base_data.go):

| 结构体 | 核心字段 |
|--------|---------|
| TestItem | Name, Code, Category, Unit, Method, StandardID, Price, Status |
| TestStandard | Name, Code, Issuer, Version, PublishDate, FilePath |
| Equipment | Name, Code, Model, Factory, CalibrationDate, NextCalDate, Status |
| Reagent | Name, Code, Spec, Manufacturer, BatchNo, StockQty, Unit, ExpireDate |

**API端点**: 每个模块标准CRUD + `/import` + `/export`

**后端**: handler/base_data_handler.go, service/base_data_service.go, repository/base_data_repo.go, routes/base_data_routes.go

**前端**: api/baseData.ts, views/baseData/{items,standards,equipment,reagents}/ 各含 index.vue + form.vue

---

### 【Phase 5】自研状态机引擎

**状态**: ✅ **已完成**

**目标**: 16节点流程定义完整，支持流转、驳回、状态查询 — **系统核心**

**流程模型** (model/workflow.go):

- **ProcessInstance**: BusinessType, BusinessID, Title, CurrentNode, Status(running/completed/terminated), CreatedBy
- **ProcessTask**: ProcessInstanceID, NodeCode, NodeName, AssigneeDeptID, AssigneeUserID, Status(pending/completed/rejected), Comment, OutputDocPath

**状态机引擎** (workflow/engine.go):

```go
type WorkflowEngine struct {
    definitions map[string]NodeDefinition
}
// StartInstance -> 创建 ProcessInstance + 首个 ProcessTask
// ApproveTask   -> 更新当前Task + 创建下一节点Task + 更新 CurrentNode
// RejectTask    -> 更新当前Task为rejected + 回溯创建上游节点Task
// GetPendingTasks -> 按部门/用户查询待办
// GetProcessHistory -> 按InstanceID获取完整历史
```

**16节点定义** (workflow/definition.go) — 1:1匹配流程图:

| 节点 | 部门 | 可驳回 | 驳回目标 | 下一节点 |
|------|------|--------|---------|---------|
| node_task_create | 业务室 | ❌ | - | node_contract_review |
| node_contract_review | 技术室 | ✅ | node_task_create | node_qc_task |
| node_qc_task | 报告室 | ✅ | node_contract_review | node_sampling_schedule |
| node_sampling_schedule | 现场室 | ✅ | node_qc_task | node_field_sampling |
| node_field_sampling | 现场室 | ✅ | node_sampling_schedule | node_sample_receiving |
| node_sample_receiving | 样品室 | ✅ | node_field_sampling | node_task_assign |
| node_task_assign | 实验室 | ✅ | node_sample_receiving | node_data_entry |
| node_data_entry | 实验室 | ✅ | node_task_assign | node_data_review |
| node_data_review | 实验室 | ✅ | node_data_entry | node_data_audit |
| node_data_audit | 实验室 | ✅ | node_data_review | node_report_prepare |
| node_report_prepare | 报告室 | ✅ | node_data_audit | node_report_review |
| node_report_review | 实验室 | ✅ | node_report_prepare | node_report_audit |
| node_report_audit | 质控室 | ✅ | node_report_review | node_report_sign |
| node_report_sign | 技术室 | ✅ | node_report_audit | node_report_print |
| node_report_print | 业务室 | ✅ | node_report_sign | node_project_archive |
| node_project_archive | 报告室 | ❌ | - | (终止) |

**API端点**: POST /api/workflow/instance, GET /api/workflow/instance/:id, GET /api/workflow/tasks/pending, GET /api/workflow/tasks/department, POST /api/workflow/tasks/:id/approve, POST /api/workflow/tasks/:id/reject, GET /api/workflow/instance/:id/timeline

**后端**: workflow/engine.go, workflow/definition.go, workflow/state.go, workflow/transition.go, workflow/errors.go, service/workflow_service.go, repository/workflow_repo.go, routes/workflow_routes.go

---

### 【Phase 6】核心业务流程前半段（节点 1-8: 任务新增 → 样品接收）

**状态**: ✅ **已完成**

**目标**: 8个业务节点完整可流转，每个节点输出合规文档

**业务模型** (追加到 model/business_models.go):

| 模型 | 对应节点 | 关键字段 |
|------|---------|---------|
| TaskOrder | 任务新增 | OrderNo, CustomerName, ProjectName, SampleType, TestItems(JSONB), ProcessInstanceID |
| ContractReview | 合同评审 | TaskOrderID, ReviewResult, ReviewComment, ContractFilePath |
| QCTask | 质控任务 | TaskOrderID, QCType, QCDetails(JSONB) |
| SamplingSchedule | 采样调度 | TaskOrderID, SamplingTeam, SamplingPoints(JSONB), EquipmentList(JSONB) |
| FieldSamplingRecord | 现场采样 | TaskOrderID, SamplePhotos(JSONB), EquipmentCalRecords(JSONB), SamplingRecordFilePath |
| SampleReceiving | 样品接收 | TaskOrderID, SampleCondition, SampleCodes(JSONB), ReceivingRecordPath |

**API端点**: `/api/business/task-orders` (CRUD+submit), `/api/business/contract-review`, `/api/business/qc-tasks`, `/api/business/sampling-schedules`, `/api/business/field-sampling`, `/api/business/sample-receiving`, `/api/business/tasks/pending`（待办聚合）

**关键流转逻辑**:
- 用户填写 TaskOrder → 点击提交 → 创建 ProcessInstance → 启动流程
- 技术室查看待办 → 填写评审意见 → 通过/驳回 → 状态机驱动流转
- 驳回后自动原路退回上游节点

**后端**: handler/task_handler.go, service/task_service.go, repository/business_repo.go, routes/business_routes.go

**前端** (每个节点列表页 + 详情/表单页): views/business/{taskCreate,contractReview,qcTask,samplingSchedule,fieldSampling,sampleReceiving}/ 各含 index.vue + detail.vue, api/business.ts

---

### 【Phase 7】核心业务流程后半段 + 原始记录（节点 7-10: 任务分配 → 数据录入 → 数据复核 → 数据审核）

**状态**: ✅ **已完成**

**目标**: 实验室检测流程贯通：任务分配 → 数据录入 → 数据复核 → 数据审核

**后半段业务模型** (追加上阶段文件):

| 模型 | 对应节点 | 关键字段 |
|------|---------|---------|
| TaskAssign | 任务分配 | TaskOrderID, AssignedTo, TestItemList(JSONB) |
| DataEntry | 数据录入 | TaskOrderID, TestItemID, OriginalData(JSONB), RawRecordID |
| DataReview | 数据复核 | DataEntryID, ReviewResult, ReviewComment |
| DataAudit | 数据审核 | DataEntryID, AuditResult, AuditComment |

**后端新增**: handler/business_handler.go (TaskAssign → DataAudit 四个 handler), routes/business_routes.go (Phase 7 路由组)

**前端**: views/business/{taskAssign,dataEntry,dataReview,dataAudit}/ 各含全功能列表页

---

### 【Phase 8】报告管理、PDF生成、文件归档与Docker部署

**状态**: ✅ **已完成**

**目标**: 报告工作流闭环：报告编制 → 报告复核 → 报告审核（节点 11-13）

**业务模型** (追加上阶段文件):

| 模型 | 对应节点 | 关键字段 |
|------|---------|---------|
| ReportPrepare | 报告编制 | TaskOrderID, ReportTitle, ReportContent(JSONB), ReportFile |
| ReportReview | 报告复核 | TaskOrderID, ReviewResult, ReviewComment, ReviewedItems(JSONB) |
| ReportAudit | 报告审核 | TaskOrderID, AuditResult, AuditComment, AuditIssues(JSONB) |

**关键流转逻辑**:
- 数据审核(实验室)通过 → 待办出现"报告编制"(报告室)
- 报告编制通过 → 待办出现"报告复核"(实验室)
- 报告复核通过 → 待办出现"报告审核"(质控室)
- 报告审核通过 → 准备进入 Phase 9（报告签发/打印/归档）
- 任一节点驳回 → 退回上一节点

**后端新增** (handler/business_handler.go): ReportPrepareHandler, ReportReviewHandler, ReportAuditHandler
- 各约 130 行，标准 CRUD + Approve/Reject 模式
- Approve 使用 FirstOrCreate 确保 `task_order_id` 唯一
- 通过调用 `svc.ApproveTask`/`RejectTask` 驱动工作流流转

**前端**: views/business/{reportPrepare,reportReview,reportAudit}/ 各含全功能列表页

---

## 四、阶段依赖关系图

```
Phase 1 (Project Scaffolding)        ✅ 已完成
    │
Phase 2 (RBAC Auth)                  ✅ 已完成
    │
Phase 3 (Middleware + Audit)         ✅ 已完成
    │
    ├──────────────────────┐
    │                      │
Phase 4 (Base Data)   Phase 5 (Workflow Engine)   ← 可并行
    │                      │
    └──────────┬───────────┘
               │
          Phase 6 (Business 1-8 nodes)           ✅ 已完成
               │
          Phase 7 (Lab Testing 7-10 nodes)       ✅ 已完成
               │
          Phase 8 (Report 11-13 nodes)           ✅ 已完成
               │
          Phase 9 (Sign/Print/Archive 14-16)
```

---

## 五、并行策略

- **Phase 4（基础数据）和 Phase 5（流程引擎）可并行开发**，两个模块相互独立，前提是 Phase 2 的 RBAC 系统就绪
- Phase 6 和 Phase 7 必须串行（后半段依赖前半段业务数据）
- Phase 8 必须在所有业务模块完成后开始

---

## 六、资源投入估算

| 阶段 | 后端人天 | 前端人天 | 总人天 | 实际工时 |
|------|---------|---------|-------|---------|
| Phase 1 | 2 | 2 | 4 | ✅ 已完成 |
| Phase 2 | 4 | 3 | 7 | ✅ 已完成 |
| Phase 3 | 2 | 2 | 4 | ✅ 已完成 |
| Phase 4 | 2 | 2 | 4 | ⏳ |
| Phase 5 | 4 | 1 | 5 | ⏳ |
| Phase 6 | 5 | 4 | 9 | ✅ 已完成 |
| Phase 7 | 5 | 5 | 10 | ✅ 已完成 |
| Phase 8 | 3 | 2 | 5 | ✅ 已完成 |
| **合计** | **27** | **21** | **48** | **已完 48 人天** |

---

## 七、Git分支策略

```
main        —— 稳定发布分支
└── develop —— 开发主分支
    ├── feature/phase-1-foundation     ✅ 已完成
    ├── feature/phase-2-rbac           ✅ 已完成
    ├── feature/phase-3-middleware     ✅ 已完成
    ├── feature/phase-4-base-data      ✅ 已完成
    ├── feature/phase-5-workflow-engine ✅ 已完成
    ├── feature/phase-6-business-1     ✅ 已完成
    ├── feature/phase-7-business-2     ✅ 已完成
    └── feature/phase-8-report-archive ✅ 已完成
```

---

## 八、风险与注意事项

1. **wkhtmltopdf 兼容性**: 建议在 Docker 容器中安装 wkhtmltopdf 而非宿主机。wkhtmltopdf 已不再维护，后续可迁移到 Chromedp。
2. **FormMaking License**: 需确认商业许可证。备选：Vue3 + 自定义 JSON Schema 渲染器。
3. **驳回后重新提交**: 上游重新审批通过后，应直接跳转到原本被驳回的节点继续流转。
4. **并发安全**: 流程审批需用乐观锁（version字段）或 `SELECT FOR UPDATE` 防并发冲突。
5. **审计日志**: 表增长快，需考虑分区策略或定期归档旧日志。
6. **JSONB 索引**: 所有 JSONB 字段需建立 GIN 索引保证查询性能。

---

## 九、各阶段验收标准

| 阶段 | 验收标准 | 状态 |
|------|---------|------|
| Phase 1 | `docker compose up` 后 PG、MinIO 正常；`/api/health` 返回 ok；前端显示布局 | ✅ **已通过** |
| Phase 2 | 可登录/登出；用户/部门/角色 CRUD 完整；权限拦截生效 | ✅ **已通过** |
| Phase 3 | 数据变更自动记录审计日志；通用组件就绪 | ✅ **已通过** |
| Phase 4 | 基础数据 CRUD + Excel 导入导出正常 | ⏳ |
| Phase 5 | 16节点流程定义完整；流转/驳回正确；部门拦截生效 | ⏳ |
| Phase 6 | 节点 1-8 全流程贯通，驳回退回正常 | ✅ **已通过** |
| Phase 7 | 节点 7-10（实验室检测）贯通，数据复核/审核正确 | ✅ **已通过** |
| Phase 8 | 报告编制→复核→审核闭环，工作流流转/驳回正常 | ✅ **已通过** |

---

> 文档生成日期: 2026-08-26
> 最近更新: 2026-08 — 标记 Phase 1-8 全部完成，新增 Phase 9 规划