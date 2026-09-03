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
        <el-table-column prop="original_data" label="原始数据" min-width="220" show-overflow-tooltip />
        <el-table-column prop="raw_record_id" label="原始记录ID" width="110" />
        <el-table-column prop="created_at" label="创建时间" width="170" />
        <el-table-column label="操作" width="240" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="openEdit(row)">编辑</el-button>
            <el-button size="small" type="success" @click="openApprove(row)">通过</el-button>
            <el-button size="small" type="danger" @click="openReject(row)">驳回</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- Form dialog -->
    <el-dialog v-model="formVisible" :title="isEdit ? '编辑数据录入' : '新增数据录入'" width="600px">
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="120px">
        <el-form-item label="委托ID" prop="task_order_id">
          <el-input-number v-model="form.task_order_id" :min="1" style="width:100%" />
        </el-form-item>
        <el-form-item label="检测项目ID">
          <el-input-number v-model="form.test_item_id" :min="0" style="width:100%" />
        </el-form-item>
        <el-form-item label="原始数据">
          <el-input v-model="form.original_data" type="textarea" :rows="4" placeholder="JSON: 原始检测数据" />
        </el-form-item>
        <el-form-item label="原始记录ID">
          <el-input-number v-model="form.raw_record_id" :min="0" style="width:100%" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="formVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSave">保存</el-button>
      </template>
    </el-dialog>

    <ApprovalDialog ref="approvalRef" title="数据录入审批" @submit="handleApprovalSubmit">
      <template #header>
        <div class="approval-header">
          <span>委托ID：{{ currentRow?.task_order_id }}</span>
        </div>
      </template>
    </ApprovalDialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import ApprovalDialog from '@/components/ApprovalDialog.vue'
import { getDataEntryList, getDataEntry, createDataEntry, updateDataEntry, deleteDataEntry, approveDataEntry, rejectDataEntry } from '@/api/business'
import type { DataEntry } from '@/api/business'

const loading = ref(false)
const items = ref<DataEntry[]>([])
const taskOrderIdFilter = ref('')

onMounted(async () => { await loadData() })

async function loadData() {
  loading.value = true
  try {
    const params: Record<string, string> = {}
    if (taskOrderIdFilter.value) params.task_order_id = taskOrderIdFilter.value
    const res = await getDataEntryList(params)
    items.value = res.data
  } finally { loading.value = false }
}

const formVisible = ref(false); const isEdit = ref(false); const formRef = ref<FormInstance>(); const saving = ref(false)
const form = reactive({ id: 0, task_order_id: 0, test_item_id: 0, original_data: '', raw_record_id: 0 })
const formRules: FormRules = { task_order_id: [{ required: true, message: '请输入委托ID', trigger: 'blur' }] }
function openCreate() { isEdit.value = false; form.id = 0; form.task_order_id = 0; form.test_item_id = 0; form.original_data = ''; form.raw_record_id = 0; formVisible.value = true }
async function openEdit(row: DataEntry) { isEdit.value = true; const res = await getDataEntry(row.id); const d = res.data; form.id = d.id; form.task_order_id = d.task_order_id; form.test_item_id = d.test_item_id; form.original_data = d.original_data; form.raw_record_id = d.raw_record_id ?? 0; formVisible.value = true }
async function handleSave() {
  const valid = await formRef.value?.validate().catch(() => false); if (!valid) return
  saving.value = true
  try {
    const data = { task_order_id: form.task_order_id, test_item_id: form.test_item_id, original_data: form.original_data, raw_record_id: form.raw_record_id || null }
    if (isEdit.value) { await updateDataEntry(form.id, data); ElMessage.success('更新成功') }
    else { await createDataEntry(data); ElMessage.success('创建成功') }
    formVisible.value = false; await loadData()
  } finally { saving.value = false }
}
async function handleDelete(id: number) { await deleteDataEntry(id); ElMessage.success('删除成功'); await loadData() }

const approvalRef = ref<InstanceType<typeof ApprovalDialog>>()
const currentRow = ref<DataEntry | null>(null)

function openApprove(row: DataEntry) { currentRow.value = row; approvalRef.value?.open('approve') }
function openReject(row: DataEntry) { currentRow.value = row; approvalRef.value?.open('reject') }

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
</style>