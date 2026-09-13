# LIMS 第三方检测实验室管理系统 · 代码评审报告

## 二、问题清单（按严重性排序）

### 问题 1（严重・安全）：业务数据跨部门越权读——List 无部门 scope、Get 无归属校验（上线阻断）
**位置**：`backend/internal/handler/generic.go::List/Get` + 全部 16 个业务 handler（如 `data_entry_handler.go::List/Get`、`report_prepare_handler.go`、`field_sampling_handler.go` 等）

**机理**：`business_routes.go` 中，所有节点的 `GET ""`（列表）与 `GET /:id`（详情）**只挂了 `PermissionMiddleware`（权限码），没有 `DeptScopeMiddleware`**——部门校验只挂在 `approve/reject` 上。而 `DataEntryHandler.List` 的 scope 仅按 `task_order_id/test_item_id` 过滤，`DataEntryHandler.Get` 直接透传 `GenericHandler.Get`（`First(&item, id)`，无任何归属条件）。后果：

- 持有 `business:data-entry` 权限码的任意部门用户，可**分页拉取全库检测数据**、按 ID 读**任意**检测原始数据/报告/样品信息；
- 权限码是全局角色授予的（`permissions` 表无部门维度），同一码可被多部门持有（如实验室主管同时持 data-entry/data-review/data-audit）；
- 检测数据、客户报告在部门间完全开放——对多部门协作的检测实验室，**这是数据机密性硬伤**（也违反 CNAS 保密性要求）。

**修复**（两层）：
1. `GenericHandler.Get` 增加归属 scope 回调（与 List 同构），各 handler 传入 `dept_id` 或"本部门参与的任务"条件；
2. 或复用引擎已有的 `GetProgressByBusiness` 校验模式（`engine.go:496`：非 admin 且非创建者时校验 `assignee_dept_id`）——把该模式下沉为通用中间件/scope 工具。

### 问题 2（严重・安全）：`/api/workflow/*` 无权限码 + 流程实例/历史 IDOR
**位置**：`backend/internal/router/routes/workflow_routes.go`（仅 `AuthMiddleware`）；`backend/internal/workflow/engine.go::GetInstance`（L475，无归属校验）、`GetProcessHistory`（L434，无归属校验）

**机理**：
- `POST /workflow/tasks/:id/approve|reject` 依赖引擎内部部门校验兜底（`engine.go:134` `AssigneeDeptID != userDeptID` → 拒绝），功能安全但**与业务路由的权限模型不一致**——业务侧有权限码，流程侧任何登录用户都能"尝试"审批任意任务（靠引擎拒绝），体验与语义混乱；
- `GET /workflow/instances/:id`、`GET /workflow/instances/:id/history`、`GET /workflow/progress/:businessType/:businessId` 中，前两者**完全无归属校验**：任意登录用户可按 instanceID 枚举读取任意项目的流程详情与**全部节点经办人、审批意见**（`GetProcessHistory` 返回 Operator/Dept/Comment）。

**修复**：workflow 组挂 `PermissionMiddleware`（如 `workflow:task`、`workflow:view`）；`GetInstance/GetProcessHistory` 增加"创建者/本部门参与/管理员"三选一校验（复用 `GetProgressByBusiness` 模式）。

### 问题 3（重要・功能）：文件上传/文档管理整体缺失——D1-D14 文档无落地载体（CNAS 核心诉求未闭合）
**位置**：后端 `grep multipart/SaveUploadedFile/MinIO` **零命中**；`model/workflow.go::ProcessTask.OutputDocPath` 字段存在但无上传 API；前端仅 `FileUpload.vue` 通用组件（`uploadUrl` 无真实后端地址）

**机理**：README 架构图标注 MinIO 文件存储、Phase 3 声称"上传通用组件"完成——但**只做了前端组件**。委托单、合同、采样记录、原始记录、报告等 D1-D14 文档无法上传/关联到流程任务，`OutputDocPath` 永远是空。项目归档节点（节点 16）名存实亡。

