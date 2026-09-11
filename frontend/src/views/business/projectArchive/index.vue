<template>
  <div class="project-archive-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>项目归档</span>
          <div>
            <el-input v-model="taskOrderIdFilter" placeholder="委托ID" clearable style="width:160px;margin-right:8px" @clear="loadData" @keyup.enter="loadData" />
            <el-button type="primary" @click="openCreate">新增归档</el-button>
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
        <el-table-column prop="archive_no" label="归档编号" width="150" />
        <el-table-column prop="archive_location" label="归档位置" width="150" />
        <el-table-column prop="archive_date" label="归档日期" width="170" />
        <el-table-column prop="retention_period" label="保存期限(月)" width="120" />
        <el-table-column prop="archive_comment" label="归档备注" min-width="150" show-overflow-tooltip />
        <el-table-column label="归档文件" min-width="220" show-overflow-tooltip>
          <template #default="{ row }">
            {{ formatArchiveFiles(row.archive_files) }}
          </template>
        </el-table-column>
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
    <el-dialog v-model="formVisible" :title="isEdit ? '编辑项目归档' : '新增项目归档'" width="600px">
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="120px">
        <el-form-item label="委托ID" prop="task_order_id">
          <el-input-number v-model="form.task_order_id" :min="1" style="width:100%" />
        </el-form-item>
        <el-form-item label="归档编号">
          <el-input v-model="form.archive_no" placeholder="归档编号" />
        </el-form-item>
        <el-form-item label="归档位置">
          <el-input v-model="form.archive_location" placeholder="归档位置（如：档案室A-12）" />
        </el-form-item>
        <el-form-item label="归档日期">
          <el-date-picker v-model="form.archive_date" type="date" placeholder="选择日期" style="width:100%" />
        </el-form-item>
        <el-form-item label="保存期限(月)">
          <el-input-number v-model="form.retention_period" :min="1" style="width:100%" />
        </el-form-item>
        <el-form-item label="归档文件">
          <el-input v-model="form.archive_files" type="textarea" :rows="3" placeholder="JSON: [{name, type, path}]" />
        </el-form-item>
        <el-form-item label="归档备注">
          <el-input v-model="form.archive_comment" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="formVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSave">保存</el-button>
      </template>
    </el-dialog>

    <ApprovalDialog ref="approvalRef" title="项目归档审批" width="500px" @submit="handleApprovalSubmit">
      <template #header>
        <div class="approval-header">
          <span>委托ID：{{ currentRow?.task_order_id }}</span>
        </div>
      </template>
      <template #extraFields>
        <template v-if="approvalForm.action === 'approve'">
          <el-form-item label="归档编号">
            <el-input v-model="approveExtra.archive_no" placeholder="归档编号" />
          </el-form-item>
          <el-form-item label="归档位置">
            <el-input v-model="approveExtra.archive_location" placeholder="归档位置" />
          </el-form-item>
          <el-form-item label="归档日期">
            <el-date-picker v-model="approveExtra.archive_date" type="date" placeholder="选择日期" style="width:100%" />
          </el-form-item>
          <el-form-item label="保存期限(月)">
            <el-input-number v-model="approveExtra.retention_period" :min="1" style="width:100%" />
          </el-form-item>
          <el-form-item label="归档文件">
            <el-alert type="info" :closable="false" class="record-preview-alert" title="归档文件清单将自动汇聚本委托全部环节文档（D1–D13）">
              提交确认后将由系统自动生成，无需手动填写。
            </el-alert>
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
import { getProjectArchiveList, getProjectArchive, createProjectArchive, updateProjectArchive, deleteProjectArchive, approveProjectArchive, rejectProjectArchive } from '@/api/business'
import type { ProjectArchive } from '@/api/business'

const loading = ref(false)
const items = ref<ProjectArchive[]>([])
const selectedRow = ref(null)
function handleRowClick(row: any) { selectedRow.value = row }
const taskOrderIdFilter = ref('')

