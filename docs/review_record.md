
## 问题清单（按严重性排序）

### 问题 1（严重・安全）：系统管理路由连续三轮无权限控制——提权漏洞依旧
`backend/internal/router/routes/system_routes.go:38-70`：`/api/system/users|depts|roles|permissions` 全部 CRUD 仍只挂 `AuthMiddleware`。任意登录用户仍可创建管理员、给自己分配角色。**三轮迭代，第一轮就指出的最严重问题原样保留。**

### 问题 2（重要・安全）：部门内职责分离（SoD）未实现，通用工作流接口无权限码
`workflow_routes.go:24-30` 的 approve/reject 未挂权限中间件；`engine.go:108` 只校验部门不校验操作人。DeptScope 加强后，**实验室录入员仍可直接调通用接口审批同部门"数据审核/报告复核"**；"复核人≠录入人"的 CNAS 双人复核要求未实现。

### 问题 3（重要・正确性）："我的待办"三轮未修——`engine.AssignTask` 依旧 0 调用方
`task_assign_handler.go:41-57` 仍只写 `AssignedTo` 字符串，`assignee_user_id` 从不写入，非 admin 的"我的待办"恒空。

### 问题 4（重要・安全）：登录限流仍可被 X-Forwarded-For 伪造绕过
`ratelimit.go:72-79` 用 `c.ClientIP()`，router 未调 `SetTrustedProxies`（gin 默认信任所有代理）→ 轮换 XFF 头即可无限爆破；按 IP 不按账号。

### 问题 5（重要・正确性）：审计日志操作人仍恒为 "system"
`audit.go:202` 的局部 `type ctxKey` 与 `gorm_context.go:10` 包级 `ctxKey` 类型不一致，context 查询必然失败。CNAS"操作可追溯"依旧失效。

### 问题 6（重要・正确性）：驳回目标选择三轮未打通
引擎支持 `rejectTargetOverride`，但 workflow handler、16 个业务 Reject handler、`workflow.ts` 都不接收该字段；`ApprovalDialog.vue` emit 的 `reject_target` 在 `pendingTasks/index.vue:103` 被接收后**丢弃**。

### 问题 7（重要・数据完整性）：业务保存与审批仍未原子化
`ApproveWithBusiness`/`RejectWithBusiness`（单事务封装）依旧无 handler 使用；各 Approve 仍是 `FirstOrCreate` 落库后另开事务推进流程。

### 问题 8（重要・正确性）：admin 跨部门操作被设计性关闭，但 README/UI 未同步
`engine.go:108` 严格 dept 校验无 admin 豁免（`permission.go:16-20` 注释确认是有意收紧）；README"单 admin 驱动全流程"失效未更新；`ErrDeptNotMatch` 仍包装成 HTTP 500 而非 403。

### 问题 9（重要・功能缺失）：文件上传/存储三轮未动——业务闭环仍断裂
无 upload 路由、无 MinIO 调用，合同/采样/报告/盖章字段仍为手填路径。

### 问题 10（一般・工程）：生产部署三轮未动——一键部署仍不可用
production 跳过 AutoMigrate + init.sql 只建扩展 + 无迁移工具 → 数据库 0 张表。

### 问题 11（一般・安全）：`config.dev.yaml:11` 真实数据库密码 `Zqc166879@` 仍入库

### 问题 12（一般・工程）：零测试 + 21.7MB 二进制仍误提交 + README 脱节

### 问题 13（一般・工程）：权限码未种子化，冷启动无默认角色模板

### 问题 14（一般・架构）：模型层三轮未动——jsonb 滥用、无样品主表、无外键

### 问题 15（一般・安全，本轮新发现）：流程进度接口无鉴权
`GET /workflow/progress/:businessType/:businessId`（`engine.go:451-457`）任意登录用户可枚举查看任意委托的流程进度与驳回意见；`First(&instance)` 无排序，多实例时结果不稳定。
