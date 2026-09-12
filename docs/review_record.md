# LIMS 实验室管理系统代码评审报告（v4 版本对比）

## 一、问题清单（按严重性排序，基于最新代码 `3599885`）

### 问题 1（严重・功能缺失）：生产部署四轮未修——一键部署依旧不可用

- **位置**：`cmd/server/main.go:88`（production 跳过 autoMigrate）+ `deploy/init-db/init.sql`（仍只有 2 行 CREATE EXTENSION）+ 全仓无迁移工具。
- **问题**：本轮把 18 张表全部加上外键约束，但约束只由 AutoMigrate 落地，而 **production 模式明确跳过 AutoMigrate**；init.sql 依旧不建表。按 README/docker-compose 部署后数据库 0 张表，`preMigrateCleanup` 也会因表不存在直接报错。**四轮迭代，这个"上线即坏"的问题原样保留。**
- **建议**：引入 golang-migrate/goose 生成完整建表+外键迁移脚本，production 走迁移；或让 init.sql 完整初始化。

### 问题 2（严重・功能缺失）：文件上传/存储四轮未动——LIMS 核心资产链依旧断裂

- **位置**：后端无 upload/download 路由、无 MinIO SDK 调用；`FileUpload.vue` 无可用 action；`contract_file_path`/`sample_photos`/`report_content`/`sign_stamp` 等仍是手填路径。
- **影响**：原始记录、合同、采样照片、报告电子档、电子印章无法真实落盘。对检测机构而言这是**最核心的合规资产**，缺失意味着"流程走得通、证据留不下"。

### 问题 3（重要・正确性）：SoD 仅检查相邻前序节点——两步以上"自己审自己"仍可穿透

- **位置**：`engine.go:46-54`（`sodCheckNodes` 只覆盖 4 对相邻节点）；`engine.go:607-624`（`checkSoD` 只查上一节点的 assignee）。
- **问题**：CNAS 要求"审核人 ≠ 录入人 且 ≠ 复核人"。当前实现只拦截"审核人==直接前序复核人"；**数据审核人仍可与 2 步前的数据录入人是同一人**（录入 → 复核 → 审核链路中，审核只查复核）。同理报告审核只查报告复核，不查报告编制。
- **建议**：SoD 校验扩大到该实例中"同类型业务链"的全部已完节点操作人（至少包含录入/复核/审核三人互异）。

### 问题 4（重要・正确性）：驳回目标后端已支持，前端仍未传——死 UI 第四轮

- **位置**：`frontend/src/api/workflow.ts:43-45`（`rejectTask` 只传 comment）；`frontend/src/views/business/pendingTasks/index.vue:103`（`handleApprovalSubmit` 收到 `reject_target` 后**丢弃**）；各业务视图的 Reject 调用未带 `reject_target`。
- **问题**：后端（workflow handler + 16 个业务 handler）已全部支持 `reject_target` 并透传引擎，但**前端没有任何一处发送该字段**——用户仍无法选择"驳回到哪一节点"。本轮只改了后端。

### 问题 5（重要・正确性）：admin 跨部门审批行为不一致——业务页按钮 403，通用接口却放行

- **位置**：`engine.go:127-131`（admin 豁免部门校验）；`business_routes.go`（各节点 approve/reject 挂 `DeptScopeMiddleware`，admin 部门=业务室 → 非业务室节点直接 403）。
- **问题**：admin 通过"待办页"（通用 `/workflow/tasks/:id/approve`，无 DeptScope）可跨部门审批；但在各**业务页面**点审批按钮（走 `business:*` 路由，有 DeptScope）会被 403。同一角色两种结果，用户会困惑；README"admin 驱动全流程"依旧未同步。
- **建议**：`DeptScopeMiddleware` 对 admin 放行，或统一走引擎层校验（引擎已做 admin 豁免）。

### 问题 6（一般・正确性）：task_assign 指派失败时响应拼接出非法 JSON

- **位置**：`task_assign_handler.go:170-174`：
  ```go
  if err := h.svc.AssignNextNodeTaskByOrder(...); err != nil {
      c.Writer.WriteString(`{"warning":"任务分配完成，但指派失败: ` + err.Error() + `"}`)
  }
  utils.Success(c, ...)
  ```
