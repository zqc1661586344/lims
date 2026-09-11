# LIMS 实验室管理系统深度代码评审报告

## 核心问题清单（按严重性排序）

### 问题 1（严重·安全性）：业务路由完全缺失 RBAC 权限校验

**涉及文件**：`backend/internal/router/routes/business_routes.go` → `RegisterBusinessRoutes`

所有 16 个业务节点的路由仅挂载了 `AuthMiddleware`（JWT 验证），完全未挂载 `PermissionMiddleware`。任何已登录用户均可访问任意节点的 CRUD 及 approve/reject 接口。例如"业务室"人员可调用合同评审审批接口、"现场室"人员可执行报告签发，完全违背 LIMS 部门职责隔离要求。

虽然 `middleware/permission.go` 已实现 `PermissionMiddleware` 且在 `system_routes.go` 中正常使用，但 business routes 完全未接入。

**流程评分：2/10**

---

### 问题 2（严重·安全性）：审批操作无用户部门归属校验

**涉及文件**：`backend/internal/workflow/engine.go` → `ApproveTaskWithTx` / `RejectTaskWithTx`

审批与驳回逻辑仅校验 task 状态是否为 `pending`，未校验当前操作用户的 `dept_id` 是否匹配该节点负责部门。任何已登录用户只需知道 `task_id` 即可越权审批任意节点。

正确做法应在审批前校验：
```
user.dept_id == dept WHERE code = nodeDef.DeptCode
```

**流程评分：2/10**

---

### 问题 3（严重·安全性）：CORS 同时开启 AllowAllOrigins 和 AllowCredentials

**涉及文件**：`backend/internal/router/router.go` → `Setup`

```go
cors.New(cors.Config{
    AllowAllOrigins:  true,
    AllowCredentials: true,  // 危险组合
    ...
})
```

`AllowAllOrigins: true` + `AllowCredentials: true` 是危险组合。CORS 规范禁止此组合，但部分浏览器实现存在差异，可能导致 CSRF 攻击。生产环境应明确指定域名白名单。

**流程评分：3/10**

---

### 问题 4（严重·安全性）：JWT Secret 硬编码且无生产环境强制替换

**涉及文件**：`backend/config/config.yaml`

```yaml
jwt:
  secret: lims-jwt-secret-change-in-production
  expire_hour: 24
```

没有环境变量覆盖的强制校验，也没有启动时检测是否仍使用默认 secret 的告警。部署时极易遗漏，且 24 小时 token 有效期偏长。

**流程评分：3/10**

---

### 问题 5（严重·正确性）：乐观锁 Version 字段定义了但从未使用

**涉及文件**：
- `backend/internal/model/workflow.go` → `ProcessTask.Version`
- `backend/internal/workflow/engine.go` → `ApproveTaskWithTx`
- `backend/internal/workflow/errors.go` → `ErrVersionConflict`

`ProcessTask` 模型定义了 `Version int` 字段，`errors.go` 也定义了 `ErrVersionConflict`，但 engine 中所有 UPDATE 均为裸 SQL：
```sql
UPDATE process_tasks SET status=?, comment=?, updated_at=NOW() WHERE id=?
```
完全没有 `WHERE version=?` 检查和 `SET version=version+1`。虽然使用了 `FOR UPDATE` 行锁保护，但 Version 字段的存在会误导维护者以为有乐观锁保护。

**流程评分：4/10**

---

### 问题 6（严重·安全性）：所有业务节点 DELETE 接口无流程状态检查

**涉及文件**：`backend/internal/router/routes/business_routes.go`（16 条 DELETE 路由）及各 handler 的 `Delete` 方法

除任务委托（TaskOrder）的 `Delete` 检查了 `status == 0` 外，其余 15 个节点的 Delete 方法均不检查该记录是否属于运行中的流程实例。进行中的流程关联业务数据可被任意删除，导致流程与数据不一致。

**流程评分：3/10**

---

### 问题 7（重要·正确性）：已审批节点的业务数据仍可通过 PUT 修改

**涉及文件**：所有业务节点 Handler 的 `Update` 方法

节点审批通过、流程推进后，当前节点的业务数据（如评审意见、检测数据）仍可通过 PUT 接口修改。这违反了 LIMS 的数据不可篡改要求。应在 Update 中检查对应流程实例的 `current_node` 是否已离开当前节点。

**流程评分：4/10**

---

### 问题 8（重要·正确性）：TaskOrder 状态流转逻辑缺失

**涉及文件**：
- `backend/internal/handler/business_handler.go` → `Submit`
- `backend/internal/model/business_models.go` → `TaskOrder.Status`

Status 注释定义为 `0=草稿, 1=已提交, 2=流程中, 3=已完成`，但代码只在 Submit 时设为 1，后续流程流转中从未更新。前端列表页的 4 种状态标签中，"流程中"和"已完成"永远不会显示。

**流程评分：5/10**

---

### 问题 9（重要·安全性）：生产代码残留大量 DEBUG 打印

**涉及文件**：
- `backend/internal/handler/business_handler.go:83`
- `backend/internal/service/business_service.go:361-396`（10 处）

```go
fmt.Printf("[DEBUG CreateTaskOrderRequest] TestItems = %q (len=%d)\n", ...)
fmt.Printf("[DEBUG parseTestItemIDs] raw = %s\n", raw)
```

