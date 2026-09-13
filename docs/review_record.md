## LIMS 系统代码评审报告

### 问题 1（严重·安全性）：JWT Token 存储于 localStorage，存在 XSS 攻击风险

**代码位置**：`frontend/src/api/request.ts` 第 11 行

```typescript
const token = localStorage.getItem('token') || sessionStorage.getItem('token')
```

以及 401 处理中的清理逻辑（第 28-29 行）。

**问题描述**：JWT Token 存储在 `localStorage`/`sessionStorage` 中，任何 XSS 注入攻击都能直接读取 Token 并冒充用户。LIMS 系统涉及大量敏感检测数据和客户信息，这是高风险漏洞。

**改进建议**：
- 将 Token 存储在 `HttpOnly` + `Secure` + `SameSite=Strict` 的 Cookie 中
- 后端登录接口改为 `Set-Cookie` 方式下发 Token
- 配合 CSRF Token 防御跨站请求伪造
- 如坚持 Bearer 方案，至少在前端做 CSP（Content Security Policy）严格限制

---

### 问题 2（严重·安全性）：Token Blacklist 仅存于内存，服务重启后失效

**代码位置**：`backend/internal/utils/token_blacklist.go`（从目录结构可见），`auth.go` 第 36-39 行调用 `utils.GetTokenBlacklist().IsRevoked(claims.ID)`

**问题描述**：Token 黑名单使用内存存储（进程级 Map），服务重启后所有已注销的 Token 重新生效。攻击者如果获取了一个已 logout 但未过期的 Token，在服务重启后仍可使用。

**改进建议**：
- 将黑名单持久化到 Redis（已在架构图中预留 Redis，应尽快集成）
- 或使用短过期时间（如 15 分钟）的 Access Token + Refresh Token 双 Token 方案
- 黑名单至少应写入数据库或 Redis，设置 TTL = Token 剩余有效期

---

### 问题 3（严重·安全性）：JWT 中嵌入的权限列表不会随角色变更实时更新

**代码位置**：`backend/internal/utils/jwt.go` 第 30-42 行 `GenerateToken` 函数，`claims.Permissions` 在 Token 生成时写入后不再更新。`backend/internal/middleware/permission.go` 第 25-33 行直接从 JWT claims 读取权限。

**问题描述**：用户的权限列表在登录时写入 JWT，此后即使管理员修改了该用户的角色/权限，在 Token 过期前旧权限仍然有效。这意味着：
1. 被降权/禁用的用户在 Token 过期前仍能访问受限资源
2. 新增权限在重新登录前不生效

**改进建议**：
- 权限校验改为每次请求实时查询数据库（或 Redis 缓存）
- 或引入 Token 版本号（`token_version` 字段），角色变更时递增版本号，中间件校验版本号
- 对于 LIMS 这种安全敏感系统，实时权限校验是必须的

---

### 问题 4（严重·安全性）：Admin 用户绕过所有权限校验

**代码位置**：`backend/internal/middleware/permission.go` 第 18-21 行

```go
if IsAdmin(c) {
    c.Next()
    return
}
```

**问题描述**：Admin 用户直接跳过 `PermissionMiddleware` 的所有权限检查。虽然注释中说 "DeptScopeMiddleware still limits admins"，但在 `permission.go` 中看不到 DeptScope 的集成调用。这意味着 Admin 可以操作任何部门的数据，违反了 CNAS 对职责分离（SoD）的要求。

**改进建议**：
- Admin 也应受到部门范围限制，除非是"系统管理类"操作（用户/角色管理）
- 业务操作（审批、数据录入等）应基于部门归属进行隔离，即使是 Admin
- 在路由注册中，确认 `DeptScopeMiddleware` 确实应用于所有业务路由

---

### 问题 5（重要·正确性）：工作流驳回只支持回退一级，不支持跨级驳回

**代码位置**：`backend/internal/workflow/definition.go` 每个节点的 `RejectTarget` 字段均指向前一个节点。例如节点 11（报告编制）驳回只能回到节点 10（数据审核），而实际业务中可能需要从报告编制直接驳回到数据录入。

`backend/internal/workflow/engine.go` 的 `RejectTaskWithTx` 函数第 207-210 行：

```go
rejectTarget := nodeDef.RejectTarget
if len(rejectTargetOverride) > 0 && rejectTargetOverride[0] != "" {
    rejectTarget = rejectTargetOverride[0]
}
```

