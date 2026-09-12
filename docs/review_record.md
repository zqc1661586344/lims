# LIMS 实验室管理系统代码评审报告（v3 版本对比）

## 一、问题清单

### 问题 1（严重・安全）：系统管理路由连续三轮无权限控制——提权漏洞依旧

- **位置**：`backend/internal/router/routes/system_routes.go:38-70`。`/api/system/users|depts|roles|permissions` 全部 CRUD 仅挂 `AuthMiddleware` + `GormContextMiddleware`，注释"authenticated + admin by default"与实际不符。
- **问题**：业务路由本轮已强化到部门级，但**系统管理仍人人可进**：任意登录用户可 `POST /system/users` 建管理员、`PUT /system/users/:id/roles` 给自己赋任意角色。**第一轮就指出的提权路径，三轮迭代后原样保留**，这是当前最严重的问题。
- **建议**：systemGroup 挂 `PermissionMiddleware(db, "system:users")` 等码，或独立 `RequireAdmin` 中间件 + 单测。

### 问题 2（重要・安全）：部门内职责分离（SoD）仍未实现，通用工作流接口无权限码

- **位置**：`backend/internal/router/routes/workflow_routes.go:24-30`（approve/reject 未挂 PermissionMiddleware/DeptScope）；`backend/internal/workflow/engine.go:108,188`（只校验部门，不校验操作人）。
- **问题**：`DeptScopeMiddleware` 把业务路由的审批限制到责任部门后，**同一部门内任何成员仍可审批该部门任意节点的任务**：实验室录入员可以直接调 `/workflow/tasks/:id/approve` 审批同部门的"数据审核/报告复核"，绕过 `business:report-review` 权限码；且**没有"复核人 ≠ 录入人、审核人 ≠ 复核人"的人员级互斥校验**（CNAS/CMA 双人复核硬性要求）。
- **建议**：通用工作流接口同步挂权限码 + 部门域；服务层增加"当前节点业务记录的操作人 != 当前用户"校验；`assignee_user_id` 落地后校验"只能操作分配给我的任务"。

### 问题 3（重要・正确性）："我的待办"三轮未修——AssignTask 机制依旧无人接线

- **位置**：`backend/internal/workflow/engine.go:333-353`（`AssignTask`，全仓 0 调用方）；`backend/internal/handler/task_assign_handler.go:41-57`（`Create` 仍只写 `AssignedTo` 字符串）。
- **问题**：`assignee_user_id` 依旧从不写入 → `GetPendingTasksByUser` 恒空。任务分配节点没有"选择实验员"的 API/UI，数据录入无按人认领。"我的待办"对非 admin 用户仍是空列表。
- **建议**：task-assign 的 Create/Update 增加 `assignee_user_id` 并调用 `engine.AssignTask`；前端任务分配表单加人员选择。

### 问题 4（重要・安全）：登录限流仍可被 X-Forwarded-For 伪造绕过

- **位置**：`backend/internal/middleware/ratelimit.go:72-79`（`key := c.ClientIP()`）；`backend/internal/router/router.go`（未调用 `SetTrustedProxies`）。
- **问题**：gin 默认信任所有代理，`c.ClientIP()` 取 XFF 首值。部署于 Nginx 后，攻击者轮换 `X-Forwarded-For` 即可无限尝试；限流按 IP 不按账号，内网 NAT 场景易误伤。
- **建议**：`r.SetTrustedProxies(nil)` + 读取 Nginx 注入的 `X-Real-IP`；增加按账号维度限流与失败锁定。

### 问题 5（重要・正确性）：审计日志操作人仍恒为 "system"（ctxKey 类型不一致三轮未修）

- **位置**：`backend/internal/middleware/audit.go:200-214`（`extractOperator` 内局部 `type ctxKey string`）；`gorm_context.go:10-15`（包级 `ctxKey`）。
- **问题**：context 按动态类型匹配 key → 查询必然失败 → 每条审计日志 `operator_id=0, operator="system"`。**CNAS"操作可追溯"核心证据依旧失效**；且引擎 `tx.Exec` 更新 process_tasks 不走 GORM 钩子，审批动作本身不在 audit_logs。
- **建议**：将 `extractOperator` 改用包级 ctxKey（或直接复用 gorm_context 导出的 key）；引擎原生 SQL 迁移到 `clause`/`Model().Updates()` 以触发审计钩子。

### 问题 6（重要・正确性）：驳回目标选择三轮未打通——后端支持、入口不接收

