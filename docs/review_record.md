# LIMS 实验室管理系统代码深度评审报告


## 一、问题列表（按严重性排序）

### 严重级别（Critical）

#### 问题 1（严重·安全性）：审批接口无部门校验，任意用户可操作任意节点

- **文件**：`backend/internal/router/routes/business_routes.go` 全文
- **函数**：`RegisterBusinessRoutes`
- **描述**：16 个业务节点路由仅挂载 `AuthMiddleware` + `PermissionMiddleware("business:xxx")`。`PermissionMiddleware` 只检查用户是否拥有权限码，不校验操作者是否属于该节点的负责部门。例如：业务室员工持有 `business:data-entry` 权限即可调用数据录入的 approve 接口，直接越权操作实验室节点。
- **修复建议**：新增 `DeptScopeMiddleware`，在每个节点路由组中注入部门白名单校验（如 `data-entry` 仅允许 `dept_lab`）。
- **评分**：2/10

#### 问题 2（严重·安全性）：引擎部门校验存在 bypass 漏洞

- **文件**：`backend/internal/workflow/engine.go`
- **函数**：`ApproveTaskWithTx`（约第 95 行）
- **代码**：`if task.AssigneeDeptID != 0 && task.AssigneeDeptID != userDeptID { return ErrDeptNotMatch }`
- **描述**：当 `AssigneeDeptID=0` 时（如 depts 表中 deptCode 找不到对应记录），部门校验被完全跳过，任何部门的用户均可审批该任务。`createTaskWithTx` 通过 `SELECT id FROM depts WHERE code = ?` 解析部门 ID，若查询失败则 `resolvedDeptID` 为 0，直接导致后续校验失效。
- **修复建议**：改为强制校验 `if task.AssigneeDeptID != userDeptID`，不允许 0 值 bypass；`createTaskWithTx` 中增加 resolvedDeptID 为 0 时的 error 返回。
- **评分**：2/10

#### 问题 3（严重·数据完整性）：已审批数据可被 DELETE 接口物理删除

- **文件**：`backend/internal/handler/data_entry_handler.go`
- **函数**：`Delete`（约第 115 行）
- **描述**：`Delete` 函数仅调用 `CheckInstanceRunning`（检查流程是否在运行），未调用 `CheckNodeNotAdvanced`（检查节点是否已审批通过）。这意味着节点 8（数据录入）审批通过后，实验原始记录仍可被 DELETE 接口物理删除。**直接违反 CNAS/ISO 17025 "原始数据不可手动删除或覆盖" 的合规底线**。
- **修复建议**：所有节点的 DELETE 接口必须增加节点状态校验，已审批的记录禁止物理删除，改用"作废"软删除 + 审计记录。
- **评分**：1/10

#### 问题 4（严重·并发安全）：乐观锁机制未实现，version 字段形同虚设

- **文件**：`backend/internal/workflow/engine.go`
- **函数**：`getTaskForUpdate`、`ApproveTaskWithTx`
- **描述**：`taskRow` 结构体包含 `Version` 字段，`getTaskForUpdate` 也查询了 version，但 `ApproveTaskWithTx` 中的 `UPDATE` 语句从未递增 version，也没有 `WHERE version=?` 条件。`ErrOptimisticLock` 被定义但从未返回。虽然 `FOR UPDATE` 行锁在大多数场景下能防止双重审批，但乐观锁机制属于未完成的半成品，在高并发或长事务场景下存在隐患。
- **修复建议**：UPDATE 语句改为 `UPDATE process_tasks SET status=?, version=version+1 WHERE id=? AND version=?`，影响行数为 0 时返回 `ErrOptimisticLock`。
- **评分**：3/10

#### 问题 5（严重·架构）：引擎层硬编码业务表操作，违反单一职责

