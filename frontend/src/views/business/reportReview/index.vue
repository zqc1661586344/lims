<template>
  <div class="report-review-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>报告复核管理</span>
          <div>
            <el-input v-model="taskOrderIdFilter" placeholder="委托ID" clearable style="width:160px;margin-right:8px" @clear="loadData" @keyup.enter="loadData" />
            <el-button type="primary" @click="openCreate">新增复核</el-button>
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
        <el-table-column prop="review_result" label="复核结果" width="100">
          <template #default="{ row }">
            <el-tag :type="row.review_result === '通过' ? 'success' : 'danger'" size="small">{{ row.review_result }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="review_comment" label="复核意见" min-width="180" show-overflow-tooltip />
        <el-table-column prop="reviewed_items" label="复核项目" min-width="200" show-overflow-tooltip />
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
    <el-dialog v-model="formVisible" :title="isEdit ? '编辑报告复核' : '新增报告复核'" width="600px">
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="120px">
        <el-form-item label="委托ID" prop="task_order_id">
          <el-input-number v-model="form.task_order_id" :min="1" style="width:100%" />
        </el-form-item>
        <el-form-item label="复核结果">
          <el-select v-model="form.review_result" placeholder="选择结果" style="width:100%">
            <el-option label="通过" value="通过" />
            <el-option label="驳回" value="驳回" />
          </el-select>
        </el-form-item>
        <el-form-item label="复核意见">
          <el-input v-model="form.review_comment" type="textarea" :rows="3" />
        </el-form-item>
        <el-form-item label="复核项目">
          <el-input v-model="form.reviewed_items" type="textarea" :rows="3" placeholder="JSON: [{item, result}]" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="formVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSave">保存</el-button>
      </template>
    </el-dialog>

    <ApprovalDialog ref="approvalRef" title="报告复核审批" width="620px" @submit="handleApprovalSubmit">
      <template #header>
        <div class="approval-header">
          <span>委托ID：{{ currentRow?.task_order_id }}</span>
        </div>
      </template>
      <template #extraFields>
        <template v-if="approvalForm.action === 'approve'">
          <el-alert v-if="sourceReport" type="info" :closable="false" title="承接报告" class="source-report-alert">
            <div class="source-report-body">
              <div><b>报告标题：</b>{{ sourceReport.report_title || '（未填写）' }}</div>
              <div><b>报告文件：</b>{{ sourceReport.report_file || '（无）' }}</div>
              <div v-if="sourceReport.report_content"><b>报告内容：</b>{{ sourceReport.report_content }}</div>
            </div>
          </el-alert>
          <el-alert v-else-if="currentRow?.task_order_id && sourceLoaded" type="warning" :closable="false" title="未找到关联的报告编制记录" class="source-report-alert" />
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
          <el-form-item label="复核结果">
            <el-select v-model="approveExtra.review_result" placeholder="选择结果" style="width:100%">
              <el-option label="通过" value="通过" />
              <el-option label="驳回" value="驳回" />
            </el-select>
          </el-form-item>
          <el-form-item label="复核项目">
            <el-input v-model="approveExtra.reviewed_items" type="textarea" :rows="3" placeholder="JSON: [{item, result}]" />
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
import { getReportReviewList, getReportReview, createReportReview, updateReportReview, deleteReportReview, approveReportReview, rejectReportReview, getReportPrepareList, getDataEntryList } from '@/api/business'
import type { ReportReview, ReportPrepare, DataEntry } from '@/api/business'

const loading = ref(false)
const items = ref<ReportReview[]>([])
const selectedRow = ref(null)
function handleRowClick(row: any) { selectedRow.value = row }
const taskOrderIdFilter = ref('')

onMounted(async () => { await loadData() })

async function loadData() {
  loading.value = true
  try {
    const params: Record<string, string> = {}
    if (taskOrderIdFilter.value) params.task_order_id = taskOrderIdFilter.value
    const res = await getReportReviewList(params)
    items.value = res.data?.items ?? res.data
  } finally { loading.value = false }
}

const formVisible = ref(false); const isEdit = ref(false); const formRef = ref<FormInstance>(); const saving = ref(false)
const form = reactive({ id: 0, task_order_id: 0, review_result: '', review_comment: '', reviewed_items: '' })
const formRules: FormRules = { task_order_id: [{ required: true, message: '请输入委托ID', trigger: 'blur' }] }
function openCreate() { isEdit.value = false; form.id = 0; form.task_order_id = 0; form.review_result = ''; form.review_comment = ''; form.reviewed_items = ''; formVisible.value = true }
async function openEdit(row: ReportReview) { isEdit.value = true; const res = await getReportReview(row.id); const d = res.data; form.id = d.id; form.task_order_id = d.task_order_id; form.review_result = d.review_result; form.review_comment = d.review_comment; form.reviewed_items = d.reviewed_items; formVisible.value = true }
async function handleSave() {
  const valid = await formRef.value?.validate().catch(() => false); if (!valid) return
  saving.value = true
  try {
    const data = { task_order_id: form.task_order_id, review_result: form.review_result, review_comment: form.review_comment, reviewed_items: form.reviewed_items }
    if (isEdit.value) { await updateReportReview(form.id, data); ElMessage.success('更新成功') }
    else { await createReportReview(data); ElMessage.success('创建成功') }
    formVisible.value = false; await loadData()
  } finally { saving.value = false }
}
async function handleDelete(id: number) { await deleteReportReview(id); ElMessage.success('删除成功'); await loadData() }

const approvalRef = ref<InstanceType<typeof ApprovalDialog>>()
const currentRow = ref<ReportReview | null>(null)
const approvalForm = reactive({ action: 'approve' as 'approve' | 'reject' })
const approveExtra = reactive({ review_result: '通过', reviewed_items: '' })

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
    const dr = await getDataEntryList({ task_order_id: String(taskOrderId) })
    rawEntries.value = (dr.data || []) as DataEntry[]
  } finally {
    sourceLoaded.value = true
  }
}

function openApprove(row: ReportReview) {
  currentRow.value = row
  approveExtra.review_result = '通过'; approveExtra.reviewed_items = row.reviewed_items
  approvalForm.action = 'approve'
  approvalRef.value?.open('approve')
  loadSourceReport(row.task_order_id)
}
function openReject(row: ReportReview) {
  currentRow.value = row
  approvalForm.action = 'reject'
  approvalRef.value?.open('reject')
}

async function handleApprovalSubmit(data: { action: string; comment: string; reject_target?: string }) {
  if (!currentRow.value) return
  const id = currentRow.value.id
  const taskId = currentRow.value.task_order_id
  if (data.action === 'approve') {
    await approveReportReview(id, { task_id: taskId, review_result: approveExtra.review_result, review_comment: data.comment, reviewed_items: approveExtra.reviewed_items })
    ElMessage.success('复核通过，流程已推进')
  } else {
    await rejectReportReview(id, { task_id: taskId, comment: data.comment, reject_target: data.reject_target })
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