- **位置**：引擎 `rejectTargetOverride` 已支持（`engine.go:193-197`），但：`workflow_handler.go:141-150` 请求体只有 comment；16 个业务 Reject handler 同；`frontend/src/api/workflow.ts:43-45` 的 `rejectTask` 只传 comment；`ApprovalDialog.vue:119` 虽然 emit 了 `reject_target`，`pendingTasks/index.vue:103` 接收后**丢弃**。
- **问题**：用户选择"驳回到 XX 节点"仍无效，一律走 definition 硬编码单点回退。

### 问题 7（重要・数据完整性）：业务保存与流程审批仍未原子化

- **位置**：`business_service.go:121-187`（`ApproveWithBusiness`/`RejectWithBusiness` 单事务封装依旧无 handler 使用）；`contract_review_handler.go:156-163`（`FirstOrCreate` 落库后另开事务审批）。
- **问题**：保存成功、审批失败（dept 不匹配/乐观锁冲突）时业务表与流程状态不一致。乐观锁解决了并发抢批，但**跨事务一致性问题仍在**。

### 问题 8（重要・正确性）：admin 跨部门操作被设计性关闭，但 README/UI 未同步

- **位置**：`engine.go:108,188`（严格 dept 校验，无 admin 豁免）；`permission.go:16-20` 注释明确"DeptScope 仍限制 admin 于本部门"；`auth_handler.go:134-140`（admin 获全部权限码但**仍被 DeptScope 拦在部门内**）。
- **问题**：这是**有意的安全收紧**（方向正确），但：① README "Admin 可跨部门驱动全流程" 承诺失效且未更新；② admin 在"部门待办"能看到全部任务却只能批业务室任务，点击其余报错；③ `ErrDeptNotMatch` 仍被包装成 HTTP 500（`workflow_handler.go:126`），应映射 403。

### 问题 9（重要・功能缺失）：文件上传/存储三轮未动——业务闭环仍断裂

- **位置**：后端无 upload/download 路由、无 MinIO SDK 调用；`FileUpload.vue` 无可用 action；合同/采样/报告/盖章字段仍为手填路径。
- **影响**：原始记录、合同、样品照片、报告电子档、电子印章无法真实落盘，LIMS 核心资产链缺失。

### 问题 10（一般・工程）：生产部署三轮未动——一键部署仍不可用

- **位置**：`router.go:22-26`（production 跳过 AutoMigrate）+ `deploy/init-db/init.sql`（仅建扩展）+ 无迁移工具 + `docker-compose.yml`（`LIMS_ENV=production`）。
- **影响**：官方部署路径下数据库 0 张表。本轮改动完全未触及部署层。

### 问题 11（一般・安全）：config.dev.yaml 真实数据库密码仍入库

- **位置**：`backend/config/config.dev.yaml:11`（`password: Zqc166879@`）。两轮未修。

### 问题 12（一般・工程）：零测试 + 21.7MB 二进制误提交 + 文档脱节

- **位置**：全仓 0 个 `*_test.go`；`backend/server` 仍在 git 跟踪（`.gitignore` 只忽略 `lims-server`）；`README.md` 未同步（admin 驱动、文件存储、Excel 导入导出等承诺与实际不符）。

### 问题 13（一般・工程）：权限码未种子化——冷启动无默认角色模板

- **位置**：`seed.go` 仅建 7 部门 + admin，无 permissions/roles/role-permission 种子。新部署需 admin 逐条手工建权限并配角色，且因为问题 1，任何用户都能绕过该配置。

### 问题 14（一般・架构）：模型层未动——jsonb 滥用、无样品主表、无外键

- **位置**：`model/business_models.go`（test_items/sample_codes/attachments/report_content 等 jsonb 明细）；无 Sample 主表；`ProcessInstance.BusinessID` 无 FK。三轮未动。

### 问题 15（一般・安全/健壮性，本轮新发现）：流程进度接口无鉴权、数据可见性无控制

- **位置**：`workflow_routes.go:35`（`GET /workflow/progress/:businessType/:businessId`）；`engine.go:451-457`（`GetProgressByBusiness` 用 `e.db` 直接查，无部门/权限过滤）。
- **问题**：任意登录用户可枚举 `businessType/businessId` 查看**任意委托的流程进度**（含当前节点、驳回意见），属轻量信息泄露；且 `First(&instance)` 无排序，同一业务多次实例时返回不稳定。

