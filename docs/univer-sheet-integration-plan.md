# Univer Sheet 集成实施计划

> **状态**: 方案已确认，待实施
> **目标版本**: LIMS v1.1
> **预估工时**: 5.5-6.5 人天

---

## 1. 背景与需求

### 1.1 现状

当前系统在数据录入等环节使用 `textarea` 手动输入 JSON 来存储原始检测数据：

```
┌──────────────────────────────────────┐
│ 原始数据 [ { "pH": 7.2, "COD": 45 } ] │ ← textarea 手输 JSON
└──────────────────────────────────────┘
```

这种方式存在明显问题：
- 用户体验差（手输 JSON 容易出错）
- 没有结构化展示（表格形式一目了然）
- 不支持公式计算（实验室检测常需要稀释倍数、平均值、标准偏差等自动计算）
- 没有模板复用能力（不同检测项对应不同表格结构）

### 1.2 目标

为 LIMS 系统引入在线检验单编辑器，替代 textarea JSON 输入：

```
┌──────────────────────────────────────┐
│  [A] 样品编号 │ [B] CODcr │ [C] 计算值 │
│   S001        │   45.2    │ =B2*1.5    │ ← 在线表格 + 公式实时计算
│   S002        │   38.7    │ =B3*1.5    │
└──────────────────────────────────────┘
```

### 1.3 核心需求

| 需求维度 | 具体要求 |
|---------|---------|
| 表格编辑 | 可编辑单元格、行增删改、列增删改 |
| 模板机制 | 不同检测项 → 不同表格结构（COD 检测表和氨氮检测表结构不同） |
| 公式计算 | 单元格支持简单公式（如 `=B2*C2`、`=SUM(B2:B10)`、`=AVERAGE(C2:C8)`），实时计算 |
| 数据持久化 | 表格数据存后端，下次打开还能看到 |
| 只读/编辑模式 | 待办任务环节可编辑，已归档后只读 |
| Vue 3 集成 | 与现有 Vue 3 + TS + Element Plus 无缝集成 |
| 后端存储 | 后端已有 PostgreSQL，利用 jsonb 字段存储 |

---

## 2. 技术选型

### 2.1 候选方案对比

| 维度 | A. Univer Sheet | B. Handsontable | C. x-spreadsheet | D. 自研 + formula-parser |
|------|----------------|-----------------|------------------|------------------------|
| 功能完整度 | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐ | ⭐⭐⭐ |
| Vue3 集成难度 | 中 | 低 | 中 | 低 |
| 公式能力 | ⭐⭐⭐⭐⭐（200+ 函数） | ⭐⭐⭐⭐（Pro 版完整） | ⭐⭐（仅基础运算） | 按需 |
| 模板机制 | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐ | 自定义 |
| 学习成本 | 中偏高 | 低 | 低 | 高 |
| 维护状态 | 🟢 活跃（字节跳动） | 🟢 活跃 | 🟡 更新不频繁 | 自己维护 |
| 免费 | ✅ Apache 2.0 完全免费 | 社区版 MIT，Pro 版付费 | ✅ MIT | ✅ |
| 体积（gzip） | ~500KB | ~200KB | ~50KB | ~30KB |
| 后端存储复杂度 | 中（snapshot JSON） | 低（二维数组） | 低（二维数组） | 低 |
| LIMS 匹配度 | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐ | ⭐⭐⭐ |

### 2.2 选型结论：Univer Sheet

**理由**：

1. **公式能力完整** — LIMS 实验室检测需要的公式比想象中多：
   - `=C2*D2/E2`（稀释倍数计算）
   - `=STDEV.P(B2:B10)`（标准偏差，质控必备）
   - `=AVERAGE(B2:B10)`（平均值）
   - `=IF(B2>0.5,"超标","合格")`（自动判定合格/不合格）
   - x-spreadsheet 根本算不了，Handsontable 社区版也有限

2. **模板机制天然适合 LIMS** — 每个检测项目对应一个模板 JSON，存在数据库的 `lab_sheet_templates` 表。不同节点（数据录入、数据复核、报告编制）加载同一个模板但读写权限不同。

3. **字节跳动背书** — Univer 是 Luckysheet 的继任者，字节跳动团队维护，持续更新，中文社区活跃。

4. **后端存储友好** — Univer 的表格数据可以序列化为一个 JSON snapshot，直接存在 PostgreSQL 的 `jsonb` 字段里。

5. **完全免费开源** — Apache 2.0 协议，无商业授权压力。

### 2.3 体积说明

Univer Sheet gzip 后约 500KB，属于较大但完全可控的水平：

