<template>
  <div class="data-entry-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>数据录入管理</span>
          <div>
            <el-input v-model="taskOrderIdFilter" placeholder="委托ID" clearable style="width:160px;margin-right:8px" @clear="loadData" @keyup.enter="loadData" />
            <el-button type="primary" @click="openCreate">新增录入</el-button>
          </div>
        </div>
      </template>
      <el-table :data="items" stripe v-loading="loading">
        <el-table-column prop="task_order_id" label="委托ID" width="80" />
        <el-table-column prop="test_item_id" label="项目ID" width="80" />
        <el-table-column label="检验单" min-width="200">
          <template #default="{ row }">
            <el-button size="small" link type="primary" @click="openSheetEditor(row)">编辑检验单</el-button>
            <el-tag v-if="row.lab_sheet_id" type="success" size="small" style="margin-left:6px">已关联 #{{ row.lab_sheet_id }}</el-tag>
            <el-tag v-else type="info" size="small" style="margin-left:6px">未关联</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="raw_record_id" label="原始记录ID" width="110" />
        <el-table-column prop="created_at" label="创建时间" width="170" />
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-button size="small" type="success" @click="openApprove(row)">通过</el-button>
            <el-button size="small" type="danger" @click="openReject(row)">驳回</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 新增录入 Dialog -->
    <el-dialog v-model="formVisible" :title="'新增数据录入'" width="500px">
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="120px">
        <el-form-item label="委托ID" prop="task_order_id">
          <el-input-number v-model="form.task_order_id" :min="1" style="width:100%" />
        </el-form-item>
        <el-form-item label="检测项目ID" prop="test_item_id">
          <el-input-number v-model="form.test_item_id" :min="1" style="width:100%" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="formVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSave">保存并填写检验单</el-button>
      </template>
    </el-dialog>

    <ApprovalDialog ref="approvalRef" title="数据录入审批" @submit="handleApprovalSubmit">
      <template #header>
        <div class="approval-header">
          <span>委托ID：{{ currentRow?.task_order_id }}</span>
          <el-tag v-if="currentRow?.lab_sheet_id" type="success" style="margin-left:8px">已关联检验单</el-tag>
        </div>
      </template>
    </ApprovalDialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import ApprovalDialog from '@/components/ApprovalDialog.vue'
import {
  getDataEntryList,
  createDataEntry,
  approveDataEntry,
  rejectDataEntry,
  listLabSheets,
} from '@/api/business'
import type { DataEntry, LabSheet } from '@/api/business'

const router = useRouter()

interface DataEntryRow extends DataEntry {
  lab_sheet_id?: number | null
}

const loading = ref(false)
const items = ref<DataEntryRow[]>([])
const taskOrderIdFilter = ref('')

onMounted(async () => { await loadData() })

async function loadData() {
  loading.value = true
  try {
    const params: Record<string, string> = {}
    if (taskOrderIdFilter.value) params.task_order_id = taskOrderIdFilter.value
    const res = await getDataEntryList(params)
    const list: DataEntryRow[] = res.data || []

    const labSheetRes = await listLabSheets({ node_code: 'node_data_entry' })
    const labSheets: LabSheet[] = labSheetRes.data || []
    const sheetMap = new Map<number, LabSheet>()
    for (const ls of labSheets) sheetMap.set(ls.task_order_id, ls)

    for (const row of list) {
      const sheet = sheetMap.get(row.task_order_id)
      row.lab_sheet_id = sheet?.id ?? null
    }

    items.value = list
  } finally { loading.value = false }
}

// ─── 新增录入 ─────────────────────────────────────────────────

const formVisible = ref(false)
const formRef = ref<FormInstance>()
const saving = ref(false)
const form = reactive({ task_order_id: 0, test_item_id: 0 })
const formRules: FormRules = {
  task_order_id: [{ required: true, message: '请输入委托ID', trigger: 'blur' }],
  test_item_id: [{ required: true, message: '请输入检测项目ID', trigger: 'blur' }],
}

function openCreate() {
  form.task_order_id = 0
  form.test_item_id = 0
  formVisible.value = true
}

async function handleSave() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  saving.value = true
  try {
    await createDataEntry({
      task_order_id: form.task_order_id,
      test_item_id: form.test_item_id,
    })
    ElMessage.success('创建成功，请填写检验单')
    formVisible.value = false
    openSheetEditor({ task_order_id: form.task_order_id, test_item_id: form.test_item_id } as DataEntryRow)
  } finally { saving.value = false }
}

// ─── 检验单编辑（路由跳转）────────────────────────────────────

function openSheetEditor(row: DataEntryRow) {
  router.push({
    name: 'LabSheetEditorPage',
    query: {
      task_order_id: row.task_order_id,
      test_item_id: row.test_item_id,
      node_code: 'node_data_entry',
    },
  })
}

// ─── 审批 ─────────────────────────────────────────────────────

const approvalRef = ref<InstanceType<typeof ApprovalDialog>>()
const currentRow = ref<DataEntryRow | null>(null)

function openApprove(row: DataEntryRow) {
  currentRow.value = row
  approvalRef.value?.open('approve')
}
function openReject(row: DataEntryRow) {
  currentRow.value = row
  approvalRef.value?.open('reject')
}

async function handleApprovalSubmit(data: { action: string; comment: string }) {
  if (!currentRow.value) return
  const id = currentRow.value.id
  const taskId = currentRow.value.task_order_id
  if (data.action === 'approve') {
    await approveDataEntry(id, { task_id: taskId, comment: data.comment })
    ElMessage.success('录入通过，流程已推进')
  } else {
    await rejectDataEntry(id, { task_id: taskId, comment: data.comment })
    ElMessage.success('已驳回')
  }
  await loadData()
}
</script>

<style scoped>
.approval-header {
  margin-bottom: 12px;
  padding: 8px 12px;
  background: #f5f7fa;
  border-radius: 4px;
  font-size: 13px;
  color: #606266;
}
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>