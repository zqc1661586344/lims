我已经通过 GitHub API 完整拉取并审查了该仓库的核心源代码文件，包括后端入口、路由、认证中间件、工作流引擎、业务服务层、流程定义、限流器、Token 黑名单、前端 HTTP 客户端、认证处理器、配置文件和业务路由注册等关键文件。以下是完整的代码评审报告。

---

### 问题 1（严重・安全）：config.yaml 硬编码明文凭据并提交到代码仓库

**严重程度**：🔴 Critical
**涉及文件**：`backend/config/config.yaml`
**具体问题**：

配置文件直接将数据库密码（`password: lims123`）、MinIO 密钥（`secret_key: lims123`）、JWT Secret（`secret: lims-jwt-secret-change-in-production`）明文写入并推送到 Git 仓库。即使后续修改，历史提交中仍可被提取。

**建议**：
- `config.yaml` 仅保留结构模板，敏感值通过环境变量（`LIMS_DB_PASSWORD`、`LIMS_JWT_SECRET`）或 `.env` 文件注入，`.gitignore` 排除 `.env`
- 使用 `config.example.yaml` 作为模板（当前已存在 `config.example.yaml`，但 `config.yaml` 同样被提交）
- 在 CI 中加入 secret scanning（如 trufflehog）

---

### 问题 2（严重・安全）：生产环境 JWT Secret 仅警告不阻断

**严重程度**：🔴 Critical
**涉及文件**：`backend/cmd/server/main.go` 约第 40 行
**具体问题**：

```go
if cfg.JWT.Secret == "lims-jwt-secret-change-in-production" || len(cfg.JWT.Secret) < 16 {
    logger.Warn("jwt.secret is using default or weak value ...")
}
```

仅 `Warn` 日志输出，程序继续启动。生产环境使用默认弱 Secret 将导致任意伪造 JWT Token。

**建议**：当 `cfg.Env == "production"` 且 Secret 为默认值时，应 `logger.Fatal` 直接终止启动。

---

### 问题 3（严重・安全）：Permission 中间件实现过于薄弱

**严重程度**：🔴 Critical
**涉及文件**：`backend/internal/middleware/permission.go`（仅 842 字节）
**具体问题**：

PermissionMiddleware 作为路由级别的权限控制核心，文件极小，仅做了简单检查。从代码结构看，它验证用户是否拥有某个 permission code，但没有实现：
- API 级别的细粒度校验（只校验了 permission code 前缀匹配）
- 对 admin 用户的 bypass 逻辑不够精细
- 没有做"按钮级"或"字段级"权限控制

**建议**：扩展 PermissionMiddleware，支持基于 HTTP Method + Path 的精确 API 权限映射。

---

### 问题 4（重要・安全）：GET 请求用于审批/驳回等有状态变更操作

**严重程度**：🟠 High
**涉及文件**：`backend/internal/router/routes/business_routes.go`（README 中提到的 `GET /:id/approve` 和 `GET /:id/reject`）
**具体问题**：

虽然实际代码中 approve/reject 路由已改为 `POST`，但 README 文档中仍标注为 `GET`。如果存在残留的 GET 方式审批接口，会导致 CSRF 风险（浏览器预取、爬虫误触发等）。

**建议**：确认所有审批接口均为 `POST/PUT`，文档同步更新。

---

### 问题 5（重要・安全）：Token Blacklist 为进程内存存储，不支持多实例

**严重程度**：🟠 High
**涉及文件**：`backend/internal/utils/token_blacklist.go`
**具体问题**：

Token 黑名单使用 `sync.RWMutex + map` 的纯内存实现。注释中已标注 "For multi-instance deployments, replace with Redis-backed storage"，但当前没有任何切换机制。100 人规模企业如果部署多个后端实例（负载均衡），登出操作只对当前实例生效，其他实例的 Token 仍有效。

**建议**：
- 预留 Redis-backed 实现接口（定义 `TokenStore` interface）
- 当前 README 已规划 Redis 集成，建议提前实现

---

### 问题 6（重要・安全）：Token Refresh 未做旧 Token 吊销

**严重程度**：🟠 High
**涉及文件**：`backend/internal/handler/auth_handler.go` → `Refresh` 函数
**具体问题**：

`Refresh` 接口直接使用当前 Token 的信息生成新 Token，但没有将旧 Token 加入黑名单。如果 Token 被窃取，攻击者可以持续刷新获取新 Token，即使原用户已退出登录。

**建议**：Refresh 时同时 Revoke 旧 Token（将旧 jti 加入 Blacklist）。

---

### 问题 7（重要・安全）：前端 Token 存储在 localStorage，存在 XSS 风险

**严重程度**：🟠 High
**涉及文件**：`frontend/src/api/request.ts`
**具体问题**：