| 项目 | gzip 体积 |
|------|----------|
| Vue 3 | ~130KB |
| Element Plus | ~800KB |
| **Univer Sheet** | **~500KB** |
| x-spreadsheet | ~50KB |
| Handsontable | ~200KB |

**对当前系统无影响**：
- 内网/办公网 100Mbps 带宽下，500KB 下载 < 0.1 秒
- Vite 代码分割 + 按需加载，不使用 Univer 的页面（Dashboard、待办任务）完全不下载
- 浏览器 HTTP 缓存，首次加载后自动缓存
- Element Plus 本身比 Univer 还大，项目已在使用

---

## 3. 分阶段实施计划

```
阶段 1 (Univer 组件)  ──────┐
                             ├──▶ 阶段 3 (前端页面改造)
阶段 2 (后端 Model+API) ────┤
                             ├──▶ 阶段 4 (流程引擎集成)
                             │
                             └──▶ 阶段 5 (增强功能)
```

---

### 阶段 1：基础搭建（Univer 安装 + 封装通用组件）

**目标**：引入 Univer Sheet，封装一个可复用的 Vue 组件

| 步 | 内容 | 涉及文件 |
|----|------|---------|
| 1.1 | 安装依赖：`npm install @univerjs/core @univerjs/ui @univerjs/sheet @univerjs/engine-formula` | `package.json` |
| 1.2 | 创建 `src/components/LabSheetEditor.vue` 封装组件 | 新建 |
| 1.3 | 支持 `:mode="'edit' | 'readonly'"` 属性 | LabSheetEditor |
| 1.4 | 支持 `:template="templateJson"` 传入模板结构 | LabSheetEditor |
| 1.5 | 支持 `v-model:sheetData` 双向绑定（snapshot JSON） | LabSheetEditor |
| 1.6 | 支持 `@formula-calculated` 事件（公式计算完成后抛出值） | LabSheetEditor |
| 1.7 | 写一个最小 demo 页面验证 Univer 能跑通 | 临时页面 |

**LabSheetEditor.vue API 设计**：

```typescript
// Props
{
  mode: 'edit' | 'readonly'       // 编辑/只读
  template?: TestTemplateSnapshot // 模板 JSON（空 = 空白表）
  sheetData?: UniverSnapshot      // 已有数据（打开旧表时传入）
  height?: string                 // 默认 '500px'
  title?: string                  // 表格标题
}

// Emits
{
  'update:sheetData': (snapshot) => void       // 数据变更
  'cell-selected': (cell) => void              // 选中单元格（用于批注）
  'formula-result': ({cell, value}) => void    // 公式计算结果
}

// Expose
{
  getSnapshot(): UniverSnapshot                // 取当前完整数据
  setReadonlyRanges(ranges: string[]): void    // 设置哪些区域只读（公式列）
  exportExcel(): Blob                          // 导出 Excel（将来用）
}
```

---

### 阶段 2：后端数据模型 + API

**目标**：新增模板表和实例表，支撑模板管理和检验单实例存储

#### 2.1 新增 Model

```sql
-- 检验单模板表
lab_sheet_templates:
  id                uint pk
  code              varchar(64) unique     -- 模板编码: "COD_DETECTION_V1"
  name              varchar(200)           -- 模板名称: "COD 检测原始记录单"
  test_item_id      uint fk                 -- 关联检测项目
  node_code         varchar(64)             -- 对应流程节点: "node_data_entry"
  structure         jsonb                   -- Univer snapshot（预定义的行/列/公式/合并单元格）
  editable_ranges   jsonb                   -- 可编辑区域 ["A2:E20"]
  readonly_ranges   jsonb                   -- 只读区域（公式列）["E2:E20"]
  version           int default 1
  status            int default 1           -- 1=active 0=disabled
  created_at        timestamp
  updated_at        timestamp

-- 检验单实例表
lab_sheets:
  id                uint pk
  task_order_id     uint index              -- 关联委托
  test_item_id      uint index              -- 关联检测项目
  template_id       uint fk                 -- 使用的模板
  node_code         varchar(64)             -- 当前属于哪个流程节点
  sheet_data        jsonb                   -- Univer snapshot（用户填完数据后的完整状态）
  formula_results   jsonb                   -- 公式计算结果缓存 { "E2": 67.8, "E3": 58.0 }
  status            int default 0           -- 0=草稿 1=已提交 2=已复核 3=已审核
  created_by        uint
  updated_by        uint
  created_at        timestamp
  updated_at        timestamp
```

#### 2.2 新增 API