- **文件**：`backend/internal/workflow/engine.go`
- **函数**：`ApproveTaskWithTx`（约第 125-140 行）
- **代码**：`tx.Exec("UPDATE task_orders SET status=2 WHERE id=?", instance.BusinessID)`
- **描述**：工作流引擎作为通用状态机，直接在核心逻辑中硬编码 `task_orders` 表名和 `status=2/3` 的业务语义。引擎应只负责状态流转，不应感知具体业务表结构。如需支持其他业务类型（如设备校准流程），必须修改引擎源码。
- **修复建议**：采用事件回调或接口抽象（如 `BusinessStatusUpdater` 接口），引擎只负责状态流转，业务表状态同步由上层 Service 处理。
- **评分**：3/10

#### 问题 6（严重·安全）：Admin 用户跳过全部权限校验

- **文件**：`backend/internal/middleware/permission.go`
- **函数**：`PermissionMiddleware`（第 17 行）
- **代码**：`if IsAdmin(c) { c.Next(); return }`
- **描述**：Admin 用户绕过所有权限检查，可以操作任何节点的审批、删除任何数据。在 CNAS 合规环境中，即便是系统管理员也不应有权审批自己创建的检测任务（职责分离原则）。
- **修复建议**：Admin 仅在系统管理类接口（`/api/system/*`）跳过权限，业务类接口（`/api/business/*`）必须走完整校验链。
- **评分**：2/10

---

### 重要级别（High）

#### 问题 7（重要·安全）：Token 黑名单基于内存，重启即失效

- **文件**：`backend/internal/middleware/auth.go`（约第 40 行）
- **描述**：`utils.GetTokenBlacklist().IsRevoked()` 使用进程内存存储，服务重启后所有已注销 token 重新有效。攻击者可在服务重启后使用已注销的旧 token 访问系统。
- **修复建议**：引入 Redis 存储 Token 黑名单，TTL 设为 Token 剩余有效期。

#### 问题 8（重要·安全）：前端 Token 存储于 localStorage，XSS 可窃取

- **文件**：`frontend/src/api/request.ts`（第 10 行）
- **描述**：`localStorage.getItem('token')` 存储 JWT，XSS 攻击可通过 `document.cookie` 或直接读取 localStorage 窃取 Token。
- **修复建议**：使用 `HttpOnly + Secure + SameSite` Cookie 存储；或至少启用 CSP 头防御 XSS。

#### 问题 9（重要·安全）：JWT 权限列表签发后不可变，权限回收延迟

- **文件**：`backend/internal/middleware/auth.go`
- **描述**：`permissions` 嵌入 JWT claims，用户权限修改后需重新登录才生效。员工离职/调岗后权限无法即时回收，存在安全隐患。
- **修复建议**：JWT 仅存储 `user_id`，每次请求从 Redis 缓存读取实时权限列表。

#### 问题 10（重要·正确性）：Approve handler 中 TaskID 命名严重误导

- **文件**：`backend/internal/handler/data_entry_handler.go`（约第 135 行）
- **函数**：`Approve`
- **代码**：`TaskID uint json:"task_id" binding:"required"`
- **描述**：`req.TaskID` 实际传入的是 `task_order_id`（业务委托单 ID），而非 workflow task ID。前端开发者极易传错参数，导致审批到错误的任务。
- **修复建议**：改名为 `TaskOrderID`，API 文档中明确标注。

#### 问题 11（重要·正确性）：Handler → Service → Engine 跨层事务缺失

- **文件**：handler → service → engine 调用链
- **描述**：Handler 中 `count==0` 检查在事务外执行，Service 调用 Engine 时开启新事务。存在 TOCTOU（Time-of-Check-to-Time-of-Use）竞态：检查通过后、事务提交前，另一个请求可能删除了最后一条数据记录。
- **修复建议**：在 Handler 层开启事务，将 `*gorm.DB` 通过 context 传递到 Engine，确保全链路事务一致性。

#### 问题 12（重要·架构）：queryTasks 返回无类型 map，丧失编译期安全

- **文件**：`backend/internal/workflow/engine.go`（约第 350 行）
- **函数**：`queryTasks`
- **描述**：返回 `[]map[string]interface{}`，调用方需要大量类型断言（如 `t["node_code"].(string)`），极易 panic。
- **修复建议**：定义 `PendingTask`、`HistoryEntry` 等强类型结构体，利用 Go 的类型系统。

