<template>
  <div class="qc-task-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>质控任务管理</span>
          <div>
            <el-input v-model="taskOrderIdFilter" placeholder="委托ID" clearable style="width:160px;margin-right:8px" @clear="loadData" @keyup.enter="loadData" />
            <el-button type="primary" @click="openCreate">新增质控任务</el-button>
          </div>
        </div>
      </template>

      <el-table :data="items" stripe v-loading="loading">
        <el-table-column prop="task_order_id" label="委托ID" width="80" />
        <el-table-column prop="qc_type" label="质控类型" width="120" />
        <el-table-column prop="qc_details" label="质控详情" min-width="200" show-overflow-tooltip />
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
    <el-dialog v-model="formVisible" :title="isEdit ? '编辑质控任务' : '新增质控任务'" width="550px">
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="100px">
        <el-form-item label="委托ID" prop="task_order_id">
          <el-input-number v-model="form.task_order_id" :min="1" style="width:100%" />
        </el-form-item>
        <el-form-item label="质控类型">
          <el-select v-model="form.qc_type" placeholder="选择类型" style="width:100%">
            <el-option label="空白样" value="blank" />
            <el-option label="平行样" value="duplicate" />
            <el-option label="标准样" value="standard" />
            <el-option label="加标回收" value="spike" />
          </el-select>
        </el-form-item>
        <el-form-item label="质控详情">
          <el-input v-model="form.qc_details" type="textarea" :rows="3" placeholder="JSON格式详情" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="formVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSave">保存</el-button>
      </template>
    </el-dialog>

    <ApprovalDialog ref="approvalRef" title="质控任务审批" @submit="handleApprovalSubmit">
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
import { getQCTaskList, getQCTask, createQCTask, updateQCTask, deleteQCTask, approveQCTask, rejectQCTask } from '@/api/business'
import type { QCTask } from '@/api/business'

const loading = ref(false)
const items = ref<QCTask[]>([])
const taskOrderIdFilter = ref('')

onMounted(async () => { await loadData() })

async function loadData() {
  loading.value = true
  try {
    const params: Record<string, string> = {}
    if (taskOrderIdFilter.value) params.task_order_id = taskOrderIdFilter.value
    const res = await getQCTaskList(params)
    items.value = res.data?.items ?? res.data
  } finally { loading.value = false }
}

const formVisible = ref(false)
const isEdit = ref(false)
const formRef = ref<FormInstance>()
const saving = ref(false)
const form = reactive({ id: 0, task_order_id: 0, qc_type: '', qc_details: '' })
const formRules: FormRules = { task_order_id: [{ required: true, message: '请输入委托ID', trigger: 'blur' }] }
function openCreate() {
  isEdit.value = false; form.id = 0; form.task_order_id = 0; form.qc_type = ''; form.qc_details = ''
  formVisible.value = true
}
async function openEdit(row: QCTask) {
  isEdit.value = true
  const res = await getQCTask(row.id)
  const d = res.data; form.id = d.id; form.task_order_id = d.task_order_id; form.qc_type = d.qc_type; form.qc_details = d.qc_details
  formVisible.value = true
}
async function handleSave() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  saving.value = true
  try {
    const data = { task_order_id: form.task_order_id, qc_type: form.qc_type, qc_details: form.qc_details }
    if (isEdit.value) { await updateQCTask(form.id, data); ElMessage.success('更新成功') }
    else { await createQCTask(data); ElMessage.success('创建成功') }
    formVisible.value = false; await loadData()
  } finally { saving.value = false }
}
async function handleDelete(id: number) {
  await deleteQCTask(id); ElMessage.success('删除成功'); await loadData()
}

const approvalRef = ref<InstanceType<typeof ApprovalDialog>>()
const currentRow = ref<QCTask | null>(null)

function openApprove(row: QCTask) { currentRow.value = row; approvalRef.value?.open('approve') }
function openReject(row: QCTask) { currentRow.value = row; approvalRef.value?.open('reject') }

async function handleApprovalSubmit(data: { action: string; comment: string }) {
  if (!currentRow.value) return
  const id = currentRow.value.id
  const taskId = currentRow.value.task_order_id
  if (data.action === 'approve') {
    await approveQCTask(id, { task_id: taskId, comment: data.comment })
    ElMessage.success('质控任务通过，流程已推进')
  } else {
    await rejectQCTask(id, { task_id: taskId, comment: data.comment })
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