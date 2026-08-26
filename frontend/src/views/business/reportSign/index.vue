<template>
  <div class="report-sign-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>报告签发</span>
          <div>
            <el-input v-model="taskOrderIdFilter" placeholder="委托ID" clearable style="width:160px;margin-right:8px" @clear="loadData" @keyup.enter="loadData" />
            <el-button type="primary" @click="openCreate">新增签发</el-button>
          </div>
        </div>
      </template>

      <el-table :data="items" stripe v-loading="loading">
        <el-table-column prop="task_order_id" label="委托ID" width="80" />
        <el-table-column prop="sign_result" label="签发结果" width="100">
          <template #default="{ row }">
            <el-tag :type="row.sign_result === '通过' ? 'success' : 'danger'" size="small">{{ row.sign_result }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="signer_name" label="签发人" width="120" />
        <el-table-column prop="sign_date" label="签发日期" width="170" />
        <el-table-column prop="sign_comment" label="签发意见" min-width="180" show-overflow-tooltip />
        <el-table-column prop="sign_stamp" label="印章文件" width="150" show-overflow-tooltip />
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
    <el-dialog v-model="formVisible" :title="isEdit ? '编辑报告签发' : '新增报告签发'" width="600px">
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="120px">
        <el-form-item label="委托ID" prop="task_order_id">
          <el-input-number v-model="form.task_order_id" :min="1" style="width:100%" />
        </el-form-item>
        <el-form-item label="签发结果">
          <el-select v-model="form.sign_result" placeholder="选择结果" style="width:100%">
            <el-option label="通过" value="通过" />
            <el-option label="驳回" value="驳回" />
          </el-select>
        </el-form-item>
        <el-form-item label="签发人">
          <el-input v-model="form.signer_name" placeholder="签发人姓名" />
        </el-form-item>
        <el-form-item label="签发日期">
          <el-date-picker v-model="form.sign_date" type="date" placeholder="选择日期" style="width:100%" />
        </el-form-item>
        <el-form-item label="印章文件">
          <el-input v-model="form.sign_stamp" placeholder="印章文件路径" />
        </el-form-item>
        <el-form-item label="签发意见">
          <el-input v-model="form.sign_comment" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="formVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSave">保存</el-button>
      </template>
    </el-dialog>

    <!-- Approve dialog -->
    <el-dialog v-model="approveVisible" title="报告签发 - 通过" width="500px">
      <el-form :model="approveForm" label-width="100px">
        <el-form-item label="委托ID">{{ approveForm.task_order_id }}</el-form-item>
        <el-form-item label="签发结果">
          <el-select v-model="approveForm.sign_result" placeholder="选择结果" style="width:100%">
            <el-option label="通过" value="通过" />
            <el-option label="驳回" value="驳回" />
          </el-select>
        </el-form-item>
        <el-form-item label="签发人">
          <el-input v-model="approveForm.signer_name" placeholder="签发人姓名" />
        </el-form-item>
        <el-form-item label="签发日期">
          <el-date-picker v-model="approveForm.sign_date" type="date" placeholder="选择日期" style="width:100%" />
        </el-form-item>
        <el-form-item label="印章文件">
          <el-input v-model="approveForm.sign_stamp" placeholder="印章文件路径" />
        </el-form-item>
        <el-form-item label="签发意见">
          <el-input v-model="approveForm.comment" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="approveVisible = false">取消</el-button>
        <el-button type="primary" :loading="approving" @click="handleApprove">确认通过</el-button>
      </template>
    </el-dialog>

    <!-- Reject dialog -->
    <el-dialog v-model="rejectVisible" title="报告签发 - 驳回" width="500px">
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
import { getReportSignList, getReportSign, createReportSign, updateReportSign, deleteReportSign, approveReportSign, rejectReportSign } from '@/api/business'
import type { ReportSign } from '@/api/business'

const loading = ref(false)
const items = ref<ReportSign[]>([])
const taskOrderIdFilter = ref('')

onMounted(async () => { await loadData() })

async function loadData() {
  loading.value = true
  try {
    const params: Record<string, string> = {}
    if (taskOrderIdFilter.value) params.task_order_id = taskOrderIdFilter.value
    const res = await getReportSignList(params)
    items.value = res.data
  } finally { loading.value = false }
}

const formVisible = ref(false); const isEdit = ref(false); const formRef = ref<FormInstance>(); const saving = ref(false)
const form = reactive({ id: 0, task_order_id: 0, sign_result: '', sign_comment: '', signer_name: '', sign_date: '', sign_stamp: '' })
const formRules: FormRules = { task_order_id: [{ required: true, message: '请输入委托ID', trigger: 'blur' }] }
function openCreate() { isEdit.value = false; form.id = 0; form.task_order_id = 0; form.sign_result = ''; form.sign_comment = ''; form.signer_name = ''; form.sign_date = ''; form.sign_stamp = ''; formVisible.value = true }
async function openEdit(row: ReportSign) { isEdit.value = true; const res = await getReportSign(row.id); const d = res.data; form.id = d.id; form.task_order_id = d.task_order_id; form.sign_result = d.sign_result; form.sign_comment = d.sign_comment; form.signer_name = d.signer_name; form.sign_date = d.sign_date; form.sign_stamp = d.sign_stamp; formVisible.value = true }
async function handleSave() {
  const valid = await formRef.value?.validate().catch(() => false); if (!valid) return
  saving.value = true
  try {
    const data = { task_order_id: form.task_order_id, sign_result: form.sign_result, sign_comment: form.sign_comment, signer_name: form.signer_name, sign_date: form.sign_date, sign_stamp: form.sign_stamp }
    if (isEdit.value) { await updateReportSign(form.id, data); ElMessage.success('更新成功') }
    else { await createReportSign(data); ElMessage.success('创建成功') }
    formVisible.value = false; await loadData()
  } finally { saving.value = false }
}
async function handleDelete(id: number) { await deleteReportSign(id); ElMessage.success('删除成功'); await loadData() }

const approveVisible = ref(false); const approving = ref(false)
const approveForm = reactive({ id: 0, task_order_id: 0, sign_result: '通过', comment: '', signer_name: '', sign_date: '', sign_stamp: '' })
function openApprove(row: ReportSign) { approveForm.id = row.id; approveForm.task_order_id = row.task_order_id; approveForm.sign_result = '通过'; approveForm.comment = ''; approveForm.signer_name = row.signer_name; approveForm.sign_date = row.sign_date; approveForm.sign_stamp = row.sign_stamp; approveVisible.value = true }
async function handleApprove() {
  approving.value = true
  try { await approveReportSign(approveForm.id, { task_id: approveForm.task_order_id, sign_result: approveForm.sign_result, sign_comment: approveForm.comment, signer_name: approveForm.signer_name, sign_date: approveForm.sign_date, sign_stamp: approveForm.sign_stamp }); ElMessage.success('签发通过，流程已推进'); approveVisible.value = false; await loadData() }
  catch (e: any) { ElMessage.error(e?.response?.data?.message || '操作失败') }
  finally { approving.value = false }
}

const rejectVisible = ref(false); const rejectFormRef = ref<FormInstance>(); const rejecting = ref(false)
const rejectForm = reactive({ id: 0, task_order_id: 0, comment: '' })
const rejectRules: FormRules = { comment: [{ required: true, message: '请填写驳回原因', trigger: 'blur' }] }
function openReject(row: ReportSign) { rejectForm.id = row.id; rejectForm.task_order_id = row.task_order_id; rejectForm.comment = ''; rejectVisible.value = true }
async function handleReject() {
  const valid = await rejectFormRef.value?.validate().catch(() => false); if (!valid) return
  rejecting.value = true
  try { await rejectReportSign(rejectForm.id, { task_id: rejectForm.task_order_id, comment: rejectForm.comment }); ElMessage.success('已驳回'); rejectVisible.value = false; await loadData() }
  catch (e: any) { ElMessage.error(e?.response?.data?.message || '操作失败') }
  finally { rejecting.value = false }
}
</script>
