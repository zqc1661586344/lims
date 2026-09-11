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
        <el-table-column prop="report_no" label="报告编号" width="140" show-overflow-tooltip />
        <el-table-column prop="report_title" label="报告标题" min-width="180" show-overflow-tooltip />
        <el-table-column prop="report_file" label="报告文件" min-width="150" show-overflow-tooltip />
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
    <el-dialog v-model="formVisible" :title="isEdit ? '编辑报告编制' : '新增报告编制'" width="600px">
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="120px">
        <el-form-item label="委托ID" prop="task_order_id">
          <el-input-number v-model="form.task_order_id" :min="1" style="width:100%" />
        </el-form-item>
        <el-form-item label="报告编号">
          <el-input v-model="form.report_no" placeholder="请输入报告编号" />
        </el-form-item>
        <el-form-item label="编制意见">
          <el-input v-model="form.prepare_opinion" type="textarea" :rows="2" placeholder="请输入编制意见" />
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

    <ApprovalDialog ref="approvalRef" title="报告编制审批" width="560px" @submit="handleApprovalSubmit">
      <template #header>
        <div class="approval-header">
          <span>委托ID：{{ currentRow?.task_order_id }}</span>
        </div>
      </template>
      <template #extraFields>
        <template v-if="approvalForm.action === 'approve'">
          <el-form-item label="报告编号">
            <el-input v-model="approveExtra.report_no" placeholder="请输入报告编号" />
          </el-form-item>
          <el-form-item label="编制意见">
            <el-input v-model="approveExtra.prepare_opinion" type="textarea" :rows="2" />
          </el-form-item>
          <el-form-item label="报告标题">
            <el-input v-model="approveExtra.report_title" placeholder="请输入报告标题" />
          </el-form-item>
          <el-form-item label="报告内容">
            <el-input v-model="approveExtra.report_content" type="textarea" :rows="3" placeholder="JSON" />
          </el-form-item>
          <el-form-item label="报告文件">
            <el-input v-model="approveExtra.report_file" placeholder="文件路径" />
          </el-form-item>
          <el-form-item label="附件信息">
            <el-input v-model="approveExtra.attachments" type="textarea" :rows="3" placeholder="JSON: [{name, url}]" />
          </el-form-item>
          <el-form-item label="原始记录">
            <div class="raw-records-panel">
              <el-table :data="rawEntries" size="small" border empty-text="暂无实验原始记录" max-height="180">
                <el-table-column prop="id" label="ID" width="60" />
                <el-table-column prop="test_item_id" label="项目ID" width="80" />
                <el-table-column prop="original_data" label="原始数据" min-width="180" show-overflow-tooltip />
                <el-table-column prop="created_at" label="录入时间" width="170" />
              </el-table>
            </div>
          </el-form-item>
        </template>
      </template>
    </ApprovalDialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import ApprovalDialog from '@/components/ApprovalDialog.vue'
import { getReportPrepareList, getReportPrepare, createReportPrepare, updateReportPrepare, deleteReportPrepare, approveReportPrepare, rejectReportPrepare, getDataEntryList } from '@/api/business'
import type { ReportPrepare, DataEntry } from '@/api/business'

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
    items.value = res.data?.items ?? res.data
  } finally { loading.value = false }
}

const formVisible = ref(false); const isEdit = ref(false); const formRef = ref<FormInstance>(); const saving = ref(false)
const form = reactive({ id: 0, task_order_id: 0, report_no: '', prepare_opinion: '', report_title: '', report_content: '', report_file: '', attachments: '' })
const formRules: FormRules = { task_order_id: [{ required: true, message: '请输入委托ID', trigger: 'blur' }], report_title: [{ required: true, message: '请输入报告标题', trigger: 'blur' }] }
function openCreate() { isEdit.value = false; form.id = 0; form.task_order_id = 0; form.report_no = ''; form.prepare_opinion = ''; form.report_title = ''; form.report_content = ''; form.report_file = ''; form.attachments = ''; formVisible.value = true }
async function openEdit(row: ReportPrepare) { isEdit.value = true; const res = await getReportPrepare(row.id); const d = res.data; form.id = d.id; form.task_order_id = d.task_order_id; form.report_no = d.report_no; form.prepare_opinion = d.prepare_opinion; form.report_title = d.report_title; form.report_content = d.report_content; form.report_file = d.report_file; form.attachments = d.attachments; formVisible.value = true }
async function handleSave() {
  const valid = await formRef.value?.validate().catch(() => false); if (!valid) return
  saving.value = true
  try {
    const data = { task_order_id: form.task_order_id, report_no: form.report_no, prepare_opinion: form.prepare_opinion, report_title: form.report_title, report_content: form.report_content, report_file: form.report_file, attachments: form.attachments }
    if (isEdit.value) { await updateReportPrepare(form.id, data); ElMessage.success('更新成功') }
    else { await createReportPrepare(data); ElMessage.success('创建成功') }
    formVisible.value = false; await loadData()
  } finally { saving.value = false }
}
async function handleDelete(id: number) { await deleteReportPrepare(id); ElMessage.success('删除成功'); await loadData() }

const approvalRef = ref<InstanceType<typeof ApprovalDialog>>()
const currentRow = ref<ReportPrepare | null>(null)
const approvalForm = reactive({ action: 'approve' as 'approve' | 'reject' })
const approveExtra = reactive({ report_no: '', prepare_opinion: '', report_title: '', report_content: '', report_file: '', attachments: '' })

const rawEntries = ref<DataEntry[]>([])

async function loadRawRecords(taskOrderId: number) {
  rawEntries.value = []
  if (!taskOrderId) return
  try {
    const dr = await getDataEntryList({ task_order_id: String(taskOrderId) })
    rawEntries.value = (dr.data || []) as DataEntry[]
  } catch { rawEntries.value = [] }
}

function openApprove(row: ReportPrepare) {
  currentRow.value = row
  approveExtra.report_no = row.report_no ?? ''
  approveExtra.prepare_opinion = row.prepare_opinion ?? ''
  approveExtra.report_title = row.report_title ?? ''
  approveExtra.report_content = row.report_content ?? ''
  approveExtra.report_file = row.report_file ?? ''
  approveExtra.attachments = row.attachments ?? ''
  approvalForm.action = 'approve'
  approvalRef.value?.open('approve')
  loadRawRecords(row.task_order_id)
}
function openReject(row: ReportPrepare) {
  currentRow.value = row
  approvalForm.action = 'reject'
  approvalRef.value?.open('reject')
}

async function handleApprovalSubmit(data: { action: string; comment: string }) {
  if (!currentRow.value) return
  const id = currentRow.value.id
  const taskId = currentRow.value.task_order_id
  if (data.action === 'approve') {
    await approveReportPrepare(id, { task_id: taskId, report_no: approveExtra.report_no, prepare_opinion: approveExtra.prepare_opinion, report_title: approveExtra.report_title, report_content: approveExtra.report_content, report_file: approveExtra.report_file, attachments: approveExtra.attachments })
    ElMessage.success('编制通过，流程已推进')
  } else {
    await rejectReportPrepare(id, { task_id: taskId, comment: data.comment })
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
.raw-records-panel {
  width: 100%;
}
</style>