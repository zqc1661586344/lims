<template>
  <div class="contract-review-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>合同评审管理</span>
          <div>
            <el-input v-model="taskOrderIdFilter" placeholder="委托单号/ID" clearable style="width:200px;margin-right:8px" @clear="loadData" @keyup.enter="loadData" />
            <el-button type="primary" @click="openCreate">新增评审</el-button>
          </div>
        </div>
      </template>

      <el-table :data="items" stripe v-loading="loading">
        <el-table-column prop="task_order_id" label="委托ID" width="80" />
        <el-table-column prop="review_result" label="评审结果" width="100">
          <template #default="{ row }">
            <el-tag :type="row.review_result === '通过' ? 'success' : row.review_result === '驳回' ? 'danger' : 'info'" size="small">
              {{ row.review_result || '待评审' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="review_comment" label="评审意见" min-width="200" show-overflow-tooltip />
        <el-table-column prop="contract_file_path" label="合同文件" min-width="160" show-overflow-tooltip />
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
    <el-dialog v-model="formVisible" :title="isEdit ? '编辑合同评审' : '新增合同评审'" width="600px">
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="100px">
        <el-form-item label="委托ID" prop="task_order_id">
          <el-input-number v-model="form.task_order_id" :min="1" style="width:100%" />
        </el-form-item>
        <el-form-item label="评审意见">
          <el-input v-model="form.review_comment" type="textarea" :rows="3" />
        </el-form-item>
        <el-form-item label="合同文件路径">
          <el-input v-model="form.contract_file_path" placeholder="上传后路径" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="formVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSave">保存</el-button>
      </template>
    </el-dialog>

    <ApprovalDialog ref="approvalRef" title="合同评审审批" width="500px" @submit="handleApprovalSubmit">
      <template #header>
        <div class="approval-header">
          <span>委托ID：{{ currentRow?.task_order_id }}</span>
        </div>
      </template>
      <template #extraFields>
        <el-form-item label="合同文件" v-if="approvalForm.action === 'approve'">
          <el-input v-model="approveExtra.contract_file_path" placeholder="上传后路径" />
        </el-form-item>
      </template>
    </ApprovalDialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import ApprovalDialog from '@/components/ApprovalDialog.vue'
import { getContractReviewList, getContractReview, createContractReview, updateContractReview, deleteContractReview, approveContractReview, rejectContractReview } from '@/api/business'
import type { ContractReview } from '@/api/business'

const loading = ref(false)
const items = ref<ContractReview[]>([])
const taskOrderIdFilter = ref('')

onMounted(async () => { await loadData() })

async function loadData() {
  loading.value = true
  try {
    const params: Record<string, string> = {}
    if (taskOrderIdFilter.value) params.task_order_id = taskOrderIdFilter.value
    const res = await getContractReviewList(params)
    items.value = res.data?.items ?? res.data
  } finally { loading.value = false }
}

const formVisible = ref(false)
const isEdit = ref(false)
const formRef = ref<FormInstance>()
const saving = ref(false)
const form = reactive({ id: 0, task_order_id: 0, review_comment: '', contract_file_path: '' })
const formRules: FormRules = {
  task_order_id: [{ required: true, message: '请输入委托ID', trigger: 'blur' }],
}
function openCreate() {
  isEdit.value = false
  form.id = 0; form.task_order_id = 0; form.review_comment = ''; form.contract_file_path = ''
  formVisible.value = true
}
async function openEdit(row: ContractReview) {
  isEdit.value = true
  const res = await getContractReview(row.id)
  const d = res.data
  form.id = d.id; form.task_order_id = d.task_order_id; form.review_comment = d.review_comment; form.contract_file_path = d.contract_file_path
  formVisible.value = true
}
async function handleSave() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  saving.value = true
  try {
    const data = { task_order_id: form.task_order_id, review_comment: form.review_comment, contract_file_path: form.contract_file_path }
    if (isEdit.value) { await updateContractReview(form.id, data); ElMessage.success('更新成功') }
    else { await createContractReview(data); ElMessage.success('创建成功') }
    formVisible.value = false; await loadData()
  } finally { saving.value = false }
}
async function handleDelete(id: number) {
  await deleteContractReview(id); ElMessage.success('删除成功'); await loadData()
}

const approvalRef = ref<InstanceType<typeof ApprovalDialog>>()
const currentRow = ref<ContractReview | null>(null)
const approvalForm = reactive({ action: 'approve' as 'approve' | 'reject' })
const approveExtra = reactive({ contract_file_path: '' })

function openApprove(row: ContractReview) {
  currentRow.value = row
  approveExtra.contract_file_path = row.contract_file_path
  approvalForm.action = 'approve'
  approvalRef.value?.open('approve')
}
function openReject(row: ContractReview) {
  currentRow.value = row
  approvalForm.action = 'reject'
  approvalRef.value?.open('reject')
}

async function handleApprovalSubmit(data: { action: string; comment: string }) {
  if (!currentRow.value) return
  const id = currentRow.value.id
  const taskId = currentRow.value.task_order_id
  if (data.action === 'approve') {
    await approveContractReview(id, { task_id: taskId, review_comment: data.comment, contract_file_path: approveExtra.contract_file_path })
    ElMessage.success('评审通过，流程已推进')
  } else {
    await rejectContractReview(id, { task_id: taskId, review_comment: data.comment })
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