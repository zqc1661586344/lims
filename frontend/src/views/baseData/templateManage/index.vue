<script setup lang="ts">
import { computed, nextTick, onMounted, ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import LabSheetEditor from '@/components/LabSheetEditor.vue'
import {
  listSamplingSheetTemplates,
  createSamplingSheetTemplate,
  updateSamplingSheetTemplate,
  deleteSamplingSheetTemplate,
  type SamplingSheetTemplate,
} from '@/api/business/samplingSheet'
import {
  listLabSheetTemplates,
  createLabSheetTemplate,
  updateLabSheetTemplate,
  deleteLabSheetTemplate,
  type LabSheetTemplate,
} from '@/api/business/labSheet'

type TabKey = 'sampling' | 'lab'

interface SamplingFormState {
  id?: number
  code: string
  name: string
  sample_type: string
  node_code: string
  editable_ranges: string
  readonly_ranges: string
  version: number
  status: number
  structure: any
}

interface LabFormState {
  id?: number
  code: string
  name: string
  test_item_id: number | null
  node_code: string
  editable_ranges: string
  readonly_ranges: string
  version: number
  status: number
  structure: any
}

const activeTab = ref<TabKey>('sampling')
const editorRef = ref<InstanceType<typeof LabSheetEditor>>()

const samplingTemplates = ref<SamplingSheetTemplate[]>([])
const labTemplates = ref<LabSheetTemplate[]>([])
const loadingList = ref(false)
const loadingEditor = ref(false)
const saving = ref(false)
const selectedId = ref<number | null>(null)
const initialized = ref(false)

const samplingForm = ref<SamplingFormState>({
  code: '', name: '', sample_type: '', node_code: '',
  editable_ranges: '', readonly_ranges: '', version: 1, status: 1, structure: null,
})

const labForm = ref<LabFormState>({
  code: '', name: '', test_item_id: null, node_code: '',
  editable_ranges: '', readonly_ranges: '', version: 1, status: 1, structure: null,
})

const currentForm = computed(() => activeTab.value === 'sampling' ? samplingForm.value : labForm.value)
const currentList = computed(() => activeTab.value === 'sampling' ? samplingTemplates.value : labTemplates.value)

function loadList() {
  loadingList.value = true
  const fn = activeTab.value === 'sampling'
    ? listSamplingSheetTemplates
    : listLabSheetTemplates
  fn({ status: '1' }).then((res: any) => {
    const list = res?.data ?? res ?? []
    if (activeTab.value === 'sampling') {
      samplingTemplates.value = list
    } else {
      labTemplates.value = list
    }
  }).finally(() => { loadingList.value = false })
}

watch(activeTab, () => {
  selectedId.value = null
  resetForm()
  initialized.value = false
  loadList()
})

function resetForm() {
  if (activeTab.value === 'sampling') {
    samplingForm.value = {
      code: '', name: '', sample_type: '', node_code: '',
      editable_ranges: '', readonly_ranges: '', version: 1, status: 1, structure: null,
    }
  } else {
    labForm.value = {
      code: '', name: '', test_item_id: null, node_code: '',
      editable_ranges: '', readonly_ranges: '', version: 1, status: 1, structure: null,
    }
  }
}

async function handleSelect(id: number) {
  if (selectedId.value === id) return
  selectedId.value = id
  const tpl = currentList.value.find(t => t.id === id)
  if (!tpl) return

  if (activeTab.value === 'sampling') {
    samplingForm.value = {
      id: tpl.id,
      code: tpl.code,
      name: tpl.name,
      sample_type: (tpl as any).sample_type,
      node_code: tpl.node_code,
      editable_ranges: (tpl.editable_ranges || []).join(', '),
      readonly_ranges: (tpl.readonly_ranges || []).join(', '),
      version: tpl.version,
      status: tpl.status,
      structure: tpl.structure,
    }
  } else {
    labForm.value = {
      id: tpl.id,
      code: tpl.code,
      name: tpl.name,
      test_item_id: (tpl as any).test_item_id ?? null,
      node_code: tpl.node_code,
      editable_ranges: (tpl.editable_ranges || []).join(', '),
      readonly_ranges: (tpl.readonly_ranges || []).join(', '),
      version: tpl.version,
      status: tpl.status,
      structure: tpl.structure,
    }
  }

  initialized.value = false
  loadingEditor.value = true
  await nextTick()
  setTimeout(() => {
    initialized.value = true
    loadingEditor.value = false
  }, 50)
}

function handleNew() {
  selectedId.value = null
  resetForm()
  initialized.value = false
  loadingEditor.value = true
  nextTick(() => {
    initialized.value = true
    loadingEditor.value = false
  })
  ElMessage.info('已创建新模板，在右侧填写元数据并设计表格结构后点击保存')
}

function parseRanges(str: string): string[] {
  return str.split(/[,，]/).map(s => s.trim()).filter(Boolean)
}

async function handleSave() {
  const f = currentForm.value
  if (!f.code || !f.name) {
    ElMessage.warning('请填写模板编码和名称')
    return
  }

  let snapshot: any = null
  try {
    snapshot = editorRef.value?.getSnapshot()
  } catch {}
  if (!snapshot) {
    ElMessage.error('无法获取表格结构，请稍候重试')
    return
  }

  const payload: any = {
    code: f.code,
    name: f.name,
    node_code: f.node_code,
    editable_ranges: parseRanges(f.editable_ranges),
    readonly_ranges: parseRanges(f.readonly_ranges),
    version: f.version,
    status: f.status,
    structure: snapshot,
  }

  if (activeTab.value === 'sampling') {
    payload.sample_type = (f as SamplingFormState).sample_type
  } else {
    payload.test_item_id = (f as LabFormState).test_item_id
  }

  saving.value = true
  try {
    if (f.id) {
      if (activeTab.value === 'sampling') {
        await updateSamplingSheetTemplate(f.id, payload)
      } else {
        await updateLabSheetTemplate(f.id, payload)
      }
      ElMessage.success('模板已更新')
    } else {
      if (activeTab.value === 'sampling') {
        const res: any = await createSamplingSheetTemplate(payload)
        f.id = res?.data?.id ?? res?.id
      } else {
        const res: any = await createLabSheetTemplate(payload)
        f.id = res?.data?.id ?? res?.id
      }
      if (f.id) selectedId.value = f.id
      ElMessage.success('模板已创建')
    }
    loadList()
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.message || '保存失败')
  } finally {
    saving.value = false
  }
}