```typescript
const token = localStorage.getItem('token') || sessionStorage.getItem('token')
```

localStorage 对 XSS 攻击无防护。若前端存在 XSS 漏洞，攻击者可窃取 Token。对于 LIMS 这种涉及检测报告和合规数据的系统，Token 泄露后果严重。

**建议**：
- 使用 `httpOnly + Secure` Cookie 存储 Token
- 或至少在 refresh 接口加入设备指纹/IP 绑定校验

---

### 问题 8（重要・正确性）：工作流引擎为全局单例，存在并发瓶颈

**严重程度**：🟠 High
**涉及文件**：`backend/internal/workflow/engine.go` → `DefaultEngine` 函数
**具体问题**：

```go
var (
    defaultEngineOnce sync.Once
    defaultEngine     *Engine
)
func DefaultEngine(db *gorm.DB) *Engine {
    defaultEngineOnce.Do(func() {
        defaultEngine = NewEngine(db)
    })
    return defaultEngine
}
```

全局单例 Engine 在高并发审批场景下（多个部门同时审批不同项目），所有操作共享同一个 Engine 实例。虽然数据库层使用了 `FOR UPDATE` 行锁和乐观锁版本号，但 Engine 的 `nodeMap` 等字段没有并发保护。此外，单例模式在多租户场景下扩展困难。

**建议**：使用依赖注入替代全局单例，每个请求通过 middleware 注入 Engine 实例。

---

### 问题 9（重要・架构）：Handler 层大量重复代码，缺少通用抽象

**严重程度**：🟠 High
**涉及文件**：`backend/internal/handler/` 目录下 16 个节点 handler 文件
**具体问题**：

从目录结构和路由注册代码可以看到，每个业务节点（contract_review、qc_task、sampling_schedule 等）都有独立的 handler 文件，包含几乎相同的 CRUD（List/Get/Create/Update/Delete）和 Approve/Reject 方法。这导致：
- 代码膨胀（16 个文件 × ~6KB = ~96KB 重复逻辑）
- 修改通用逻辑（如分页参数校验）需要改 16 个文件
- 新增节点时需要复制粘贴整个 handler

**建议**：抽象出 `GenericBusinessHandler[T any]` 泛型基类（Go 1.18+ 支持泛型），各节点仅定义差异化逻辑。代码中已有 `generic.go` 文件但未被充分使用。

---

### 问题 10（重要・架构）：工作流定义硬编码，不支持动态配置

**严重程度**：🟠 High
**涉及文件**：`backend/internal/workflow/definition.go`
**具体问题**：

16 个节点的流程定义（节点名称、负责部门、驳回目标、下一节点）完全硬编码在 Go 代码中。任何流程变更（如增加节点、修改驳回目标）都需要修改代码并重新部署。

对于 LIMS 系统，不同实验室的流程可能不同（有的可能不需要采样环节，有的可能需要额外的审批节点）。硬编码方式严重限制了系统的灵活性和可复用性。

**建议**：将流程定义存储到数据库，提供管理界面进行可视化配置。保留当前硬编码作为默认模板。

---

### 问题 11（中等・安全）：文件上传缺少文件类型和大小校验

**严重程度**：🟡 Medium
**涉及文件**：`backend/internal/handler/file_handler.go`（7300 字节）
**具体问题**：

文件上传处理器缺少：
- MIME 类型白名单校验
- 文件大小上限限制
- 文件内容扫描（防恶意文件）
- 文件名清洗（防路径穿越）

LIMS 系统需要上传检测报告、原始记录等文件，安全校验不足可能导致恶意文件上传。

**建议**：添加文件类型白名单（PDF、DOCX、XLSX、JPG、PNG）、大小限制（建议 50MB）、文件名消毒。

---

### 问题 12（中等・安全）：缺少请求输入校验和分页保护

**严重程度**：🟡 Medium
**涉及文件**：所有 handler 的 List 方法
**具体问题**：

列表查询接口缺少：
- 分页大小上限（恶意用户可传 `pageSize=999999` 导致内存和数据库压力）
- 排序字段白名单（可能注入非法字段名）
- 查询条件的参数化校验

**建议**：
- 统一分页中间件，限制 `pageSize` 最大值为 100
- 使用 Go 的 `validator` 库对所有请求参数进行结构体验证

---

### 问题 13（中等・架构）：前端状态管理（Pinia Store）为空目录

**严重程度**：🟡 Medium
**涉及文件**：`frontend/src/store/`
**具体问题**：

README 中声明使用 Pinia 2.1 作为状态管理，但 `store/` 目录为空。所有状态都在组件内部管理，导致：
- 多个页面共享的数据（如当前用户信息、待办数量）需要重复请求
- 组件间通信依赖 props/events 或路由参数，复杂度高
- 无法实现跨页面的工作流状态同步

