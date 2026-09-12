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
        <el-table-column prop="report_no" label="报告编号" width="140" show-overflow-tooltip />
        <el-table-column prop="report_title" label="报告标题" min-width="160" show-overflow-tooltip />
        <el-table-column prop="sign_result" label="签发结果" width="100">
          <template #default="{ row }">
            <el-tag :type="row.sign_result === '通过' ? 'success' : 'danger'" size="small">{{ row.sign_result }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="signer_name" label="签发人" width="120" />
        <el-table-column prop="sign_date" label="签发日期" width="170" />
        <el-table-column prop="sign_stamp" label="印章文件" width="150" show-overflow-tooltip />
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

    <ApprovalDialog ref="approvalRef" title="报告签发审批" width="680px" @submit="handleApprovalSubmit">
      <template #header>
        <div class="approval-header">
          <span>委托ID：{{ currentRow?.task_order_id }}</span>
        </div>
      </template>
      <template #extraFields>
        <template v-if="approvalForm.action === 'approve'">
          <el-alert v-if="sourceReport" type="info" :closable="false" title="承接报告" class="source-report-alert">
            <div class="source-report-body">
              <div><b>报告编号：</b>{{ sourceReport.report_no || '（未填写）' }}</div>
              <div><b>报告标题：</b>{{ sourceReport.report_title || '（未填写）' }}</div>
              <div><b>报告文件：</b>{{ sourceReport.report_file || '（无）' }}</div>
              <div v-if="sourceReport.report_content"><b>报告内容：</b>{{ sourceReport.report_content }}</div>
            </div>
          </el-alert>
          <el-alert v-else-if="currentRow?.task_order_id && sourceLoaded" type="warning" :closable="false" title="未找到关联的报告编制记录" class="source-report-alert" />
          <el-form-item label="编制人意见" v-if="sourceReport">
            <el-input :model-value="sourceReport.prepare_opinion || '（未填写）'" type="textarea" :rows="2" readonly />
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
          <el-form-item label="报告编号">
            <el-input v-model="approveExtra.report_no" placeholder="沿用报告编制赋号" />
          </el-form-item>
          <el-form-item label="签发结果">
            <el-select v-model="approveExtra.sign_result" placeholder="选择结果" style="width:100%">
              <el-option label="通过" value="通过" />
              <el-option label="驳回" value="驳回" />
            </el-select>
          </el-form-item>
          <el-form-item label="签发人">
            <el-input v-model="approveExtra.signer_name" placeholder="签发人姓名" />
          </el-form-item>
          <el-form-item label="签发日期">
            <el-date-picker v-model="approveExtra.sign_date" type="date" placeholder="选择日期" style="width:100%" />
          </el-form-item>
          <el-form-item label="印章文件">
            <el-input v-model="approveExtra.sign_stamp" placeholder="印章文件路径" />
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
import TaskProcessBar from '@/components/TaskProcessBar.vue'
import { getReportSignList, getReportSign, createReportSign, updateReportSign, deleteReportSign, approveReportSign, rejectReportSign, getReportPrepareList, getDataEntryList } from '@/api/business'
import type { ReportSign, ReportPrepare, DataEntry } from '@/api/business'

const loading = ref(false)
const items = ref<ReportSign[]>([])
const selectedRow = ref(null)
function handleRowClick(row: any) { selectedRow.value = row }
const taskOrderIdFilter = ref('')

onMounted(async () => { await loadData() })

async function loadData() {
  loading.value = true
  try {
    const params: Record<string, string> = {}
    if (taskOrderIdFilter.value) params.task_order_id = taskOrderIdFilter.value
    const res = await getReportSignList(params)
    items.value = res.data?.items ?? res.data
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

const approvalRef = ref<InstanceType<typeof ApprovalDialog>>()
const currentRow = ref<ReportSign | null>(null)
const approvalForm = reactive({ action: 'approve' as 'approve' | 'reject' })
const approveExtra = reactive({ report_no: '', sign_result: '通过', signer_name: '', sign_date: '', sign_stamp: '' })

const sourceReport = ref<ReportPrepare | null>(null)
const sourceLoaded = ref(false)
const rawEntries = ref<DataEntry[]>([])

async function loadSourceReport(taskOrderId: number) {
  sourceReport.value = null
  sourceLoaded.value = false
  rawEntries.value = []
  if (!taskOrderId) return
  try {
    const res = await getReportPrepareList({ task_order_id: String(taskOrderId) })
    const list = ((res.data?.items ?? res.data) || []) as ReportPrepare[]
    sourceReport.value = list.length ? list[0] : null
    if (list.length) approveExtra.report_no = list[0].report_no || ''
    const dr = await getDataEntryList({ task_order_id: String(taskOrderId) })
    rawEntries.value = (dr.data || []) as DataEntry[]
  } finally {
    sourceLoaded.value = true
  }
}

function openApprove(row: ReportSign) {
  currentRow.value = row
  approveExtra.report_no = row.report_no ?? ''
  approveExtra.sign_result = '通过'
  approveExtra.signer_name = row.signer_name ?? ''
  approveExtra.sign_date = row.sign_date ?? ''
  approveExtra.sign_stamp = row.sign_stamp ?? ''
  approvalForm.action = 'approve'
  approvalRef.value?.open('approve')
  loadSourceReport(row.task_order_id)
}
function openReject(row: ReportSign) {
  currentRow.value = row
  approvalForm.action = 'reject'
  approvalRef.value?.open('reject')
}

async function handleApprovalSubmit(data: { action: string; comment: string; reject_target?: string }) {
  if (!currentRow.value) return
  const id = currentRow.value.id
  const taskId = currentRow.value.task_order_id
  if (data.action === 'approve') {
    await approveReportSign(id, { task_id: taskId, report_no: approveExtra.report_no, sign_result: approveExtra.sign_result, sign_comment: data.comment, signer_name: approveExtra.signer_name, sign_date: approveExtra.sign_date, sign_stamp: approveExtra.sign_stamp })
    ElMessage.success('签发通过，流程已推进')
  } else {
    await rejectReportSign(id, { task_id: taskId, comment: data.comment, reject_target: data.reject_target })
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
.source-report-alert {
  margin-bottom: 14px;
}
.source-report-body {
  font-size: 13px;
  line-height: 1.7;
  margin-top: 4px;
}
.raw-records-panel {
  width: 100%;
}
</style>