async function handleDelete() {
  if (!currentForm.value.id) {
    ElMessage.warning('请先选择一个模板')
    return
  }
  try {
    await ElMessageBox.confirm(
      `确认删除模板"${currentForm.value.name}"吗？删除后不可恢复。`,
      '删除确认',
      { type: 'warning' }
    )
  } catch { return }

  try {
    if (activeTab.value === 'sampling') {
      await deleteSamplingSheetTemplate(currentForm.value.id)
    } else {
      await deleteLabSheetTemplate(currentForm.value.id)
    }
    ElMessage.success('已删除')
    selectedId.value = null
    resetForm()
    initialized.value = false
    loadList()
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.message || '删除失败')
  }
}

onMounted(() => { loadList() })
</script>

<template>
  <div class="template-manage">
    <el-tabs v-model="activeTab" class="top-tabs">
      <el-tab-pane label="采样单模板" name="sampling" />
      <el-tab-pane label="检验单模板" name="lab" />
    </el-tabs>

    <div class="content-area">
      <div class="list-panel">
        <div class="list-header">
          <span class="list-title">
            {{ activeTab === 'sampling' ? '采样单' : '检验单' }}模板列表
            <span class="count">{{ currentList.length }}</span>
          </span>
          <el-button type="primary" size="small" :icon="Plus" @click="handleNew">新建</el-button>
        </div>
        <div class="list-scroll">
          <el-scrollbar v-loading="loadingList">
            <div
              v-for="tpl in currentList"
              :key="tpl.id"
              class="list-item"
              :class="{ active: selectedId === tpl.id }"
              @click="handleSelect(tpl.id)"
            >
              <div class="item-name">{{ tpl.name }}</div>
              <div class="item-meta">
                <el-tag size="small" type="info">v{{ tpl.version }}</el-tag>
                <el-tag size="small" v-if="activeTab === 'sampling' && (tpl as any).sample_type">
                  {{ (tpl as any).sample_type }}
                </el-tag>
                <el-tag size="small" v-if="activeTab === 'lab' && (tpl as any).test_item">
                  {{ (tpl as any).test_item?.name || '未绑定项目' }}
                </el-tag>
              </div>
            </div>
            <el-empty v-if="!loadingList && currentList.length === 0" description="暂无模板" :image-size="80" />
          </el-scrollbar>
        </div>
      </div>

      <div class="editor-panel">
        <div class="meta-form">
          <el-form :model="currentForm" label-width="100px" size="default" inline>
            <el-form-item label="模板编码">
              <el-input v-model="currentForm.code" placeholder="如 surface_water_v1" style="width: 220px" />
            </el-form-item>
            <el-form-item label="模板名称">
              <el-input v-model="currentForm.name" placeholder="如 地表水现场采样记录" style="width: 260px" />
            </el-form-item>
            <el-form-item v-if="activeTab === 'sampling'" label="样品类型">
              <el-select v-model="(currentForm as SamplingFormState).sample_type" placeholder="请选择" style="width: 160px">
                <el-option label="地表水" value="地表水" />
                <el-option label="地下水" value="地下水" />
                <el-option label="土壤" value="土壤" />
                <el-option label="大气" value="大气" />
                <el-option label="噪声" value="噪声" />
                <el-option label="工业废水" value="工业废水" />
                <el-option label="生活污水" value="生活污水" />
              </el-select>
            </el-form-item>
            <el-form-item v-else label="检测项目ID">
              <el-input-number v-model="(currentForm as LabFormState).test_item_id" :min="1" placeholder="关联检测项目ID" style="width: 180px" />
            </el-form-item>
            <el-form-item label="绑定节点">
              <el-select v-model="currentForm.node_code" placeholder="请选择" clearable style="width: 200px">
                <el-option v-if="activeTab === 'sampling'" label="现场采样 (node_field_sampling)" value="node_field_sampling" />
                <el-option v-if="activeTab === 'lab'" label="数据录入 (node_data_entry)" value="node_data_entry" />
                <el-option label="数据复核 (node_data_review)" value="node_data_review" />
                <el-option label="数据审核 (node_data_audit)" value="node_data_audit" />
                <el-option label="报告编制 (node_report_prepare)" value="node_report_prepare" />
              </el-select>
            </el-form-item>
            <el-form-item label="可编辑区域">
              <el-input v-model="currentForm.editable_ranges" placeholder="A3:E20, G3:G20" style="width: 240px" />
            </el-form-item>
            <el-form-item label="只读区域">
              <el-input v-model="currentForm.readonly_ranges" placeholder="A1:E2（表头/公式列）" style="width: 240px" />
            </el-form-item>
            <el-form-item label="版本号">
              <el-input-number v-model="currentForm.version" :min="1" style="width: 110px" />
            </el-form-item>
            <el-form-item label="状态">
              <el-switch
                :model-value="currentForm.status === 1"
                @update:model-value="(v: boolean) => currentForm.status = v ? 1 : 0"
                active-text="启用"
                inactive-text="停用"
              />
            </el-form-item>
          </el-form>
        </div>

        <div class="sheet-editor-wrap">
          <div v-if="!initialized" class="sheet-placeholder">
            <el-icon class="placeholder-icon"><Document /></el-icon>
            <span>
              {{ selectedId ? '正在加载模板表格结构...' : '点击左侧列表选择模板，或点击"新建"开始设计新模板' }}
            </span>
          </div>
          <LabSheetEditor
            v-else
            ref="editorRef"
            mode="edit"
            :sheet-data="currentForm.structure || undefined"
            height="100%"
          />
        </div>

        <div class="action-bar">
          <span class="save-tip">
            <el-icon><InfoFilled /></el-icon>
            保存后模板将在"现场采样 / 数据录入"节点自动匹配加载
          </span>
          <div class="actions">
            <el-button @click="handleDelete" :disabled="!currentForm.id" type="danger" plain>删除模板</el-button>
            <el-button type="primary" :loading="saving" @click="handleSave">保存模板</el-button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { Plus, Document, InfoFilled } from '@element-plus/icons-vue'
