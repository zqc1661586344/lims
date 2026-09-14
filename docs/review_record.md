# LIMS 实验室管理系统代码评审报告（v8 版本对比）



## 二、问题清单（按严重性排序，基于最新代码 `76cbf02`）

### 问题 1（严重・正确性/回归）：workflow:view / workflow:task 权限码未绑定任何业务角色——非 admin 用户工作流功能全部 403

- **位置**：`routes/workflow_routes.go:22-40`（viewG 挂 `PermissionMiddleware(db,"workflow:view")`，taskG 挂 `"workflow:task"`）；`seed/seed.go:43-77`（13 个业务角色的 codes 列表**只含 business:\* 权限，没有 workflow:view/workflow:task**；仅 `super_admin` 通过 `"*"` 通配绑定）。
- **影响**：所有非 admin 用户——
  - `GET /workflow/tasks/pending`、`/tasks/pending/user`（**待办中心**）→ 403"无权限访问"；
  - `GET /instances/:id/history`（流程历史）、`/progress/:businessType/:businessId`（**进度条**）、`/nodes` → 403；
  - `POST /tasks/:id/approve|reject`（**通用审批通道**）→ 403。
  - 业务页面直连的 `/business/*/approve` 不受影响（挂业务权限），但**待办中心、ProcessTimeline、TaskProcessBar 全线瘫痪**。v7 中这些接口仅有 AuthMiddleware、全部可用——**本轮回退**。
- **验证方法**：`grep workflow:view seed.go` 仅命中权限定义，无绑定；`PermissionMiddleware`（`permission.go:28-40`）严格在用户权限列表中查找要求码。
- **修复**：① 把 `workflow:view`/`workflow:task` 追加到全部 13 个业务角色 codes（seed 幂等 Replace 可补齐）；② 或去掉这两个权限码、仅保留 AuthMiddleware（引擎层已有 dept/assignee/SoD 四重校验，权限码是冗余）；③ 必须补一条回归测试："业务角色可访问 /workflow/tasks/pending/user"。

### 问题 2（重要・设计反转/回归）：admin 豁免被系统性移除——admin 业务可见性收缩到仅业务室

- **位置**：`middleware/dept_scope.go:12`（**v5 明确修复的 `IsAdmin` 放行被删除**）；`workflow/engine.go:120-135`（`ctxIsAdmin` 删除后 `ApproveTaskWithTx/RejectTaskWithTx` 无条件要求部门+指派匹配）；`seed/seed.go:225-234`（admin 用户部门 = `dept_business`）。
- **影响**：admin（属业务室）现在：
  - 访问技术室/质控室/实验室/报告室等 15 个节点业务路由 → `DeptScopeMiddleware` 403；
  - 审批非业务室任务 → 引擎 `ErrDeptNotMatch/ErrAssigneeNotMatch` 403——**v5 曾修复"admin 业务页 403"，本轮回退**；
  - 与 `seed.go:78` super_admin 角色备注"**拥有全部业务权限（不受部门约束）**"自相矛盾。
- **判断**：若为有意设计（CNAS 严格化：admin 也不得越权），需同步修改角色备注与 README 并明确"admin 仅系统管理 + 业务室业务"；若为无意回退，应恢复 admin 在 DeptScope 的放行（引擎层指派校验可保留）。**当前状态是"注释与实现打架"的中间态，必须二选一并落地。**

### 问题 3（严重・功能缺失）：文件上传/存储八轮未动（唯一 P0 继续）

- **位置**：无 upload/download 路由、无 MinIO SDK 调用；`contract_file_path`/`sample_photos`/`sign_stamp` 等仍为手填路径；`FileUpload.vue` 无可用 action。
- **影响**：原始记录、合同、报告电子档、电子印章无法真实落盘，CNAS/CMA 存档闭环不成立。

### 问题 4（一般・正确性）：JSONB 空值写 NULL——与 '{}' 初始化并存，前端需判空

- **位置**：`model/jsonb.go:14-17`（`Value()` 对空返回 `nil`）；`business_service.go` `ensureBusinessRecord`（仍初始化 `model.JSONB("{}")`）。
- **问题**：新写入的空明细字段落库为 **NULL**，旧数据与 `ensureBusinessRecord` 初始化为 **'{}'**——同列两种语义。前端若按字符串处理（`JSON.parse(row.sample_codes)`）遇 null 会抛错；`MarshalJSON` 空时返回 `null` 也放大了该风险。
- **修复**：`Value()` 空时返回 `[]byte("{}")`（与初始化口径一致），或前端统一判空。

### 问题 5（一般・工程）：engine_test 与 SoD 新链脱节——测试断言过时且无新链用例

- **位置**：`workflow/engine_test.go:131-137` `TestSoDCheck_SkipsNonSoDNodes` 仍断言 `NodeContractReview`/`NodeReportSign` 为非 SoD 节点，与 `engine.go` 新增 3 条 SoD 链矛盾（因测试库无任务行而"假绿"）；无新链的正向/负向用例。
- **修复**：更新该测试 + 为合同评审≠任务创建、样品接收≠现场采样、报告签发≠报告编制补用例（本轮 SoD 改动是合规敏感逻辑，必须测）。

### 问题 6（重要・工程）：测试覆盖仍仅 workflow 包（八轮未扩）

- **位置**：`go test ./...` 仅 `internal/workflow` ok。本轮新增的**权限码校验（403 回归正是因无测试而漏网）、ApplyTaskOrderScope 数据范围、canAccessInstance、JSONB Scan/Value、路由权限分组**全部无测试。
- **建议**：优先补 ①"业务角色待办接口 200"（防问题 1 复现）；②"非本部门用户跨任务单读取 403"；③"admin 业务路由行为"（固化问题 2 的决策）。

### 问题 7（一般・安全）：seedUsers 生产演示账号未动（v6 问题 2 持续）

- **位置**：`seed.go:170-243`（13 个 `<username>123` 账号生产 migrate 同样创建）。建议仅 development seed 或强制首登改密。

### 问题 8（一般・工程）：内存锁定/黑名单 + Refresh 无限续期（持续）

- **位置**：`ratelimit.go`、`utils/token_blacklist.go`（进程内）；`auth_handler.go` Refresh。权限已每请求 DB 刷新，建议剩余时效项一并下沉 Redis/DB。

### 问题 9（一般・工程）：非版本化迁移 + 自动派单 LIMIT 1 + 锁定计数共用桶（持续）

- **位置**：`cmd/migrate/main.go`（AutoMigrate 一次性）；`engine.go:692-706` `resolveAssigneeForNode`（LIMIT 1 无负载均衡）；`ratelimit.go`（成功登录累计进失败计数）。
