<template>
  <div class="field-sampling-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>现场采样管理</span>
          <div>
            <el-input v-model="taskOrderIdFilter" placeholder="委托ID" clearable style="width:160px;margin-right:8px" @clear="loadData" @keyup.enter="loadData" />
            <el-button type="primary" @click="openCreate">新增采样记录</el-button>
          </div>
        </div>
      </template>

      <el-table :data="items" stripe v-loading="loading">
        <el-table-column prop="task_order_id" label="委托ID" width="80" />
        <el-table-column label="采样单" min-width="200">
          <template #default="{ row }">
            <el-button size="small" link type="primary" @click="openSheetEditor(row)">编辑采样单</el-button>
            <el-tag v-if="row.sampling_sheet_id" type="success" size="small" style="margin-left:6px">已关联 #{{ row.sampling_sheet_id }}</el-tag>
            <el-tag v-else type="info" size="small" style="margin-left:6px">未关联</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="sample_photos" label="现场照片" min-width="160" show-overflow-tooltip />
        <el-table-column prop="equipment_cal_records" label="设备校准" min-width="160" show-overflow-tooltip />
        <el-table-column prop="sampling_record_file_path" label="采样记录文件" min-width="180" show-overflow-tooltip />
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
    <el-dialog v-model="formVisible" :title="isEdit ? '编辑现场采样' : '新增现场采样'" width="600px">
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="120px">
        <el-form-item label="委托ID" prop="task_order_id">
          <el-input-number v-model="form.task_order_id" :min="1" style="width:100%" />
        </el-form-item>
        <el-form-item label="现场照片">
          <el-input v-model="form.sample_photos" type="textarea" :rows="2" placeholder="JSON: [{url,description}]" />
        </el-form-item>
        <el-form-item label="设备校准记录">
          <el-input v-model="form.equipment_cal_records" type="textarea" :rows="2" placeholder="JSON: [{equipment_id,cal_result}]" />
        </el-form-item>
        <el-form-item label="采样记录文件路径">
          <el-input v-model="form.sampling_record_file_path" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="formVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSave">保存</el-button>
      </template>
    </el-dialog>

    <ApprovalDialog ref="approvalRef" title="现场采样审批" @submit="handleApprovalSubmit">
      <template #header>
        <div class="approval-header">
          <span>委托ID：{{ currentRow?.task_order_id }}</span>
        </div>
      </template>
    </ApprovalDialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import ApprovalDialog from '@/components/ApprovalDialog.vue'
import { getFieldSamplingList, getFieldSampling, createFieldSampling, updateFieldSampling, deleteFieldSampling, approveFieldSampling, rejectFieldSampling, listSamplingSheets } from '@/api/business'
import type { FieldSamplingRecord, SamplingSheet } from '@/api/business'

const router = useRouter()

interface FieldSamplingRow extends FieldSamplingRecord {
  sampling_sheet_id?: number | null
}

const loading = ref(false)
const items = ref<FieldSamplingRow[]>([])
const taskOrderIdFilter = ref('')

onMounted(async () => { await loadData() })

async function loadData() {
  loading.value = true
  try {
    const params: Record<string, string> = {}
    if (taskOrderIdFilter.value) params.task_order_id = taskOrderIdFilter.value
    const res = await getFieldSamplingList(params)
    const list: FieldSamplingRow[] = res.data || []

    const sheetRes = await listSamplingSheets({ node_code: 'node_field_sampling' })
    const sheets: SamplingSheet[] = sheetRes.data || []
    const sheetMap = new Map<number, SamplingSheet>()
    for (const s of sheets) sheetMap.set(s.task_order_id, s)

    for (const row of list) {
      const sheet = sheetMap.get(row.task_order_id)
      row.sampling_sheet_id = sheet?.id ?? null
    }

    items.value = list
  } finally { loading.value = false }
}

function openSheetEditor(row: FieldSamplingRow) {
  router.push({
    name: 'SamplingSheetEditorPage',
    query: {
      task_order_id: row.task_order_id,
      node_code: 'node_field_sampling',
    },
  })
}

const formVisible = ref(false); const isEdit = ref(false); const formRef = ref<FormInstance>(); const saving = ref(false)
const form = reactive({ id: 0, task_order_id: 0, sample_photos: '', equipment_cal_records: '', sampling_record_file_path: '' })
const formRules: FormRules = { task_order_id: [{ required: true, message: '请输入委托ID', trigger: 'blur' }] }
function openCreate() { isEdit.value = false; form.id = 0; form.task_order_id = 0; form.sample_photos = ''; form.equipment_cal_records = ''; form.sampling_record_file_path = ''; formVisible.value = true }
async function openEdit(row: FieldSamplingRecord) { isEdit.value = true; const res = await getFieldSampling(row.id); const d = res.data; form.id = d.id; form.task_order_id = d.task_order_id; form.sample_photos = d.sample_photos; form.equipment_cal_records = d.equipment_cal_records; form.sampling_record_file_path = d.sampling_record_file_path; formVisible.value = true }
async function handleSave() {
  const valid = await formRef.value?.validate().catch(() => false); if (!valid) return
  saving.value = true
  try {
    const data = { task_order_id: form.task_order_id, sample_photos: form.sample_photos, equipment_cal_records: form.equipment_cal_records, sampling_record_file_path: form.sampling_record_file_path }
    if (isEdit.value) { await updateFieldSampling(form.id, data); ElMessage.success('更新成功') }
    else { await createFieldSampling(data); ElMessage.success('创建成功') }
    formVisible.value = false; await loadData()
  } finally { saving.value = false }
}
async function handleDelete(id: number) { await deleteFieldSampling(id); ElMessage.success('删除成功'); await loadData() }

const approvalRef = ref<InstanceType<typeof ApprovalDialog>>()
const currentRow = ref<FieldSamplingRecord | null>(null)

function openApprove(row: FieldSamplingRecord) { currentRow.value = row; approvalRef.value?.open('approve') }
function openReject(row: FieldSamplingRecord) { currentRow.value = row; approvalRef.value?.open('reject') }

async function handleApprovalSubmit(data: { action: string; comment: string }) {
  if (!currentRow.value) return
  const id = currentRow.value.id
  const taskId = currentRow.value.task_order_id
  if (data.action === 'approve') {
    await approveFieldSampling(id, { task_id: taskId, comment: data.comment })
    ElMessage.success('采样通过，流程已推进')
  } else {
    await rejectFieldSampling(id, { task_id: taskId, comment: data.comment })
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