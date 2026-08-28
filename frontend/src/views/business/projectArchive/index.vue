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

      <el-table :data="items" stripe v-loading="loading">
        <el-table-column prop="task_order_id" label="委托ID" width="80" />
        <el-table-column prop="archive_no" label="归档编号" width="150" />
        <el-table-column prop="archive_location" label="归档位置" width="150" />
        <el-table-column prop="archive_date" label="归档日期" width="170" />
        <el-table-column prop="retention_period" label="保存期限(月)" width="120" />
        <el-table-column prop="archive_comment" label="归档备注" min-width="180" show-overflow-tooltip />
        <el-table-column prop="archive_files" label="归档文件" min-width="200" show-overflow-tooltip />
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

    <!-- Approve dialog -->
    <el-dialog v-model="approveVisible" title="项目归档 - 通过" width="500px">
      <el-form :model="approveForm" label-width="100px">
        <el-form-item label="委托ID">{{ approveForm.task_order_id }}</el-form-item>
        <el-form-item label="归档编号">
          <el-input v-model="approveForm.archive_no" placeholder="归档编号" />
        </el-form-item>
        <el-form-item label="归档位置">
          <el-input v-model="approveForm.archive_location" placeholder="归档位置" />
        </el-form-item>
        <el-form-item label="归档日期">
          <el-date-picker v-model="approveForm.archive_date" type="date" placeholder="选择日期" style="width:100%" />
        </el-form-item>
        <el-form-item label="保存期限(月)">
          <el-input-number v-model="approveForm.retention_period" :min="1" style="width:100%" />
        </el-form-item>
        <el-form-item label="归档文件">
          <el-input v-model="approveForm.archive_files" type="textarea" :rows="3" placeholder="JSON: [{name, type, path}]" />
        </el-form-item>
        <el-form-item label="归档备注">
          <el-input v-model="approveForm.comment" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="approveVisible = false">取消</el-button>
        <el-button type="primary" :loading="approving" @click="handleApprove">确认归档</el-button>
      </template>
    </el-dialog>

    <!-- Reject dialog -->
    <el-dialog v-model="rejectVisible" title="项目归档 - 驳回" width="500px">
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
import { getProjectArchiveList, getProjectArchive, createProjectArchive, updateProjectArchive, deleteProjectArchive, approveProjectArchive, rejectProjectArchive } from '@/api/business'
import type { ProjectArchive } from '@/api/business'

const loading = ref(false)
const items = ref<ProjectArchive[]>([])
const taskOrderIdFilter = ref('')

onMounted(async () => { await loadData() })

async function loadData() {
  loading.value = true
  try {
    const params: Record<string, string> = {}
    if (taskOrderIdFilter.value) params.task_order_id = taskOrderIdFilter.value
    const res = await getProjectArchiveList(params)
    items.value = res.data
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

const approveVisible = ref(false); const approving = ref(false)
const approveForm = reactive({ id: 0, task_order_id: 0, archive_no: '', archive_location: '', archive_date: '', archive_files: '', comment: '', retention_period: 36 })
function openApprove(row: ProjectArchive) { approveForm.id = row.id; approveForm.task_order_id = row.task_order_id; approveForm.archive_no = row.archive_no; approveForm.archive_location = row.archive_location; approveForm.archive_date = row.archive_date; approveForm.archive_files = row.archive_files; approveForm.retention_period = row.retention_period || 36; approveForm.comment = ''; approveVisible.value = true }
async function handleApprove() {
  approving.value = true
  try { await approveProjectArchive(approveForm.id, { task_id: approveForm.task_order_id, archive_no: approveForm.archive_no, archive_location: approveForm.archive_location, archive_date: approveForm.archive_date, archive_files: approveForm.archive_files, archive_comment: approveForm.comment, retention_period: approveForm.retention_period }); ElMessage.success('项目归档完成，流程已结束'); approveVisible.value = false; await loadData() }
  catch (e: any) { ElMessage.error(e?.response?.data?.message || '操作失败') }
  finally { approving.value = false }
}

const rejectVisible = ref(false); const rejectFormRef = ref<FormInstance>(); const rejecting = ref(false)
const rejectForm = reactive({ id: 0, task_order_id: 0, comment: '' })
const rejectRules: FormRules = { comment: [{ required: true, message: '请填写驳回原因', trigger: 'blur' }] }
function openReject(row: ProjectArchive) { rejectForm.id = row.id; rejectForm.task_order_id = row.task_order_id; rejectForm.comment = ''; rejectVisible.value = true }
async function handleReject() {
  const valid = await rejectFormRef.value?.validate().catch(() => false); if (!valid) return
  rejecting.value = true
  try { await rejectProjectArchive(rejectForm.id, { task_id: rejectForm.task_order_id, comment: rejectForm.comment }); ElMessage.success('已驳回'); rejectVisible.value = false; await loadData() }
  catch (e: any) { ElMessage.error(e?.response?.data?.message || '操作失败') }
  finally { rejecting.value = false }
}
</script>
