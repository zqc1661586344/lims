<template>
  <div class="report-prepare-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>报告编制管理</span>
          <div>
            <el-input v-model="taskOrderIdFilter" placeholder="委托ID" clearable style="width:160px;margin-right:8px" @clear="loadData" @keyup.enter="loadData" />
            <el-button type="primary" @click="openCreate">新增编制</el-button>
          </div>
        </div>
      </template>

      <el-table :data="items" stripe v-loading="loading">
        <el-table-column prop="task_order_id" label="委托ID" width="80" />
        <el-table-column prop="report_title" label="报告标题" min-width="180" show-overflow-tooltip />
        <el-table-column prop="report_file" label="报告文件" min-width="150" show-overflow-tooltip />
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
    <el-dialog v-model="formVisible" :title="isEdit ? '编辑报告编制' : '新增报告编制'" width="600px">
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="120px">
        <el-form-item label="委托ID" prop="task_order_id">
          <el-input-number v-model="form.task_order_id" :min="1" style="width:100%" />
        </el-form-item>
        <el-form-item label="报告标题" prop="report_title">
          <el-input v-model="form.report_title" placeholder="请输入报告标题" />
        </el-form-item>
        <el-form-item label="报告内容">
          <el-input v-model="form.report_content" type="textarea" :rows="3" placeholder="JSON: 报告内容" />
        </el-form-item>
        <el-form-item label="报告文件">
          <el-input v-model="form.report_file" placeholder="文件路径" />
        </el-form-item>
        <el-form-item label="附件信息">
          <el-input v-model="form.attachments" type="textarea" :rows="3" placeholder="JSON: [{name, url}]" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="formVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSave">保存</el-button>
      </template>
    </el-dialog>

    <!-- Approve dialog -->
    <el-dialog v-model="approveVisible" title="报告编制 - 通过" width="500px">
      <el-form :model="approveForm" label-width="100px">
        <el-form-item label="委托ID">{{ approveForm.task_order_id }}</el-form-item>
        <el-form-item label="报告标题">
          <el-input v-model="approveForm.report_title" placeholder="请输入报告标题" />
        </el-form-item>
        <el-form-item label="报告内容">
          <el-input v-model="approveForm.report_content" type="textarea" :rows="3" placeholder="JSON: 报告内容" />
        </el-form-item>
        <el-form-item label="报告文件">
          <el-input v-model="approveForm.report_file" placeholder="文件路径" />
        </el-form-item>
        <el-form-item label="附件信息">
          <el-input v-model="approveForm.attachments" type="textarea" :rows="3" placeholder="JSON: [{name, url}]" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="approveForm.comment" type="textarea" :rows="2" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="approveVisible = false">取消</el-button>
        <el-button type="primary" :loading="approving" @click="handleApprove">确认通过</el-button>
      </template>
    </el-dialog>

    <!-- Reject dialog -->
    <el-dialog v-model="rejectVisible" title="报告编制 - 驳回" width="500px">
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
import { getReportPrepareList, getReportPrepare, createReportPrepare, updateReportPrepare, deleteReportPrepare, approveReportPrepare, rejectReportPrepare } from '@/api/business'
import type { ReportPrepare } from '@/api/business'

const loading = ref(false)
const items = ref<ReportPrepare[]>([])
const taskOrderIdFilter = ref('')

onMounted(async () => { await loadData() })

async function loadData() {
  loading.value = true
  try {
    const params: Record<string, string> = {}
    if (taskOrderIdFilter.value) params.task_order_id = taskOrderIdFilter.value
    const res = await getReportPrepareList(params)
    items.value = res.data
  } finally { loading.value = false }
}

const formVisible = ref(false); const isEdit = ref(false); const formRef = ref<FormInstance>(); const saving = ref(false)
const form = reactive({ id: 0, task_order_id: 0, report_title: '', report_content: '', report_file: '', attachments: '' })
const formRules: FormRules = { task_order_id: [{ required: true, message: '请输入委托ID', trigger: 'blur' }], report_title: [{ required: true, message: '请输入报告标题', trigger: 'blur' }] }
function openCreate() { isEdit.value = false; form.id = 0; form.task_order_id = 0; form.report_title = ''; form.report_content = ''; form.report_file = ''; form.attachments = ''; formVisible.value = true }
async function openEdit(row: ReportPrepare) { isEdit.value = true; const res = await getReportPrepare(row.id); const d = res.data; form.id = d.id; form.task_order_id = d.task_order_id; form.report_title = d.report_title; form.report_content = d.report_content; form.report_file = d.report_file; form.attachments = d.attachments; formVisible.value = true }
async function handleSave() {
  const valid = await formRef.value?.validate().catch(() => false); if (!valid) return
  saving.value = true
  try {
    const data = { task_order_id: form.task_order_id, report_title: form.report_title, report_content: form.report_content, report_file: form.report_file, attachments: form.attachments }
    if (isEdit.value) { await updateReportPrepare(form.id, data); ElMessage.success('更新成功') }
    else { await createReportPrepare(data); ElMessage.success('创建成功') }
    formVisible.value = false; await loadData()
  } finally { saving.value = false }
}
async function handleDelete(id: number) { await deleteReportPrepare(id); ElMessage.success('删除成功'); await loadData() }

const approveVisible = ref(false); const approving = ref(false)
const approveForm = reactive({ id: 0, task_order_id: 0, report_title: '', report_content: '', report_file: '', attachments: '', comment: '' })
function openApprove(row: ReportPrepare) { approveForm.id = row.id; approveForm.task_order_id = row.task_order_id; approveForm.report_title = row.report_title; approveForm.report_content = row.report_content; approveForm.report_file = row.report_file; approveForm.attachments = row.attachments; approveForm.comment = ''; approveVisible.value = true }
async function handleApprove() {
  approving.value = true
  try { await approveReportPrepare(approveForm.id, { task_id: approveForm.task_order_id, report_title: approveForm.report_title, report_content: approveForm.report_content, report_file: approveForm.report_file, attachments: approveForm.attachments }); ElMessage.success('编制通过，流程已推进'); approveVisible.value = false; await loadData() }
  catch (e: any) { ElMessage.error(e?.response?.data?.message || '操作失败') }
  finally { approving.value = false }
}

const rejectVisible = ref(false); const rejectFormRef = ref<FormInstance>(); const rejecting = ref(false)
const rejectForm = reactive({ id: 0, task_order_id: 0, comment: '' })
const rejectRules: FormRules = { comment: [{ required: true, message: '请填写驳回原因', trigger: 'blur' }] }
function openReject(row: ReportPrepare) { rejectForm.id = row.id; rejectForm.task_order_id = row.task_order_id; rejectForm.comment = ''; rejectVisible.value = true }
async function handleReject() {
  const valid = await rejectFormRef.value?.validate().catch(() => false); if (!valid) return
  rejecting.value = true
  try { await rejectReportPrepare(rejectForm.id, { task_id: rejectForm.task_order_id, comment: rejectForm.comment }); ElMessage.success('已驳回'); rejectVisible.value = false; await loadData() }
  catch (e: any) { ElMessage.error(e?.response?.data?.message || '操作失败') }
  finally { rejecting.value = false }
}
</script>