**修复**：新增 `POST /api/files`（multipart，校验扩展名白名单/MIME/大小上限）+ `GET /api/files/:id`（带权限与归属校验，**不要做成公开 URL**）+ 上传后写回 `process_tasks.output_doc_path`；MinIO 客户端（`minio-go`）接入，存储桶按部门/项目隔离。

### 问题 4（重要・安全）：登录暴力破解防护偏弱——60 次/分钟 + 无验证码
**位置**：`backend/internal/router/routes/system_routes.go:17`（`NewRateLimiter(time.Minute, 60)`）；`backend/internal/middleware/ratelimit.go`

**机理**：限流窗口 1 分钟 60 次（账号级），bcrypt 单次验证 ~100ms，60 次/分钟意味着攻击者可持续试错而不触发长时间锁定（锁定时间取决于 `RecordFail` 阈值，未配置时偏宽松）；无验证码、无 IP 级限制。

**修复**（对 100 人内部系统中等优先级）：限流收紧到 `5 次/分钟` + `RecordFail` 后锁定 15 分钟；登录接口加简单验证码或至少 IP 维度限流。

### 问题 5（重要・安全）：前端 token 存 localStorage——XSS 即窃取
**位置**：`frontend/src/api/request.ts`（`localStorage.getItem('token')`）、`stores/user.ts`

**机理**：JWT 存 localStorage，任何前端 XSS（报告模板渲染、文件名回显等）可直接读取 token；无 httpOnly cookie、无 refresh token 机制。

**修复**：短期（当前架构）至少设置较短 token 过期（`config.go::JWT.ExpireHour` 收紧）并提示风险；中期迁移到 httpOnly cookie + CSRF 防护，或短期 access + refresh token 双令牌。

### 问题 6（一般・功能）：报告生成与电子签章缺失——签发节点无实质产出
**位置**：`report_sign_handler.go`、`report_prepare_handler.go`

**机理**：节点 11-14（编制→复核→审核→签发）只是流程状态流转，**无 PDF 报告模板引擎、无电子签章/防伪**。第三方检测机构（CNAS 场景）报告签发是核心交付物，当前系统发不出报告文件。

**建议**：接入报告模板（如 Go 模板 + wkhtmltopdf/weasyprint 或 Chromium 渲染）+ 签章图片合成；至少先实现"报告 PDF 生成 + 签发人签章图叠加"，后续再考虑 CA 数字签名。

### 问题 7（一般・功能）：相对业界 LIMS 缺失的能力清单
| 缺失项 | 业界标准 | 当前状态 |
|---|---|---|
| 样品条码/标签打印 | 样品接收即生成条码+打印标签 | `sample_receiving_handler.go` 仅编码登记 |
| 仪器数据对接（LIS/串口/文件） | LIMS 核心价值之一，自动采集检测数据 | **无** |
| 待办通知（邮件/站内/短信） | 流程推进自动通知 | **无**（靠人肉刷新待办） |
| 客户门户（自助查进度/报告） | 标配 | 无 |
| 经营统计/质量看板 | 委托量、检测周期、返工率、报告及时率 | 无（仅 `dept` CRUD） |
| 报告模板版本管理 | 模板+版本+审批 | 无 |
| 批量委托/批量导入 | 高频操作 | 无（基础数据有 Excel，业务无） |
| 软删除/回收站 | 审计可追溯 | `Delete` 硬删除（有审计记录但数据不可恢复） |

### 问题 8（一般・工程）：部署与配置细节
**位置**：`deploy/docker-compose.yml`（PG/MinIO 密码 `lims123` 明文）、`config.go`（JWT Secret 默认值——需确认生产是否强制环境变量注入）

**建议**：compose 密码改环境变量占位；`JWT.Secret` 生产环境启动时校验非默认值；`gin.SetMode` 已按 Env 切换 ✅；`TrustedProxies` 已配置 ✅。

---