#### 问题 13（重要·架构）：16 节点路由注册代码严重重复

- **文件**：`backend/internal/router/routes/business_routes.go`（226 行）
- **描述**：16 个节点的路由注册代码结构完全相同（GET/POST/GET/:id/PUT/:id/DELETE/:id/POST/:id/approve/POST/:id/reject），仅资源名和 handler 不同。修改一处需同步 16 处。
- **修复建议**：封装 `registerNodeRoutes(name string, handler NodeHandler)` 函数，通过循环注册。

#### 问题 14（重要·功能）：前端无 Token 刷新机制，用户体验差

- **文件**：`frontend/src/api/request.ts`（约第 25 行）
- **描述**：401 响应直接 `router.push('/login')`，没有尝试调用 `/api/auth/refresh`。后端已有 refresh 接口但前端完全未使用。
- **修复建议**：实现 Axios 拦截器中的 Token 无感刷新逻辑（401 时自动 refresh → 重试原请求）。

#### 问题 15（重要·功能）：前端无路由守卫，未登录可访问受保护页面

- **文件**：`frontend/src/router/index.ts`
- **描述**：无 `beforeEach` 路由守卫，未登录用户可通过直接输入 URL 访问受保护页面（虽然 API 会 401，但页面仍会渲染空白内容）。
- **修复建议**：添加全局路由守卫，检查 Token 存在性，不存在则重定向到 `/login`。

---

### 中等级别（Medium）

#### 问题 16（中等·工程）：零单元测试覆盖

- **描述**：Makefile 中有 `go test ./...` 但项目中无 `_test.go` 文件。工作流引擎作为核心模块，零测试覆盖率风险极高。
- **修复建议**：优先为 `engine.go` 的 `StartInstance`/`ApproveTask`/`RejectTask` 编写表驱动测试。

#### 问题 17（中等·工程）：无 API 文档

- **描述**：无 Swagger/OpenAPI 注解，16 个节点的接口参数全靠阅读源码理解。前后端协作效率低。
- **修复建议**：引入 `swaggo/swag` 自动生成 Swagger 文档。

#### 问题 18（中等·安全）：ratelimit.go 存在但未挂载使用

- **文件**：`backend/internal/middleware/ratelimit.go`
- **描述**：限流中间件已编写但未在任何路由中挂载。登录接口无限流，存在暴力破解风险。
- **修复建议**：在 `/api/auth/login` 路由挂载限流中间件（如 5次/分钟）。

#### 问题 19（中等·安全）：全局接口限流缺失

- **描述**：所有 API 无请求频率限制，恶意用户可高频调用审批/创建接口。
- **修复建议**：全局限流（如 100 req/s）+ 关键接口分级限流。

#### 问题 20（中等·架构）：前端 Pinia Store 空目录

- **文件**：`frontend/src/stores/`
- **描述**：状态管理目录为空，用户信息、权限列表、待办数量等全局状态应存于 Store 中。
- **修复建议**：实现 `userStore`（用户信息）、`permissionStore`（权限）、`workflowStore`（待办）。

#### 问题 21（中等·运维）：日志缺乏业务语义

- **文件**：`backend/internal/middleware/logger.go`
- **描述**：请求日志仅记录 HTTP 方法/路径/耗时，不记录业务操作语义（如"张三审批了合同评审#123"）。排查问题时效率低。
- **修复建议**：在 Handler 层增加业务日志中间件，记录关键业务操作。

#### 问题 22（中等·部署）：Docker Compose 缺乏生产级配置

- **文件**：`deploy/docker-compose.yml`
- **描述**：缺少容器健康检查、自动重启策略（`restart: always`）、资源限制（CPU/内存）、日志轮转等生产关键配置。
- **修复建议**：添加 `healthcheck`、`restart: always`、`deploy.resources` 配置。

---

