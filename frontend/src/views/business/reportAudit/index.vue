<template>
  <div class="report-audit-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>报告审核管理</span>
          <div>
            <el-input v-model="taskOrderIdFilter" placeholder="委托ID" clearable style="width:160px;margin-right:8px" @clear="loadData" @keyup.enter="loadData" />
            <el-button type="primary" @click="openCreate">新增审核</el-button>
          </div>
        </div>
      </template>

      <el-table :data="items" stripe v-loading="loading">
        <el-table-column prop="task_order_id" label="委托ID" width="80" />
        <el-table-column prop="audit_result" label="审核结果" width="100">
          <template #default="{ row }">
            <el-tag :type="row.audit_result === '通过' ? 'success' : 'danger'" size="small">{{ row.audit_result }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="audit_comment" label="审核意见" min-width="180" show-overflow-tooltip />
        <el-table-column prop="audit_issues" label="问题列表" min-width="200" show-overflow-tooltip />
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
    <el-dialog v-model="formVisible" :title="isEdit ? '编辑报告审核' : '新增报告审核'" width="600px">
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="120px">
        <el-form-item label="委托ID" prop="task_order_id">
          <el-input-number v-model="form.task_order_id" :min="1" style="width:100%" />
        </el-form-item>
        <el-form-item label="审核结果">
          <el-select v-model="form.audit_result" placeholder="选择结果" style="width:100%">
            <el-option label="通过" value="通过" />
            <el-option label="驳回" value="驳回" />
          </el-select>
        </el-form-item>
        <el-form-item label="审核意见">
          <el-input v-model="form.audit_comment" type="textarea" :rows="3" />
        </el-form-item>
        <el-form-item label="问题列表">
          <el-input v-model="form.audit_issues" type="textarea" :rows="3" placeholder="JSON: [{issue, severity, action}]" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="formVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSave">保存</el-button>
      </template>
    </el-dialog>

    <!-- Approve dialog -->
    <el-dialog v-model="approveVisible" title="报告审核 - 通过" width="500px">
      <el-form :model="approveForm" label-width="100px">
        <el-form-item label="委托ID">{{ approveForm.task_order_id }}</el-form-item>
        <el-form-item label="审核结果">
          <el-select v-model="approveForm.audit_result" placeholder="选择结果" style="width:100%">
            <el-option label="通过" value="通过" />
            <el-option label="驳回" value="驳回" />
          </el-select>
        </el-form-item>
        <el-form-item label="审核意见">
          <el-input v-model="approveForm.comment" type="textarea" :rows="3" />
        </el-form-item>
        <el-form-item label="问题列表">
          <el-input v-model="approveForm.audit_issues" type="textarea" :rows="3" placeholder="JSON: [{issue, severity, action}]" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="approveVisible = false">取消</el-button>
        <el-button type="primary" :loading="approving" @click="handleApprove">确认通过</el-button>
      </template>
    </el-dialog>

    <!-- Reject dialog -->
    <el-dialog v-model="rejectVisible" title="报告审核 - 驳回" width="500px">
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
import { getReportAuditList, getReportAudit, createReportAudit, updateReportAudit, deleteReportAudit, approveReportAudit, rejectReportAudit } from '@/api/business'
import type { ReportAudit } from '@/api/business'

const loading = ref(false)
const items = ref<ReportAudit[]>([])
const taskOrderIdFilter = ref('')

onMounted(async () => { await loadData() })

async function loadData() {
  loading.value = true
  try {
    const params: Record<string, string> = {}
    if (taskOrderIdFilter.value) params.task_order_id = taskOrderIdFilter.value
    const res = await getReportAuditList(params)
    items.value = res.data
  } finally { loading.value = false }
}

const formVisible = ref(false); const isEdit = ref(false); const formRef = ref<FormInstance>(); const saving = ref(false)
const form = reactive({ id: 0, task_order_id: 0, audit_result: '', audit_comment: '', audit_issues: '' })
const formRules: FormRules = { task_order_id: [{ required: true, message: '请输入委托ID', trigger: 'blur' }] }
function openCreate() { isEdit.value = false; form.id = 0; form.task_order_id = 0; form.audit_result = ''; form.audit_comment = ''; form.audit_issues = ''; formVisible.value = true }
async function openEdit(row: ReportAudit) { isEdit.value = true; const res = await getReportAudit(row.id); const d = res.data; form.id = d.id; form.task_order_id = d.task_order_id; form.audit_result = d.audit_result; form.audit_comment = d.audit_comment; form.audit_issues = d.audit_issues; formVisible.value = true }
async function handleSave() {
  const valid = await formRef.value?.validate().catch(() => false); if (!valid) return
  saving.value = true
  try {
    const data = { task_order_id: form.task_order_id, audit_result: form.audit_result, audit_comment: form.audit_comment, audit_issues: form.audit_issues }
    if (isEdit.value) { await updateReportAudit(form.id, data); ElMessage.success('更新成功') }
    else { await createReportAudit(data); ElMessage.success('创建成功') }
    formVisible.value = false; await loadData()
  } finally { saving.value = false }
}
async function handleDelete(id: number) { await deleteReportAudit(id); ElMessage.success('删除成功'); await loadData() }

const approveVisible = ref(false); const approving = ref(false)
const approveForm = reactive({ id: 0, task_order_id: 0, audit_result: '通过', comment: '', audit_issues: '' })
function openApprove(row: ReportAudit) { approveForm.id = row.id; approveForm.task_order_id = row.task_order_id; approveForm.audit_result = '通过'; approveForm.comment = ''; approveForm.audit_issues = row.audit_issues; approveVisible.value = true }
async function handleApprove() {
  approving.value = true
  try { await approveReportAudit(approveForm.id, { task_id: approveForm.task_order_id, audit_result: approveForm.audit_result, audit_comment: approveForm.comment, audit_issues: approveForm.audit_issues }); ElMessage.success('审核通过，流程已推进'); approveVisible.value = false; await loadData() }
  catch (e: any) { ElMessage.error(e?.response?.data?.message || '操作失败') }
  finally { approving.value = false }
}

const rejectVisible = ref(false); const rejectFormRef = ref<FormInstance>(); const rejecting = ref(false)
const rejectForm = reactive({ id: 0, task_order_id: 0, comment: '' })
const rejectRules: FormRules = { comment: [{ required: true, message: '请填写驳回原因', trigger: 'blur' }] }
function openReject(row: ReportAudit) { rejectForm.id = row.id; rejectForm.task_order_id = row.task_order_id; rejectForm.comment = ''; rejectVisible.value = true }
async function handleReject() {
  const valid = await rejectFormRef.value?.validate().catch(() => false); if (!valid) return
  rejecting.value = true
  try { await rejectReportAudit(rejectForm.id, { task_id: rejectForm.task_order_id, comment: rejectForm.comment }); ElMessage.success('已驳回'); rejectVisible.value = false; await loadData() }
  catch (e: any) { ElMessage.error(e?.response?.data?.message || '操作失败') }
  finally { rejecting.value = false }
}
</script>