<template>
  <div class="equipment-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>仪器设备管理</span>
          <div>
            <el-input v-model="keyword" placeholder="搜索名称/编号" clearable style="width:200px;margin-right:8px" @clear="loadData" @keyup.enter="loadData" />
            <el-button type="primary" @click="openCreate">新增设备</el-button>
          </div>
        </div>
      </template>

      <el-table :data="list" stripe v-loading="loading">
        <el-table-column prop="id" label="ID" width="60" />
        <el-table-column prop="name" label="设备名称" min-width="160" />
        <el-table-column prop="code" label="编号" width="120" />
        <el-table-column prop="model" label="规格型号" min-width="140" />
        <el-table-column prop="factory" label="生产厂家" min-width="160" />
        <el-table-column prop="calibration_date" label="校准日期" width="120" />
        <el-table-column prop="next_cal_date" label="下次校准" width="120" />
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : row.status === 0 ? 'warning' : 'info'" size="small">
              {{ row.status === 1 ? '正常' : row.status === 0 ? '维护中' : '已报废' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="openEdit(row)">编辑</el-button>
            <el-popconfirm title="确认删除？" @confirm="handleDelete(row.id)">
              <template #reference>
                <el-button size="small" type="danger">删除</el-button>
              </template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- Form dialog -->
    <el-dialog v-model="formVisible" :title="isEdit ? '编辑设备' : '新增设备'" width="600px">
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="100px">
        <el-form-item label="设备名称" prop="name">
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="编号" prop="code">
          <el-input v-model="form.code" :disabled="isEdit" />
        </el-form-item>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="规格型号">
              <el-input v-model="form.model" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="生产厂家">
              <el-input v-model="form.factory" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="校准日期">
              <el-date-picker v-model="form.calibration_date" type="date" placeholder="选择日期" style="width:100%" value-format="YYYY-MM-DD" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="下次校准日期">
              <el-date-picker v-model="form.next_cal_date" type="date" placeholder="选择日期" style="width:100%" value-format="YYYY-MM-DD" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="状态">
          <el-select v-model="form.status" style="width:100%">
            <el-option :label="'正常'" :value="1" />
            <el-option :label="'维护中'" :value="0" />
            <el-option :label="'已报废'" :value="2" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="formVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSave">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { getEquipmentList, getEquipment, createEquipment, updateEquipment, deleteEquipment } from '@/api/baseData'
import type { Equipment } from '@/api/baseData'

const loading = ref(false)
const list = ref<Equipment[]>([])
const keyword = ref('')

onMounted(() => loadData())

async function loadData() {
  loading.value = true
  try {
    const params: Record<string, string> = {}
    if (keyword.value) params.keyword = keyword.value
    const res = await getEquipmentList(params)
    list.value = res.data
  } finally {
    loading.value = false
  }
}

// --- Form ---
const formVisible = ref(false)
const isEdit = ref(false)
const formRef = ref<FormInstance>()
const saving = ref(false)

const form = reactive({
  id: 0,
  name: '',
  code: '',
  model: '',
  factory: '',
  calibration_date: '',
  next_cal_date: '',
  status: 1,
})

const formRules: FormRules = {
  name: [{ required: true, message: '请输入设备名称', trigger: 'blur' }],
  code: [{ required: true, message: '请输入编号', trigger: 'blur' }],
}

function openCreate() {
  isEdit.value = false
  form.id = 0
  form.name = ''
  form.code = ''
  form.model = ''
  form.factory = ''
  form.calibration_date = ''
  form.next_cal_date = ''
  form.status = 1
  formVisible.value = true
}

async function openEdit(row: Equipment) {
  isEdit.value = true
  const res = await getEquipment(row.id)
  const d = res.data
  form.id = d.id
  form.name = d.name
  form.code = d.code
  form.model = d.model
  form.factory = d.factory
  form.calibration_date = d.calibration_date
  form.next_cal_date = d.next_cal_date
  form.status = d.status
  formVisible.value = true
}

async function handleSave() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  saving.value = true
  try {
    if (isEdit.value) {
      await updateEquipment(form.id, {
        name: form.name,
        model: form.model,
        factory: form.factory,
        calibration_date: form.calibration_date || undefined,
        next_cal_date: form.next_cal_date || undefined,
        status: form.status,
      })
      ElMessage.success('更新成功')
    } else {
      await createEquipment({
        name: form.name,
        code: form.code,
        model: form.model,
        factory: form.factory,
        calibration_date: form.calibration_date || undefined,
        next_cal_date: form.next_cal_date || undefined,
      })
      ElMessage.success('创建成功')
    }
    formVisible.value = false
    await loadData()
  } finally {
    saving.value = false
  }
}

async function handleDelete(id: number) {
  await deleteEquipment(id)
  ElMessage.success('删除成功')
  await loadData()
}
</script>