虽然支持 `rejectTargetOverride` 参数，但前端和 Handler 层并未传递此参数，实际效果是固定回退一级。

**问题描述**：在真实 LIMS 场景中，报告审核阶段发现数据录入错误时，需要直接退回到数据录入节点，而非逐级回退。当前设计导致效率低下且中间节点可能产生无意义的空审批。

**改进建议**：
- 前端审批弹窗增加"驳回目标节点"选择下拉框（列出可回退的已完成节点）
- Handler 层传递用户选择的 `rejectTarget` 到 Engine
- 驳回后，被跳过的中间节点的任务状态应标记为 `skipped` 而非保持 `pending`

---

### 问题 6（重要·正确性）：DataEntry 模型与 TaskOrder 的关系设计不一致

**代码位置**：`backend/internal/model/business_models.go`

- `DataEntry` 结构体（约第 107 行）使用 `gorm:"index"` 而非 `uniqueIndex`
- 其他 15 个节点模型（如 `ContractReview`、`QCTask` 等）全部使用 `uniqueIndex`

**问题描述**：`DataEntry` 使用普通 index 意味着一个 TaskOrder 可以有多条数据录入记录（一对多），但其他节点全部是一对一。这个设计意图可能是为了支持"每个检测项目一条录入记录"，但：
1. `preMigrateCleanup` 中的清理 SQL 假定所有子表与 task_orders 是引用关系，未区分一对多
2. 前端和业务逻辑层是否正确处理了一对多场景不确定
3. 其他节点如 `DataReview` 和 `DataAudit` 又使用了 `uniqueIndex`（一对一），与 DataEntry 的多条记录无法对齐

**改进建议**：
- 如果 DataEntry 是一对多，则 DataReview/DataAudit 也应相应调整为一对多，或引入中间聚合层
- 在业务模型层明确标注每个节点与 TaskOrder 的关系（1:1 或 1:N）
- 添加对应的数据库约束确保数据一致性

---

### 问题 7（重要·架构）：职责分离（SoD）检查覆盖不完整

**代码位置**：`backend/internal/workflow/engine.go` 第 23-28 行

```go
var sodCheckNodes = map[string]string{
    NodeDataReview:  NodeDataEntry,
    NodeDataAudit:   NodeDataReview,
    NodeReportReview: NodeReportPrepare,
    NodeReportAudit:  NodeReportReview,
}
```

以及 `checkSoD` 函数（约第 580 行）。

**问题描述**：SoD 只覆盖了 4 对节点（数据录入→复核、复核→审核、报告编制→复核、报告复核→审核）。但 CNAS 认证要求的职责分离范围更广，至少还应包括：
- 样品接收人员不能是采样人员（节点 5→6）
- 合同评审人员不能是任务委托人（节点 1→2）
- 报告签发人员不能是报告编制人员（节点 11→14）

**改进建议**：
- 扩展 `sodCheckNodes` 映射，覆盖所有 CNAS 要求的职责分离点
- 考虑更灵活的 SoD 规则配置（如数据库可配置），而非硬编码

---

### 问题 8（重要·架构）：查询方法返回 `[]map[string]interface{}` 而非类型安全结构体

**代码位置**：`backend/internal/workflow/engine.go` 的 `queryTasks` 函数（约第 640 行）、`GetPendingTasksByUser`、`GetPendingTasksByDept`、`GetPendingTasksByUser`、`GetProcessHistory` 等函数。

**问题描述**：核心查询全部使用 `map[string]interface{}` 返回，丧失了 Go 的类型安全优势。这导致：
1. 字段名拼写错误无法在编译期发现
2. 类型断言散落各处（如 `t["node_code"].(string)`），容易 panic
3. 前端接口契约无法通过 Go 类型系统自动校验

**改进建议**：
- 定义专门的 DTO 结构体（如 `PendingTaskDTO`、`ProcessHistoryDTO`）
- 使用 GORM 的 `Scan(&dto)` 直接映射到结构体
- 这也有助于自动生成 API 文档（如 Swagger）

---

### 问题 9（重要·架构）：业务模型缺少独立的"样品"实体

**代码位置**：`backend/internal/model/business_models.go` 整体