function formatArchiveFiles(files: string) {
  if (!files) return '未生成'
  try {
    const arr = JSON.parse(files)
    if (!Array.isArray(arr)) return files
    return arr.map((f: any) => `${f.doc_name || f.stage || ''}${f.ref ? ': ' + f.ref : ''}`).join('；')
  } catch {
    return files
  }
}

onMounted(async () => { await loadData() })

async function loadData() {
  loading.value = true
  try {
    const params: Record<string, string> = {}
    if (taskOrderIdFilter.value) params.task_order_id = taskOrderIdFilter.value
    const res = await getProjectArchiveList(params)
    items.value = res.data?.items ?? res.data
  } finally { loading.value = false }
}

const formVisible = ref(false); const isEdit = ref(false); const formRef = ref<FormInstance>(); const saving = ref(false)
const form = reactive({ id: 0, task_order_id: 0, archive_no: '', archive_location: '', archive_date: '', archive_files: '', archive_comment: '', retention_period: 36 })
const formRules: FormRules = { task_order_id: [{ required: true, message: '请输入委托ID', trigger: 'blur' }] }
function openCreate() { isEdit.value = false; form.id = 0; form.task_order_id = 0; form.archive_no = ''; form.archive_location = ''; form.archive_date = ''; form.archive_files = ''; form.archive_comment = ''; form.retention_period = 36; formVisible.value = true }
async function openEdit(row: ProjectArchive) { isEdit.value = true; const res = await getProjectArchive(row.id); const d = res.data; form.id = d.id; form.task_order_id = d.task_order_id; form.archive_no = d.archive_no; form.archive_location = d.archive_location; form.archive_date = d.archive_date; form.archive_files = d.archive_files; form.archive_comment = d.archive_comment; form.retention_period = d.retention_period || 36; formVisible.value = true }
async function handleSave() {
  const valid = await formRef.value?.validate().catch(() => false); if (!valid) return
  saving.value = true
  try {
    const data = { task_order_id: form.task_order_id, archive_no: form.archive_no, archive_location: form.archive_location, archive_date: form.archive_date, archive_files: form.archive_files, archive_comment: form.archive_comment, retention_period: form.retention_period }
    if (isEdit.value) { await updateProjectArchive(form.id, data); ElMessage.success('更新成功') }
    else { await createProjectArchive(data); ElMessage.success('创建成功') }
    formVisible.value = false; await loadData()
  } finally { saving.value = false }
}
async function handleDelete(id: number) { await deleteProjectArchive(id); ElMessage.success('删除成功'); await loadData() }

const approvalRef = ref<InstanceType<typeof ApprovalDialog>>()
const currentRow = ref<ProjectArchive | null>(null)
const approvalForm = reactive({ action: 'approve' as 'approve' | 'reject' })
const approveExtra = reactive({ archive_no: '', archive_location: '', archive_date: '', retention_period: 36 })

function openApprove(row: ProjectArchive) {
  currentRow.value = row
  approveExtra.archive_no = row.archive_no ?? ''
  approveExtra.archive_location = row.archive_location ?? ''
  approveExtra.archive_date = row.archive_date ?? ''
  approveExtra.retention_period = row.retention_period || 36
  approvalForm.action = 'approve'
  approvalRef.value?.open('approve')
}
function openReject(row: ProjectArchive) {
  currentRow.value = row
  approvalForm.action = 'reject'
  approvalRef.value?.open('reject')
}

async function handleApprovalSubmit(data: { action: string; comment: string }) {
  if (!currentRow.value) return
  const id = currentRow.value.id
  const taskId = currentRow.value.task_order_id
  if (data.action === 'approve') {
    await approveProjectArchive(id, { task_id: taskId, archive_no: approveExtra.archive_no, archive_location: approveExtra.archive_location, archive_date: approveExtra.archive_date, archive_files: '', archive_comment: data.comment, retention_period: approveExtra.retention_period })
    ElMessage.success('项目归档完成，流程已结束')
  } else {
    await rejectProjectArchive(id, { task_id: taskId, comment: data.comment })
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
.record-preview-alert {
  width: 100%;
  margin-bottom: 0;
}
</style>