<template>
  <div class="sample-receiving-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>样品接收管理</span>
          <div>
            <el-input v-model="taskOrderIdFilter" placeholder="委托ID" clearable style="width:160px;margin-right:8px" @clear="loadData" @keyup.enter="loadData" />
            <el-button type="primary" @click="openCreate">新增接收记录</el-button>
          </div>
        </div>
      </template>

      <TaskProcessBar
        :business-type="'task_order'"
        :selected-id="selectedRow?.task_order_id ?? null"
        :items="items"
        :sticky="true"
      />

      <el-table :data="items" stripe v-loading="loading"
          highlight-current-row
          @row-click="handleRowClick">
        <el-table-column prop="task_order_id" label="委托ID" width="80" />
        <el-table-column prop="sample_condition" label="样品状态" width="120">
          <template #default="{ row }">
            <el-tag :type="row.sample_condition === '完好' ? 'success' : row.sample_condition === '损坏' ? 'danger' : 'warning'" size="small">
              {{ row.sample_condition || '待确认' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="sample_codes" label="样品编码" min-width="200" show-overflow-tooltip />
        <el-table-column prop="receiving_record_path" label="接收记录文件" min-width="180" show-overflow-tooltip />
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
    <el-dialog v-model="formVisible" :title="isEdit ? '编辑样品接收' : '新增样品接收'" width="600px">
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="120px">
        <el-form-item label="委托ID" prop="task_order_id">
          <el-input-number v-model="form.task_order_id" :min="1" style="width:100%" />
        </el-form-item>
        <el-form-item label="样品状态">
          <el-select v-model="form.sample_condition" placeholder="选择状态" style="width:100%">
            <el-option label="完好" value="完好" />
            <el-option label="损坏" value="损坏" />
            <el-option label="异常" value="异常" />
            <el-option label="部分损坏" value="部分损坏" />
          </el-select>
        </el-form-item>
        <el-form-item label="样品编码">
          <el-input v-model="form.sample_codes" type="textarea" :rows="2" placeholder="JSON: [{code,name,status}]" />
        </el-form-item>
        <el-form-item label="接收记录文件">
          <el-input v-model="form.receiving_record_path" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="formVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSave">保存</el-button>
      </template>
    </el-dialog>

    <ApprovalDialog ref="approvalRef" title="样品接收审批" @submit="handleApprovalSubmit">
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
import TaskProcessBar from '@/components/TaskProcessBar.vue'
import { getSampleReceivingList, getSampleReceiving, createSampleReceiving, updateSampleReceiving, deleteSampleReceiving, approveSampleReceiving, rejectSampleReceiving } from '@/api/business'
import type { SampleReceiving } from '@/api/business'

const loading = ref(false)
const items = ref<SampleReceiving[]>([])
const selectedRow = ref(null)
function handleRowClick(row: any) { selectedRow.value = row }
const taskOrderIdFilter = ref('')

onMounted(async () => { await loadData() })

async function loadData() {
  loading.value = true
  try {
    const params: Record<string, string> = {}
    if (taskOrderIdFilter.value) params.task_order_id = taskOrderIdFilter.value
    const res = await getSampleReceivingList(params)
    items.value = res.data?.items ?? res.data
  } finally { loading.value = false }
}

const formVisible = ref(false); const isEdit = ref(false); const formRef = ref<FormInstance>(); const saving = ref(false)
const form = reactive({ id: 0, task_order_id: 0, sample_condition: '', sample_codes: '', receiving_record_path: '' })
const formRules: FormRules = { task_order_id: [{ required: true, message: '请输入委托ID', trigger: 'blur' }] }
function openCreate() { isEdit.value = false; form.id = 0; form.task_order_id = 0; form.sample_condition = ''; form.sample_codes = ''; form.receiving_record_path = ''; formVisible.value = true }
async function openEdit(row: SampleReceiving) { isEdit.value = true; const res = await getSampleReceiving(row.id); const d = res.data; form.id = d.id; form.task_order_id = d.task_order_id; form.sample_condition = d.sample_condition; form.sample_codes = d.sample_codes; form.receiving_record_path = d.receiving_record_path; formVisible.value = true }
async function handleSave() {
  const valid = await formRef.value?.validate().catch(() => false); if (!valid) return
  saving.value = true
  try {
    const data = { task_order_id: form.task_order_id, sample_condition: form.sample_condition, sample_codes: form.sample_codes, receiving_record_path: form.receiving_record_path }
    if (isEdit.value) { await updateSampleReceiving(form.id, data); ElMessage.success('更新成功') }
    else { await createSampleReceiving(data); ElMessage.success('创建成功') }
    formVisible.value = false; await loadData()
  } finally { saving.value = false }
}
async function handleDelete(id: number) { await deleteSampleReceiving(id); ElMessage.success('删除成功'); await loadData() }

const approvalRef = ref<InstanceType<typeof ApprovalDialog>>()
const currentRow = ref<SampleReceiving | null>(null)

function openApprove(row: SampleReceiving) { currentRow.value = row; approvalRef.value?.open('approve') }
function openReject(row: SampleReceiving) { currentRow.value = row; approvalRef.value?.open('reject') }

async function handleApprovalSubmit(data: { action: string; comment: string }) {
  if (!currentRow.value) return
  const id = currentRow.value.id
  const taskId = currentRow.value.task_order_id
  if (data.action === 'approve') {
    await approveSampleReceiving(id, { task_id: taskId, comment: data.comment })
    ElMessage.success('接收通过，流程已推进')
  } else {
    await rejectSampleReceiving(id, { task_id: taskId, comment: data.comment })
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