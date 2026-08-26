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
        <el-table-column label="操作" width="200" fixed="right">
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

    <!-- Approve dialog -->
    <el-dialog v-model="approveVisible" title="合同评审 - 通过" width="500px">
      <el-form ref="approveFormRef" :model="approveForm" label-width="100px">
        <el-form-item label="委托ID">{{ approveForm.task_order_id }}</el-form-item>
        <el-form-item label="评审结果">
          <el-tag type="success" size="default">通过</el-tag>
        </el-form-item>
        <el-form-item label="评审意见">
          <el-input v-model="approveForm.comment" type="textarea" :rows="3" />
        </el-form-item>
        <el-form-item label="合同文件">
          <el-input v-model="approveForm.contract_file_path" placeholder="上传后路径" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="approveVisible = false">取消</el-button>
        <el-button type="primary" :loading="approving" @click="handleApprove">确认通过</el-button>
      </template>
    </el-dialog>

    <!-- Reject dialog -->
    <el-dialog v-model="rejectVisible" title="合同评审 - 驳回" width="500px">
      <el-form ref="rejectFormRef" :model="rejectForm" :rules="rejectRules" label-width="100px">
        <el-form-item label="委托ID">{{ rejectForm.task_order_id }}</el-form-item>
        <el-form-item label="驳回原因" prop="comment">
          <el-input v-model="rejectForm.comment" type="textarea" :rows="3" placeholder="请填写驳回原因" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="rejectVisible = false">取消</el-button>
        <el-button type="danger" :loading="rejecting" @click="handleReject">确认驳回</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
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
    items.value = res.data
  } finally { loading.value = false }
}

// --- Form dialog ---
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

// --- Approve ---
const approveVisible = ref(false)
const approveFormRef = ref<FormInstance>()
const approving = ref(false)
const approveForm = reactive({ task_order_id: 0, id: 0, comment: '', contract_file_path: '' })
function openApprove(row: ContractReview) {
  approveForm.id = row.id; approveForm.task_order_id = row.task_order_id; approveForm.comment = ''; approveForm.contract_file_path = row.contract_file_path
  approveVisible.value = true
}
async function handleApprove() {
  approving.value = true
  try {
    await approveContractReview(approveForm.id, { task_id: approveForm.task_order_id, review_comment: approveForm.comment, contract_file_path: approveForm.contract_file_path })
    ElMessage.success('评审通过，流程已推进')
    approveVisible.value = false; await loadData()
  } catch (e: any) { ElMessage.error(e?.response?.data?.message || '操作失败') }
  finally { approving.value = false }
}

// --- Reject ---
const rejectVisible = ref(false)
const rejectFormRef = ref<FormInstance>()
const rejecting = ref(false)
const rejectForm = reactive({ task_order_id: 0, id: 0, comment: '' })
const rejectRules: FormRules = { comment: [{ required: true, message: '请填写驳回原因', trigger: 'blur' }] }
function openReject(row: ContractReview) {
  rejectForm.id = row.id; rejectForm.task_order_id = row.task_order_id; rejectForm.comment = ''
  rejectVisible.value = true
}
async function handleReject() {
  const valid = await rejectFormRef.value?.validate().catch(() => false)
  if (!valid) return
  rejecting.value = true
  try {
    await rejectContractReview(rejectForm.id, { task_id: rejectForm.task_order_id, review_comment: rejectForm.comment })
    ElMessage.success('已驳回')
    rejectVisible.value = false; await loadData()
  } catch (e: any) { ElMessage.error(e?.response?.data?.message || '操作失败') }
  finally { rejecting.value = false }
}
</script>