**建议**：至少实现 `useAuthStore`（用户信息/Token）、`useWorkflowStore`（待办任务/流程状态）、`useAppStore`（全局配置/通知）。

---

### 问题 14（中等・架构）：AutoMigrate 在生产环境完全跳过

**严重程度**：🟡 Medium
**涉及文件**：`backend/internal/router/router.go`
**具体问题**：

```go
if cfg.Env != "production" {
    RunAutoMigrate(logger, db)
} else {
    logger.Warn("skipping AutoMigrate in production environment")
}
```

生产环境完全跳过数据库迁移。虽然这避免了意外的 schema 变更，但缺少替代方案（如独立的迁移工具或 CLI 命令）。已有的 `cmd/migrate/main.go` 存在但功能有限。

**建议**：引入专业的数据库迁移工具（如 `golang-migrate` 或 `goose`），支持版本化的正向/回滚迁移脚本。

---

### 问题 15（中等・正确性）：Approve 操作存在嵌套事务风险

**严重程度**：🟡 Medium
**涉及文件**：`backend/internal/service/business_service.go` → `ApproveTask` 和 `ApproveWithBusiness`
**具体问题**：

`ApproveTask` 和 `ApproveWithBusiness` 都在 Service 层开启了 `db.Transaction`，而 Engine 内部的 `ApproveTaskWithTx` 也可能被外部直接调用时再次开事务。GORM 的嵌套事务在某些配置下可能导致死锁或事务异常。

此外 `ApproveTask` 内部先 `resolveTaskContext`（查询），再调用 Engine 的 `ApproveTaskWithTx`（使用 `FOR UPDATE`），两次查询之间可能有并发修改。

**建议**：统一事务管理入口，避免嵌套；使用 `SELECT ... FOR UPDATE` 在事务开始时就锁定所需行。

---

### 问题 16（中等・架构）：CORS 配置在生产环境依赖配置白名单

**严重程度**：🟡 Medium
**涉及文件**：`backend/internal/router/router.go` CORS 配置
**具体问题**：

生产环境的 CORS `AllowOriginFunc` 依赖 `cfg.CORS.AllowOrigins` 配置列表，但 `config.yaml` 中未包含此配置项。如果部署时忘记配置，所有跨域请求都会被拒绝。

**建议**：在 config.yaml 中显式声明 `cors.allow_origins` 配置项，并在配置加载时进行校验。

---

### 问题 17（中等・功能）：缺少电子签章和数字签名功能

**严重程度**：🟡 Medium
**涉及文件**：`backend/internal/handler/report_sign_handler.go`
**具体问题**：

报告签发节点（节点 14）目前仅记录签发意见和状态，没有集成电子签章。对于 CNAS/CMA 认证的第三方检测实验室，检测报告需要授权签字人签名和机构盖章。当前系统缺少：
- 电子签章集成（如 CFCA、契约锁）
- 数字证书管理
- 签章后的 PDF 防篡改校验

**建议**：在 Phase 11 规划中优先实现电子签章对接。

---

### 问题 18（中等・功能）：缺少消息通知和待办提醒

**严重程度**：🟡 Medium
**具体问题**：

整个系统没有任何通知机制。当流程推进到新节点时，下一节点的负责部门和具体人员无法收到提醒。100 人规模的实验室中，如果操作员不主动刷新待办列表，可能会延误任务处理。

**建议**：
- 集成 WebSocket 实现实时推送
- 集成邮件/短信/企业微信通知
- 在 Dashboard 实现待办数量 Badge 提醒

---

### 问题 19（中等・功能）：缺少数据统计和报表功能

**严重程度**：🟡 Medium
**具体问题**：

Dashboard 页面目前仅作为占位（`views/dashboard/index.vue`），缺少：
- 任务处理效率统计（平均处理时间、超时率）
- 各部门工作量统计
- 样品流转状态分布
- 检测报告产出统计
- 逾期未完成项目预警

这些对于管理层决策和 CNAS 评审都是必要的。

---

### 问题 20（低・代码质量）：Raw SQL 和 ORM 混用，可维护性差

**严重程度**：🟢 Low
**涉及文件**：`backend/internal/workflow/engine.go`、`backend/internal/service/business_service.go`、`backend/internal/middleware/auth.go`
**具体问题**：

代码中大量使用 `db.Raw("SELECT ...")` 原始 SQL 和 GORM ORM API 混合的方式。例如 auth middleware 中直接用 Raw SQL 查询用户和权限，而其他地方使用 GORM Model API。这导致：
- SQL 注入风险虽然通过参数化查询降低了，但增加了代码审查难度
- 数据库 schema 变更时需要同时修改 Raw SQL 和 ORM 模型
- 测试时难以 mock