- **问题**：先写 body 再 `c.JSON` → 响应为 `{"warning":...}{"code":0,...}` 非法 JSON；且把 err 原文拼进响应有信息泄露面。指派失败应在事务内回滚或单独返回错误码。

### 问题 7（一般・安全）：种子角色与节点部门不匹配——新部署部分节点无人能审批

- **位置**：`seed.go` 角色定义 vs `definition.go` 节点部门：合同评审=dept_tech，但唯一含 `business:contract-review` 的角色是"业务经理"（业务室）；质控任务=dept_qc，但 `lab_technician`（实验室）却含 `business:qc-task`。
- **问题**：`DeptScopeMiddleware` 要求审批人部门==节点部门。技术室用户没有可用的"合同评审"种子角色，业务室业务经理持权限却过不了部门校验 → **全新部署下技术室无人能审批合同评审、质控任务无对应部门角色**，需 admin 手工造角色。
- **建议**：种子角色按"部门 × 节点"对齐（技术室→合同评审/报告签发、质控室→质控任务等）。

### 问题 8（一般・安全）：SetTrustedProxies(nil) 修复 XFF 绕过，但 Nginx 后全司共享限流配额

- **位置**：`router.go:36`（`r.SetTrustedProxies(nil)`）+ `ratelimit.go:84`（key=ClientIP）。
- **问题**：不再信任任何代理后，经 Nginx 反代的请求全部显示为 nginx 内网 IP → **10 次/分钟的登录限流变成全公司共享配额**，办公室多人同时登录即集体 429；同时账号级限流（`user:username`）尚未覆盖错误密码累计（仅登录尝试计数）。
- **建议**：`SetTrustedProxies([]string{"nginx内网IP"})` + 读取 `X-Real-IP`；账号级限流改为"失败次数累计，超限锁定 N 分钟"。

### 问题 9（一般・安全）：JWT 权限快照/自续期/内存黑名单依旧

- **位置**：`utils/jwt.go`（权限写 claims，24h 快照）；`auth_handler.go` Refresh（旧 claims 续签）；`utils/token_blacklist.go`（进程内 map，重启失效、多实例失效）。
- **问题**：角色变更不即时生效；token 可无限续期；登出黑名单重启即失效。四轮未动。

### 问题 10（一般・工程）：零测试——四轮迭代后仍为 0 个测试文件

- **位置**：全仓 0 个 `*_test.go`。本轮新增了 SoD、指派、乐观锁、原子化事务等**高风险逻辑**，却无任何单元/集成测试；`preMigrateCleanup` 的破坏性 DELETE 也无测试保护。
- **建议**：至少为 `engine.ApproveTaskWithTx`（dept/assignee/SoD/乐观锁四类分支）与 `HandleWorkflowError` 补单元测试。

### 问题 11（一般・数据模型）：jsonb 明细与无样品主表依旧

- **位置**：`business_models.go`（test_items/sample_codes/sampling_points/attachments 等 jsonb）；无 Sample 主表。
- **问题**：无法按样品/项目/客户做统计；样品台账（条码/留样/处置）仍是缺失功能清单第一名。

### 问题 12（一般・工程）：preMigrateCleanup 每次启动静默删孤儿行

- **位置**：`router.go:133-180`。
- **问题**：每次 dev 启动执行 16 条 `DELETE ... WHERE task_order_id NOT IN (...)`，若出现数据错位（如业务记录先于 task_order 创建）会**静默删除**且无备份；建议改为启动时校验+日志，或只提示不删除。

### 问题 13（一般・功能）：部门待办之外，"我的待办"覆盖仍有限

- **位置**：`task_assign_handler.go`（仅把"任务分配"的**下一节点**指派到人）；`engine.go`（其余节点审批时才 COALESCE 写入 assignee）。
- **问题**：复核/审核等节点的任务在审批前没有 assignee → "我的待办"仍主要展示数据录入任务；复核/审核人仍需依赖部门待办列表人工认领。可接受，但与"任务到人"的行业标准仍有距离。