```
模板管理（基础数据）:
  GET    /api/base/lab-sheet-templates                    -- 列表
  POST   /api/base/lab-sheet-templates                    -- 新建
  GET    /api/base/lab-sheet-templates/:id                -- 详情
  PUT    /api/base/lab-sheet-templates/:id                -- 更新（保存模板结构）
  DELETE /api/base/lab-sheet-templates/:id                -- 删除
  GET    /api/base/lab-sheet-templates/by-item/:testItemID -- 按检测项目查模板

检验单实例（业务）:
  GET    /api/business/lab-sheets?task_order_id=&node_code=  -- 列表
  POST   /api/business/lab-sheets                             -- 新建（选模板 → 生成实例）
  GET    /api/business/lab-sheets/:id                         -- 详情
  PUT    /api/business/lab-sheets/:id                         -- 保存（存 Univer snapshot）
  POST   /api/business/lab-sheets/:id/submit                  -- 提交（状态变更）
  POST   /api/business/lab-sheets/:id/calculate               -- 触发公式重算
```

#### 2.3 新增文件

```
backend/internal/model/lab_sheet.go         -- 新建 Model
backend/internal/service/lab_sheet_service.go  -- Service 层
backend/internal/handler/lab_sheet_handler.go  -- Handler 层
backend/internal/router/routes/lab_sheet_routes.go  -- 路由注册
```

---

### 阶段 3：前端页面改造

**目标**：用 LabSheetEditor 替换 textarea，覆盖核心业务页面

#### 3.1 改造清单

| 优先级 | 页面 | 改造内容 | 模式 |
|--------|------|---------|------|
| P0 | **模板管理**（新建页面，在基础数据下） | Univer 所见即所得设计模板 → 保存为 template JSON | edit |
| P1 | **dataEntry**（数据录入） | textarea → LabSheetEditor，加载模板 → 用户填数据 → 保存 snapshot | edit |
| P2 | **dataReview**（数据复核） | 只读模式 + 批注功能（选中单元格加批注） | readonly + 批注 |
| P3 | **dataAudit**（数据审核） | 只读模式 + 审核意见 | readonly |
| P4 | **reportPrepare**（报告编制） | 只读展示原始数据，自动拉取公式计算值填入报告 | readonly |

#### 3.2 dataEntry 改造前后对比

**改前**：
```
┌─ el-dialog: 新增录入 ─────────────────────┐
│ 委托ID    [____13____]                      │
│ 检测项目  [____5____]                       │
│ 原始数据  [{ "pH": 7.2 }]  ← textarea 手输  │
└────────────────────────────────────────────┘
```

**改后**：
```
┌─ el-dialog: 新增录入 ─────────────────────┐
│ 委托ID    [____13____]                      │
│ 检测项目  [COD 检测 ▼]                      │ ← 下拉选检测项目
│ 模板      [COD_V1 (自动加载)]               │ ← 自动匹配模板
│ ┌─ Univer Sheet 编辑器 ─────────────────┐  │
│ │ 样品编号 │ 取样量 │ 稀释倍数 │ COD计算值 │  │
│ │ S001     │ 20     │ 1       │ =B2*C2   │  │ ← 公式实时算
│ │ S002     │ 10     │ 2       │ =B3*C3   │  │
│ └────────────────────────────────────────┘  │
│ 💡 公式列自动锁定只读，防止误改              │
└────────────────────────────────────────────┘
```

#### 3.3 新增/改造文件

```
frontend/src/components/LabSheetEditor.vue       -- 新建（阶段 1）
frontend/src/views/baseData/labSheetTemplate/    -- 新建（模板管理页面）
frontend/src/views/business/dataEntry/index.vue  -- 改造
frontend/src/views/business/dataReview/index.vue -- 改造
frontend/src/views/business/dataAudit/index.vue  -- 改造
frontend/src/views/business/reportPrepare/index.vue -- 改造
frontend/src/api/business/labSheet.ts            -- 新建 API
frontend/src/api/base/labSheetTemplate.ts         -- 新建 API
```

---

### 阶段 4：流程引擎集成

**目标**：流程推进到 dataEntry 节点时，自动为该委托关联的检测项目创建空白检验单实例

改造 `BusinessService.ensureBusinessRecord`：

```go
case workflow.NodeDataEntry:
    // ① 创建 data_entry 记录（已有逻辑）
    tx.Create(&model.DataEntry{TaskOrderID: orderID, OriginalData: "{}"})

    // ② 解析 task_assign.test_item_list（用户分派了哪些检测项目）
    var assign model.TaskAssign
    tx.Where("task_order_id", orderID).First(&assign)
    var items []uint
    json.Unmarshal([]byte(assign.TestItemList), &items)

    // ③ 对每个检测项目，查 lab_sheet_templates 取最新版本
    // ④ 用模板 structure 初始化 lab_sheets.sheet_data
    // ⑤ 自动关联 template_id
    for _, itemID := range items {
        var tpl model.LabSheetTemplate
        tx.Where("test_item_id", itemID).Order("version DESC").First(&tpl)
        tx.Create(&model.LabSheet{
            TaskOrderID: orderID,
            TestItemID:  itemID,
            TemplateID:  tpl.ID,
            NodeCode:    workflow.NodeDataEntry,
            SheetData:   tpl.Structure,   // 模板结构就是初始数据
        })
    }
```

