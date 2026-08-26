<template>
  <div class="sampling-schedule-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>采样调度管理</span>
          <div>
            <el-input v-model="taskOrderIdFilter" placeholder="委托ID" clearable style="width:160px;margin-right:8px" @clear="loadData" @keyup.enter="loadData" />
            <el-button type="primary" @click="openCreate">新增采样调度</el-button>
          </div>
        </div>
      </template>

      <el-table :data="items" stripe v-loading="loading">
        <el-table-column prop="task_order_id" label="委托ID" width="80" />
        <el-table-column prop="sampling_team" label="采样团队" width="150" />
        <el-table-column prop="sampling_points" label="采样点位" min-width="180" show-overflow-tooltip />
        <el-table-column prop="equipment_list" label="设备清单" min-width="180" show-overflow-tooltip />
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
    <el-dialog v-model="formVisible" :title="isEdit ? '编辑采样调度' : '新增采样调度'" width="600px">
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="100px">
        <el-form-item label="委托ID" prop="task_order_id">
          <el-input-number v-model="form.task_order_id" :min="1" style="width:100%" />
        </el-form-item>
        <el-form-item label="采样团队">
          <el-input v-model="form.sampling_team" />
        </el-form-item>
        <el-form-item label="采样点位">
          <el-input v-model="form.sampling_points" type="textarea" :rows="2" placeholder="JSON: [{location,type,count}]" />
        </el-form-item>
        <el-form-item label="设备清单">
          <el-input v-model="form.equipment_list" type="textarea" :rows="2" placeholder="JSON: [{equipment_id,name,model}]" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="formVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSave">保存</el-button>
      </template>
    </el-dialog>

    <!-- Approve dialog -->
    <el-dialog v-model="approveVisible" title="采样调度 - 通过" width="500px">
      <el-form :model="approveForm" label-width="100px">
        <el-form-item label="委托ID">{{ approveForm.task_order_id }}</el-form-item>
        <el-form-item label="审批意见">
          <el-input v-model="approveForm.comment" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="approveVisible = false">取消</el-button>
        <el-button type="primary" :loading="approving" @click="handleApprove">确认通过</el-button>
      </template>
    </el-dialog>

    <!-- Reject dialog -->
    <el-dialog v-model="rejectVisible" title="采样调度 - 驳回" width="500px">
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
import { getSamplingScheduleList, getSamplingSchedule, createSamplingSchedule, updateSamplingSchedule, deleteSamplingSchedule, approveSamplingSchedule, rejectSamplingSchedule } from '@/api/business'
import type { SamplingSchedule } from '@/api/business'

const loading = ref(false)
const items = ref<SamplingSchedule[]>([])
const taskOrderIdFilter = ref('')

onMounted(async () => { await loadData() })

async function loadData() {
  loading.value = true
  try {
    const params: Record<string, string> = {}
    if (taskOrderIdFilter.value) params.task_order_id = taskOrderIdFilter.value
    const res = await getSamplingScheduleList(params)
    items.value = res.data
  } finally { loading.value = false }
}

const formVisible = ref(false)
const isEdit = ref(false)
const formRef = ref<FormInstance>()
const saving = ref(false)
const form = reactive({ id: 0, task_order_id: 0, sampling_team: '', sampling_points: '', equipment_list: '' })
const formRules: FormRules = { task_order_id: [{ required: true, message: '请输入委托ID', trigger: 'blur' }] }
function openCreate() { isEdit.value = false; form.id = 0; form.task_order_id = 0; form.sampling_team = ''; form.sampling_points = ''; form.equipment_list = ''; formVisible.value = true }
async function openEdit(row: SamplingSchedule) {
  isEdit.value = true; const res = await getSamplingSchedule(row.id); const d = res.data
  form.id = d.id; form.task_order_id = d.task_order_id; form.sampling_team = d.sampling_team; form.sampling_points = d.sampling_points; form.equipment_list = d.equipment_list
  formVisible.value = true
}
async function handleSave() {
  const valid = await formRef.value?.validate().catch(() => false); if (!valid) return
  saving.value = true
  try {
    const data = { task_order_id: form.task_order_id, sampling_team: form.sampling_team, sampling_points: form.sampling_points, equipment_list: form.equipment_list }
    if (isEdit.value) { await updateSamplingSchedule(form.id, data); ElMessage.success('更新成功') }
    else { await createSamplingSchedule(data); ElMessage.success('创建成功') }
    formVisible.value = false; await loadData()
  } finally { saving.value = false }
}
async function handleDelete(id: number) { await deleteSamplingSchedule(id); ElMessage.success('删除成功'); await loadData() }

const approveVisible = ref(false); const approving = ref(false)
const approveForm = reactive({ id: 0, task_order_id: 0, comment: '' })
function openApprove(row: SamplingSchedule) { approveForm.id = row.id; approveForm.task_order_id = row.task_order_id; approveForm.comment = ''; approveVisible.value = true }
async function handleApprove() {
  approving.value = true
  try { await approveSamplingSchedule(approveForm.id, { task_id: approveForm.task_order_id, comment: approveForm.comment }); ElMessage.success('调度通过，流程已推进'); approveVisible.value = false; await loadData() }
  catch (e: any) { ElMessage.error(e?.response?.data?.message || '操作失败') }
  finally { approving.value = false }
}

const rejectVisible = ref(false); const rejectFormRef = ref<FormInstance>(); const rejecting = ref(false)
const rejectForm = reactive({ id: 0, task_order_id: 0, comment: '' })
const rejectRules: FormRules = { comment: [{ required: true, message: '请填写驳回原因', trigger: 'blur' }] }
function openReject(row: SamplingSchedule) { rejectForm.id = row.id; rejectForm.task_order_id = row.task_order_id; rejectForm.comment = ''; rejectVisible.value = true }
async function handleReject() {
  const valid = await rejectFormRef.value?.validate().catch(() => false); if (!valid) return
  rejecting.value = true
  try { await rejectSamplingSchedule(rejectForm.id, { task_id: rejectForm.task_order_id, comment: rejectForm.comment }); ElMessage.success('已驳回'); rejectVisible.value = false; await loadData() }
  catch (e: any) { ElMessage.error(e?.response?.data?.message || '操作失败') }
  finally { rejecting.value = false }
}
</script>