这些 debug 打印会在生产环境将用户提交的敏感数据输出到 stdout，可能被日志收集系统采集。

**流程评分：5/10**

---

### 问题 10（重要·安全性）：Token Blacklist 仅内存存储

**涉及文件**：`backend/internal/utils/token_blacklist.go`

Token 黑名单基于 `sync.RWMutex + map` 实现，进程重启后全部失效，多实例部署时 logout 不同步。docker-compose 中未包含 Redis 服务。

**流程评分：5/10**

---

### 问题 11（重要·安全性）：登录接口无频率限制

**涉及文件**：`backend/internal/handler/auth_handler.go` → `Login`

登录接口没有任何 rate limiting，也没有连续失败锁定机制，存在暴力破解风险。建议引入 `gin-rate-limit` 中间件或 IP 级限流。

**流程评分：5/10**

---

### 问题 12（重要·架构）：16 个 Handler 存在约 80% 重复代码

**涉及文件**：`backend/internal/handler/` 下 16 个 `*_handler.go`

每个 Handler 的 List/Get/Create/Update/Delete 方法结构完全一致，仅 model type 不同。应使用 Go 1.18+ 泛型 `GenericHandler[T any]` 或表驱动模式统一抽象。

**流程评分：5/10**

---

### 问题 13（重要·正确性）：审计日志 afterUpdate 旧数据获取逻辑缺陷

**涉及文件**：`backend/internal/middleware/audit.go` → `afterUpdate` / `cloneModel`

`cloneModel` 函数仅返回同一指针（`return m`），未实现真正的深拷贝。当用此指针查询"旧数据"时，GORM 可能已修改内存值，导致 `old_data` 与 `new_data` 几乎相同，审计对比失去意义。

**流程评分：5/10**

---

### 问题 14（重要·可用性）：所有列表接口无分页

**涉及文件**：所有 Handler 的 `List` 方法

所有 List 均使用 `query.Find(&items)` 不带分页。100 人企业使用半年后数据量可达数万条，一次性全量返回将导致性能问题和前端卡顿。

**流程评分：5/10**

---

### 问题 15（重要·正确性）：ProcessTask 的 AssigneeUserID 从未设置

**涉及文件**：`backend/internal/workflow/engine.go` → `createTaskWithTx`

创建任务时仅设置 `AssigneeDeptID`，`AssigneeUserID` 始终为 nil。数据录入、报告编制等需要指定具体操作人的节点，缺少个人级别的任务分配能力。

**流程评分：5/10**

---

### 问题 16（中等·流程合理性）：驳回仅支持退回上一节点，不支持跨节点

**涉及文件**：
- `backend/internal/workflow/definition.go` → 各节点 `RejectTarget`
- `backend/internal/workflow/engine.go` → `RejectTaskWithTx`
- `frontend/src/components/ApprovalDialog.vue` → 驳回目标选择 UI

前端已有驳回目标选择 UI，但后端 `RejectTaskWithTx` 固定使用节点定义中的 `RejectTarget`，忽略前端传入的选择。实际 LIMS 场景中，报告审核发现数据录入有误时需直接驳回到数据录入节点。

**流程评分：6/10**

---

### 问题 17（中等·流程合理性）：流程定义硬编码，无法配置化

**涉及文件**：`backend/internal/workflow/definition.go`

16 个节点写死在代码中，无法通过管理界面调整。真实 LIMS 需支持不同检测项目有不同流程（如某些项目不需要现场采样），需要可配置的流程模板。

**流程评分：6/10**

---

### 问题 18（中等·缺失功能）：核心业务模块缺失

| 缺失模块 | 影响 |
|----------|------|
| 客户管理 | 客户信息仅为字符串字段，无独立实体、联系人、历史委托 |
| 样品全生命周期 | 缺少编码规则、状态流转（收样→分发→检测→留样/销毁）|
| 消息通知 | 流程流转后无邮件/短信/站内信提醒 |
| Dashboard | 仅显示欢迎语，无待办统计、超期任务、处理时长等 |
| 报告自动生成 | Dockerfile 安装了 wkhtmltopdf 但无代码调用 |

**流程评分：6/10**

---

### 问题 19（中等·安全性）：前端权限信息可被客户端篡改

**涉及文件**：`frontend/src/stores/user.ts`

`is_admin` 和 `permissions` 存储在 `localStorage` 中，用户可通过开发者工具修改。虽然后端 `PermissionMiddleware` 可做二次校验，但当前 business routes 未使用，导致前端成为唯一防线。

**流程评分：6/10**

---

### 问题 20（中等·质量）：前后端完全缺失单元测试

整个代码库无任何 `_test.go` 文件或前端测试文件。对 LIMS 这种需 CNAS 认证的系统，缺少测试覆盖是严重的质量与合规风险。

**流程评分：6/10**

---

### 问题 21（中等·运维）：Recovery 中间件引用但未自定义实现

**涉及文件**：`backend/internal/router/router.go`

使用 Gin 默认的 `gin.Recovery()` 而非自定义 recovery，未对 panic 做结构化日志记录或告警。README 中提到的 `recovery.go` 文件不存在。

**流程评分：7/10**

---