---

### 阶段 5：增强功能（可选，视时间而定）

| 功能 | 说明 | 优先级 |
|------|------|--------|
| **Excel 导入导出** | 从 Excel 文件读取数据到 Univer，或导出 Univer 为 Excel | P1 |
| **单元格批注** | dataReview 环节复核人可以在单元格加批注（如"稀释倍数需复核"） | P1 |
| **公式结果自动回填** | Univer 计算完公式后，自动把结果存到 `formula_results` jsonb，后续节点直接用数值不用重算 | P2 |
| **模板版本管理** | 模板支持版本迭代，旧的 lab_sheets 仍用旧模板，新流转的用新模板 | P2 |
| **打印/PDF 导出** | 检验单一键导出 PDF（报告归档需要） | P2 |
| **后端公式重算** | 后端用 Go formula parser 重算一遍公式结果，作为前端计算的校验 | P3 |

---

## 4. 数据流转全景

```
检测项目 (TestItem)
    │
    │ 1:N
    ▼
检验单模板 (LabSheetTemplate)
    │  structure: Univer snapshot（预定义行/列/公式）
    │
    ▼  (流程推进到 dataEntry 时自动创建)
检验单实例 (LabSheet)
    │  sheet_data: Univer snapshot（用户填完数据后的完整状态）
    │  formula_results: { "E2": 67.8 }（公式计算结果缓存）
    │
    ▼
各业务页面加载:
  - dataEntry:    加载模板 → 用户填数据 → 保存 sheet_data → edit 模式
  - dataReview:   加载已保存的 sheet_data → readonly 模式 + 批注
  - dataAudit:    加载已保存的 sheet_data → readonly 模式 + 审核意见
  - reportPrepare: 加载 formula_results 自动填入报告
```

---

## 5. 风险与应对

| 风险 | 影响 | 应对方案 |
|------|------|---------|
| Univer 体积较大 | 首次下载 ~500KB | Vite 代码分割 + 按需加载，内网环境无感 |
| Univer snapshot JSON 结构复杂 | 深嵌套结构存储和读取 | 直接存 PostgreSQL jsonb，不做二次解析；前端首次加载完整 snapshot，后续增量变更 |
| Univer 与 Element Plus CSS 冲突 | 样式错乱 | Univer 有自己的隔离容器，实际验证；冲突时用 `!important` 局部覆盖 |
| 公式计算结果同步 | 前端算完怎么传后端 | 方案 A（推荐）：前端 Univer 算完抛 event，存 snapshot + formula_results；方案 B：后端 Go formula parser 重算（阶段 5 可选） |
| 模板版本与历史数据兼容 | 模板更新后旧实例怎么办 | 每个 lab_sheet 记录 template_id，打开时用对应版本的模板渲染；模板更新产生新版本，旧实例不迁移 |

---

## 6. 预估工时

| 阶段 | 预估 | 主要工作 |
|------|------|---------|
| 1. Univer 组件封装 | 1 天 | 学习 Univer API + 封装 LabSheetEditor.vue |
| 2. 后端 Model + API | 1 天 | GORM model + CRUD handler + 路由注册 |
| 3. 前端页面改造 | 2 天 | 模板管理 + dataEntry + dataReview + dataAudit |
| 4. 流程引擎集成 | 0.5 天 | 改 ensureBusinessRecord 自动创建 lab_sheets |
| 5. 增强功能 | 1-2 天 | 批注、导入导出、公式回填 |
| **合计** | **5.5-6.5 天** | |

---

## 7. 验收标准

- [ ] `npm install` Univer 依赖无报错，vite dev 能正常启动
- [ ] LabSheetEditor.vue 在 demo 页面能正常渲染 Univer，支持编辑和只读两种模式
- [ ] 后端 `lab_sheet_templates` 和 `lab_sheets` 表 GORM AutoMigrate 成功
- [ ] 模板管理页面能所见即所得设计模板并保存
- [ ] dataEntry 页面能用 LabSheetEditor 填数据，保存后刷新能看到
- [ ] 公式单元格能实时计算（如 `=B2*C2`）
- [ ] 流程推进到 dataEntry 节点时，自动为该委托的每个检测项目创建空白检验单
- [ ] dataReview / dataAudit 页面能只读展示已填数据
- [ ] `go build ./...` 零错误，`npm run build` 零错误