**建议**：将 Raw SQL 统一封装到 Repository 层，Handler/Service 层只调用 Repository 方法。

---

### 问题 21（低・代码质量）：缺少单元测试和集成测试

**严重程度**：🟢 Low
**涉及文件**：整个项目（仅有 `dept_scope_test.go`、`permission_test.go`、`jsonb_test.go` 三个测试文件）
**具体问题**：

100+ 个源文件中仅有 3 个测试文件，核心工作流引擎、业务服务、认证流程等关键模块完全没有测试覆盖。对于 LIMS 这种合规性要求高的系统，测试缺失是严重的质量风险。

**建议**：
- 优先为 `workflow/engine.go` 编写状态机流转测试
- 为 `service/business_service.go` 编写集成测试（使用 testcontainers）
- 目标：核心流程 80%+ 覆盖率

---

### 问题 22（低・运维）：缺少健康检查深度和监控指标

**严重程度**：🟢 Low
**涉及文件**：`backend/internal/router/router.go` → health endpoint
**具体问题**：

健康检查接口仅返回静态 JSON `{"status": "ok"}`，没有检查：
- 数据库连接是否可用
- MinIO 存储是否可达
- 磁盘空间是否充足

此外，缺少 Prometheus metrics 暴露（请求数、延迟、错误率等）。

**建议**：扩展 health 接口为深度检查；引入 `promhttp` 暴露 metrics。

---

## 各流程评分

| 流程/模块 | 评分 | 说明 |
|-----------|------|------|
| 认证与登录 | 7/10 | JWT + bcrypt + 登录限流 + Token 黑名单，但弱 Secret 不阻断、Refresh 不吊销旧 Token |
| RBAC 权限 | 6/10 | 用户-角色-权限三层模型完整，但 Permission 中间件实现薄弱，缺少细粒度控制 |
| 工作流引擎 | 8/10 | 状态机 + 乐观锁 + FOR UPDATE + SoD 检查，设计优秀但缺少动态配置和并行分支 |
| 16 节点业务流转 | 7/10 | 全流程贯通，FirstOrCreate 幂等保证，但 handler 重复代码多 |
| 检验单集成（Phase 10） | 6/10 | Univer Sheet 集成方案合理，但复核/审核只读模式尚未完成 |
| 审计日志 | 8/10 | GORM Plugin 自动记录，满足 CNAS 追溯要求 |
| 文件管理 | 5/10 | MinIO 集成但缺少文件类型校验和安全管理 |
| 前端架构 | 6/10 | Vue 3 + Element Plus 选型正确，但缺少状态管理、TypeScript 类型覆盖不足 |
| 部署与运维 | 6/10 | Docker Compose 完善，但缺少 CI/CD、监控、日志聚合 |

## 100 人中小企业适用性评估

**基本可用，但需补充**：
- ✅ 当前架构支持 100 人并发使用（Go + PostgreSQL 性能足够）
- ✅ 7 部门 RBAC 模型适配中小实验室组织结构
- ⚠️ 缺少多实例部署支持（Token Blacklist 需 Redis）
- ⚠️ 缺少数据备份和恢复策略
- ⚠️ 缺少操作手册和用户培训文档

## 与业界标准 LIMS 对比缺失功能

| 缺失功能 | 优先级 | 说明 |
|----------|--------|------|
| 电子签章/数字签名 | P0 | CNAS 合规必备 |
| 消息通知/待办提醒 | P0 | 日常运营必备 |
| 客户自助委托/进度查询 | P1 | 提升客户体验 |
| 仪器校准/期间核查管理 | P1 | CNAS 要求 |
| 标准物质/标准溶液管理 | P1 | 实验室基础管理 |
| 检测方法验证/确认 | P1 | 技术能力管理 |
| 不符合工作控制 | P1 | 质量体系核心要素 |
| 数据趋势分析/统计报表 | P1 | 管理决策支持 |
| 条码/二维码样品追踪 | P2 | 提升样品管理效率 |
| 多实验室/多分支支持 | P2 | 业务扩展需求 |
| ELN（电子实验记录本） | P2 | 原始记录数字化 |
| 能力验证/实验室间比对 | P2 | CNAS 要求 |

---

**总体评价**：项目架构设计合理，Go + Vue 3 技术栈选型恰当，16 节点工作流引擎是核心亮点（含乐观锁、SoD 检查、事务保证）。Phase 1-9 全流程贯通完成度较高。主要短板在安全防护（凭据管理、权限细粒度）、前端状态管理、测试覆盖率和行业标准功能（电子签章、通知、统计）方面。建议优先解决安全问题（问题 1-7），然后推进 Phase 10 检验单集成和消息通知功能。