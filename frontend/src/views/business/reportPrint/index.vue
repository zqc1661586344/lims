<template>
  <div class="report-print-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>报告打印发放</span>
          <div>
            <el-input v-model="taskOrderIdFilter" placeholder="委托ID" clearable style="width:160px;margin-right:8px" @clear="loadData" @keyup.enter="loadData" />
            <el-button type="primary" @click="openCreate">新增打印</el-button>
          </div>
        </div>
      </template>

      <el-table :data="items" stripe v-loading="loading">
        <el-table-column prop="task_order_id" label="委托ID" width="80" />
        <el-table-column prop="print_count" label="打印份数" width="100" />
        <el-table-column prop="print_result" label="打印结果" width="100">
          <template #default="{ row }">
            <el-tag :type="row.print_result === '通过' ? 'success' : 'danger'" size="small">{{ row.print_result }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="delivery_method" label="领取方式" width="100">
          <template #default="{ row }">
            <el-tag size="small">{{ row.delivery_method || '-' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="recipient_name" label="领取人" width="120" />
        <el-table-column prop="recipient_date" label="领取日期" width="170" />
        <el-table-column prop="tracking_no" label="快递单号" width="150" show-overflow-tooltip />
        <el-table-column prop="print_comment" label="打印备注" min-width="180" show-overflow-tooltip />
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
    <el-dialog v-model="formVisible" :title="isEdit ? '编辑报告打印' : '新增报告打印'" width="600px">
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="120px">
        <el-form-item label="委托ID" prop="task_order_id">
          <el-input-number v-model="form.task_order_id" :min="1" style="width:100%" />
        </el-form-item>
        <el-form-item label="打印份数">
          <el-input-number v-model="form.print_count" :min="1" style="width:100%" />
        </el-form-item>
        <el-form-item label="打印结果">
          <el-select v-model="form.print_result" placeholder="选择结果" style="width:100%">
            <el-option label="通过" value="通过" />
            <el-option label="驳回" value="驳回" />
          </el-select>
        </el-form-item>
        <el-form-item label="领取方式">
          <el-select v-model="form.delivery_method" placeholder="选择方式" style="width:100%">
            <el-option label="自取" value="自取" />
            <el-option label="邮寄" value="邮寄" />
          </el-select>
        </el-form-item>
        <el-form-item label="领取人">
          <el-input v-model="form.recipient_name" placeholder="领取人姓名" />
        </el-form-item>
        <el-form-item label="领取日期">
          <el-date-picker v-model="form.recipient_date" type="date" placeholder="选择日期" style="width:100%" />
        </el-form-item>
        <el-form-item label="快递单号">
          <el-input v-model="form.tracking_no" placeholder="快递单号（邮寄时填写）" />
        </el-form-item>
        <el-form-item label="打印备注">
          <el-input v-model="form.print_comment" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="formVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSave">保存</el-button>
      </template>
    </el-dialog>

    <!-- Approve dialog -->
    <el-dialog v-model="approveVisible" title="报告打印发放 - 通过" width="500px">
      <el-form :model="approveForm" label-width="100px">
        <el-form-item label="委托ID">{{ approveForm.task_order_id }}</el-form-item>
        <el-form-item label="打印份数">
          <el-input-number v-model="approveForm.print_count" :min="1" style="width:100%" />
        </el-form-item>
        <el-form-item label="打印结果">
          <el-select v-model="approveForm.print_result" placeholder="选择结果" style="width:100%">
            <el-option label="通过" value="通过" />
            <el-option label="驳回" value="驳回" />
          </el-select>
        </el-form-item>
        <el-form-item label="领取方式">
          <el-select v-model="approveForm.delivery_method" placeholder="选择方式" style="width:100%">
            <el-option label="自取" value="自取" />
            <el-option label="邮寄" value="邮寄" />
          </el-select>
        </el-form-item>
        <el-form-item label="领取人">
          <el-input v-model="approveForm.recipient_name" placeholder="领取人姓名" />
        </el-form-item>
        <el-form-item label="领取日期">
          <el-date-picker v-model="approveForm.recipient_date" type="date" placeholder="选择日期" style="width:100%" />
        </el-form-item>
        <el-form-item label="快递单号">
          <el-input v-model="approveForm.tracking_no" placeholder="快递单号" />
        </el-form-item>
        <el-form-item label="打印备注">
          <el-input v-model="approveForm.comment" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="approveVisible = false">取消</el-button>
        <el-button type="primary" :loading="approving" @click="handleApprove">确认通过</el-button>
      </template>
    </el-dialog>

    <!-- Reject dialog -->
    <el-dialog v-model="rejectVisible" title="报告打印发放 - 驳回" width="500px">
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
import { getReportPrintList, getReportPrint, createReportPrint, updateReportPrint, deleteReportPrint, approveReportPrint, rejectReportPrint } from '@/api/business'
import type { ReportPrint } from '@/api/business'

const loading = ref(false)
const items = ref<ReportPrint[]>([])
const taskOrderIdFilter = ref('')

onMounted(async () => { await loadData() })

async function loadData() {
  loading.value = true
  try {
    const params: Record<string, string> = {}
    if (taskOrderIdFilter.value) params.task_order_id = taskOrderIdFilter.value
    const res = await getReportPrintList(params)
    items.value = res.data
  } finally { loading.value = false }
}

const formVisible = ref(false); const isEdit = ref(false); const formRef = ref<FormInstance>(); const saving = ref(false)
const form = reactive({ id: 0, task_order_id: 0, print_count: 1, print_result: '', delivery_method: '', recipient_name: '', recipient_date: '', tracking_no: '', print_comment: '' })
const formRules: FormRules = { task_order_id: [{ required: true, message: '请输入委托ID', trigger: 'blur' }] }
function openCreate() { isEdit.value = false; form.id = 0; form.task_order_id = 0; form.print_count = 1; form.print_result = ''; form.delivery_method = ''; form.recipient_name = ''; form.recipient_date = ''; form.tracking_no = ''; form.print_comment = ''; formVisible.value = true }
async function openEdit(row: ReportPrint) { isEdit.value = true; const res = await getReportPrint(row.id); const d = res.data; form.id = d.id; form.task_order_id = d.task_order_id; form.print_count = d.print_count || 1; form.print_result = d.print_result; form.delivery_method = d.delivery_method; form.recipient_name = d.recipient_name; form.recipient_date = d.recipient_date; form.tracking_no = d.tracking_no; form.print_comment = d.print_comment; formVisible.value = true }
async function handleSave() {
  const valid = await formRef.value?.validate().catch(() => false); if (!valid) return
  saving.value = true
  try {
    const data = { task_order_id: form.task_order_id, print_count: form.print_count, print_result: form.print_result, delivery_method: form.delivery_method, recipient_name: form.recipient_name, recipient_date: form.recipient_date, tracking_no: form.tracking_no, print_comment: form.print_comment }
    if (isEdit.value) { await updateReportPrint(form.id, data); ElMessage.success('更新成功') }
    else { await createReportPrint(data); ElMessage.success('创建成功') }
    formVisible.value = false; await loadData()
  } finally { saving.value = false }
}
async function handleDelete(id: number) { await deleteReportPrint(id); ElMessage.success('删除成功'); await loadData() }

const approveVisible = ref(false); const approving = ref(false)
const approveForm = reactive({ id: 0, task_order_id: 0, print_count: 1, print_result: '通过', delivery_method: '', recipient_name: '', recipient_date: '', tracking_no: '', comment: '' })
function openApprove(row: ReportPrint) { approveForm.id = row.id; approveForm.task_order_id = row.task_order_id; approveForm.print_count = row.print_count || 1; approveForm.print_result = '通过'; approveForm.delivery_method = row.delivery_method; approveForm.recipient_name = row.recipient_name; approveForm.recipient_date = row.recipient_date; approveForm.tracking_no = row.tracking_no; approveForm.comment = ''; approveVisible.value = true }
async function handleApprove() {
  approving.value = true
  try { await approveReportPrint(approveForm.id, { task_id: approveForm.task_order_id, print_count: approveForm.print_count, print_result: approveForm.print_result, delivery_method: approveForm.delivery_method, recipient_name: approveForm.recipient_name, recipient_date: approveForm.recipient_date, tracking_no: approveForm.tracking_no, print_comment: approveForm.comment }); ElMessage.success('打印发放通过，流程已推进'); approveVisible.value = false; await loadData() }
  catch (e: any) { ElMessage.error(e?.response?.data?.message || '操作失败') }
  finally { approving.value = false }
}

const rejectVisible = ref(false); const rejectFormRef = ref<FormInstance>(); const rejecting = ref(false)
const rejectForm = reactive({ id: 0, task_order_id: 0, comment: '' })
const rejectRules: FormRules = { comment: [{ required: true, message: '请填写驳回原因', trigger: 'blur' }] }
function openReject(row: ReportPrint) { rejectForm.id = row.id; rejectForm.task_order_id = row.task_order_id; rejectForm.comment = ''; rejectVisible.value = true }
async function handleReject() {
  const valid = await rejectFormRef.value?.validate().catch(() => false); if (!valid) return
  rejecting.value = true
  try { await rejectReportPrint(rejectForm.id, { task_id: rejectForm.task_order_id, comment: rejectForm.comment }); ElMessage.success('已驳回'); rejectVisible.value = false; await loadData() }
  catch (e: any) { ElMessage.error(e?.response?.data?.message || '操作失败') }
  finally { rejecting.value = false }
}
</script>