export default {
  components: { Plus, Document, InfoFilled },
}
</script>

<style scoped>
.template-manage {
  height: 100%;
  display: flex;
  flex-direction: column;
  padding: 0;
  background: #f5f7fa;
}

.top-tabs {
  padding: 12px 16px 0;
  background: #fff;
  border-bottom: 1px solid #ebeef5;
  margin-bottom: 0;
}

.top-tabs :deep(.el-tabs__content) {
  display: none;
}

.content-area {
  flex: 1;
  display: flex;
  gap: 12px;
  padding: 12px;
  overflow: hidden;
}

.list-panel {
  width: 260px;
  flex-shrink: 0;
  background: #fff;
  border-radius: 6px;
  border: 1px solid #ebeef5;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.list-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 14px;
  border-bottom: 1px solid #ebeef5;
}

.list-title {
  font-size: 14px;
  font-weight: 600;
  color: #303133;
}

.count {
  display: inline-block;
  background: #e4e7ed;
  color: #606266;
  border-radius: 10px;
  font-size: 11px;
  padding: 1px 7px;
  margin-left: 6px;
}

.list-scroll {
  flex: 1;
  overflow: hidden;
}

.list-item {
  padding: 10px 14px;
  cursor: pointer;
  border-bottom: 1px solid #f2f6fc;
  transition: background 0.15s;
}

.list-item:hover {
  background: #f5f7fa;
}

.list-item.active {
  background: #ecf5ff;
  border-left: 3px solid #409eff;
}

.item-name {
  font-size: 13px;
  font-weight: 500;
  color: #303133;
  margin-bottom: 5px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.item-meta {
  display: flex;
  gap: 4px;
  flex-wrap: wrap;
}

.editor-panel {
  flex: 1;
  background: #fff;
  border-radius: 6px;
  border: 1px solid #ebeef5;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.meta-form {
  padding: 12px 16px;
  border-bottom: 1px solid #ebeef5;
  background: #fafbfc;
}

.meta-form :deep(.el-form-item) {
  margin-bottom: 6px;
  margin-right: 16px;
}

.meta-form :deep(.el-form-item:last-child) {
  margin-bottom: 0;
}

.meta-form :deep(.el-form-item__label) {
  font-size: 13px;
  color: #606266;
}

.sheet-editor-wrap {
  flex: 1;
  min-height: 0;
  position: relative;
  background: #fff;
}

.sheet-placeholder {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 14px;
  color: #909399;
  font-size: 14px;
}

.placeholder-icon {
  font-size: 48px;
  color: #c0c4cc;
}

.action-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 16px;
  border-top: 1px solid #ebeef5;
  background: #fafbfc;
}

.save-tip {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: #909399;
}

.actions {
  display: flex;
  gap: 10px;
}
</style>