**问题描述**：作为 LIMS 系统，"样品"（Sample）是最核心的实体之一，但当前数据模型中没有独立的 `Sample` 表。样品信息散落在：
- `TaskOrder.SampleType`（字符串字段）
- `SampleReceiving.SampleCodes`（JSONB 字段）
- `FieldSamplingRecord.SamplePhotos`（JSONB 字段）

这导致：
1. 无法对单个样品进行全生命周期追踪
2. 无法管理样品的流转、留样、销毁等状态
3. 样品编码无法系统化管理（CNAS 要求）

**改进建议**：
- 新增 `Sample` 模型：id, sample_code, sample_type, status, task_order_id, received_at, storage_location, retention_period 等
- 新增 `SampleChain` (样品流转链) 模型，记录每次交接
- `SampleReceiving.SampleCodes` 改为引用 Sample 表的外键关系

---

### 问题 10（重要·架构）：`preMigrateCleanup` 每次启动都执行 DELETE 操作

**代码位置**：`backend/internal/router/router.go` 第 128-175 行 `preMigrateCleanup` 函数

**问题描述**：每次应用启动时，都会执行 16 条 DELETE 语句清理孤儿数据。这存在严重风险：
1. 在生产环境如果外键约束暂时不一致（如迁移过程中），会误删数据
2. 没有任何确认机制或 dry-run 模式
3. 虽然只在非 production 环境运行 AutoMigrate，但 `preMigrateCleanup` 的调用在 `autoMigrate` 内部，如果环境变量配置错误可能导致意外删除

**改进建议**：
- 改用正确的数据库迁移工具（如 golang-migrate/goose），而非 AutoMigrate + 手动清理
- 清理操作应作为独立的运维脚本，不应嵌入启动流程
- 添加事务保护和回滚机制

---

### 问题 11（重要·前端）：前端路由守卫权限校验过于简单

**代码位置**：`frontend/src/router/index.ts` 第 218-223 行

```typescript
const requiredPerm = to.meta.permission as string | undefined
if (requiredPerm && !userStore.hasPermission(requiredPerm)) {
    next('/dashboard')
    return
}
```

**问题描述**：
1. 权限不足时静默跳转到 dashboard，没有提示用户
2. 只支持单权限校验，不支持 AND/OR 组合
3. 权限信息来自本地 store（JWT 解码），与问题 3 一样存在过期问题
4. `univer-test` 测试路由暴露在生产环境中

**改进建议**：
- 权限不足时跳转到 403 页面并给出提示
- 移除 `univer-test` 测试路由或在生产构建中过滤
- 考虑动态路由：根据后端返回的权限列表动态生成菜单和路由

---

### 问题 12（中等·架构）：工作流引擎不支持任务超时/SLA 管理

**代码位置**：`backend/internal/workflow/engine.go` 整体

**问题描述**：当前工作流引擎没有 SLA/超时机制。在真实 LIMS 场景中：
- 样品有保质期，检测必须在规定时间内完成
- 客户对报告交付有时间要求
- CNAS 审核会检查任务时效性

**改进建议**：
- 在 `NodeDefinition` 或 `ProcessTask` 中添加 `sla_deadline` 字段
- 实现定时任务（如 cron job），检查超时任务并触发告警
- 前端工作台展示即将超时的任务，按紧急程度排序
- 可集成消息通知（邮件/企业微信/钉钉）

---

### 问题 13（中等·架构）：不支持任务委派/代理审批

**代码位置**：`backend/internal/workflow/engine.go` 的 `AssignTask`/`AssignTaskWithTx` 函数

**问题描述**：当前只支持将任务分配给指定用户，但不支持：
1. 请假时的代理审批人设置
2. 任务转派（A 转给 B 处理）
3. 会签（多人同时审批）

对于 100 人规模的团队，人员请假/出差是常态，缺少委派机制会导致流程卡死。

**改进建议**：
- 新增 `Delegation` 模型：delegator_id, delegate_id, start_date, end_date, scope
- 在 `ApproveTaskWithTx` 中检查是否存在有效委派
- 审批历史中记录"代审"信息

---

### 问题 14（中等·安全）：JSONB 字段在 Go 模型中使用 string 类型

**代码位置**：`backend/internal/model/business_models.go` 多处，例如：

```go
TestItems    string `gorm:"type:jsonb" json:"test_items"`
QCDetails    string `gorm:"type:jsonb" json:"qc_details"`
SampleCodes  string `gorm:"type:jsonb" json:"sample_codes"`
```

**问题描述**：JSONB 字段使用 `string` 类型存储，GORM 会将其作为普通字符串处理而非 JSON 对象。这导致：
1. 无法利用 GORM 的 JSON 查询能力
2. 写入时无法校验 JSON 格式合法性
3. 序列化/反序列化需要手动 `json.Marshal`/`Unmarshal`

**改进建议**：
- 使用 `datatypes.JSON`（gorm.io/datatypes）或自定义类型实现 `Scanner`/`Valuer` 接口
- 或直接定义 Go 结构体来映射 JSONB 内容

---

### 问题 15（中等·架构）：缺少 Redis 集成和缓存层

**代码位置**：架构图中标注 Redis "待集成"，代码中未使用 Redis

**问题描述**：对于 100 人规模的 LIMS 系统：
1. Token Blacklist 需要 Redis 做持久化（问题 2）
2. 频繁的待办任务查询（每次刷新页面都会调用）无缓存
3. `deptNameMap()` 函数每次调用都查数据库（engine.go 约第 410 行）
4. 无分布式会话支持

**改进建议**：
- 尽快集成 Redis（已在架构图中预留）
- 缓存部门映射、权限列表等不常变动的数据
- 待办任务列表可设置 30 秒的缓存窗口

---

### 问题 16（中等·功能缺失）：缺少客户管理（CRM）模块

**问题描述**：当前系统只有 `TaskOrder.CustomerName` 字符串字段来记录客户信息。一个成熟的 LIMS 系统需要：
- 客户档案管理（联系方式、资质、历史合作）
- 客户自助下单/进度查询
- 客户合同管理
- 客户关系维护

**改进建议**：新增 `Customer` 模型，与 TaskOrder 建立关联。

---

### 问题 17（中等·功能缺失）：缺少报告模板和电子签章集成

**代码位置**：`backend/internal/model/business_models.go` 的 `ReportPrepare` 和 `ReportSign` 模型

**问题描述**：
- `ReportSign.SignStamp` 只是一个 500 字符的字符串路径，不是真正的电子签章
- 报告编制没有模板系统，`ReportContent` 是 JSONB 自由格式
- 无 PDF 自动生成能力

在 CNAS 认证审查中，报告签章是重点检查项。

**改进建议**：
- 集成电子签章服务（如 e签宝、法大大 API）
- 实现报告模板引擎（基于 Word/PDF 模板填充数据）
- 报告 PDF 自动生成与归档

---

### 问题 18（中等·工程化）：缺少单元测试（仅 workflow engine 有测试）

**代码位置**：从目录结构可见，仅 `backend/internal/workflow/engine_test.go` 有测试文件

**问题描述**：
- 16 个业务 Handler 无测试
- 权限中间件无测试
- JWT 工具无测试
- 前端无测试

对于要通过 CNAS 认证的 LIMS 系统，测试覆盖率是审查的重要指标。

**改进建议**：
- 优先为核心工作流引擎补充边界测试（驳回后再审批、并发审批等）
- 为认证/鉴权中间件添加测试
- 前端至少对关键业务流程添加 E2E 测试

---

### 问题 19（建议·可维护性）：前端 Pinia Store 使用不充分

**代码位置**：`frontend/src/stores/` 目录（仅 user store 可见）

**问题描述**：
- 待办任务数据、流程进度数据等未通过 Pinia 统一管理
- 多处组件可能重复请求相同数据
- 缺少全局 loading 状态管理

**改进建议**：
- 新增 `useWorkflowStore` 管理待办任务和流程状态
- 新增 `useNotificationStore` 管理消息通知
- 统一数据获取和缓存策略

---

### 问题 20（建议·可维护性）：CORS 开发模式过于宽松

**代码位置**：`backend/internal/router/router.go` 第 47-51 行

```go
if cfg.Env == "development" {
    return strings.HasPrefix(origin, "http://localhost:") ||
        strings.HasPrefix(origin, "http://127.0.0.1:") ||
        strings.HasPrefix(origin, "http://0.0.0.0:")
}
```

**问题描述**：开发环境下接受任何 localhost 端口的请求。虽然风险较低，但如果开发环境同时运行了恶意页面，可能被利用。

**改进建议**：开发环境也应限定具体端口（如只允许 3000、5173）。

---
