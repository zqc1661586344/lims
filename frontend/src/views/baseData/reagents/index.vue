<template>
  <div class="reagents-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>物资管理</span>
          <div>
            <el-input v-model="keyword" placeholder="搜索名称/编号" clearable style="width:200px;margin-right:8px" @clear="loadData" @keyup.enter="loadData" />
            <el-button type="primary" @click="openCreate">新增物资</el-button>
          </div>
        </div>
      </template>

      <el-table :data="list" stripe v-loading="loading">
        <el-table-column prop="id" label="ID" width="60" />
        <el-table-column prop="name" label="名称" min-width="140" />
        <el-table-column prop="code" label="编号" width="120" />
        <el-table-column prop="spec" label="规格" width="100" />
        <el-table-column prop="manufacturer" label="生产厂商" min-width="160" />
        <el-table-column prop="batch_no" label="批号" width="120" />
        <el-table-column prop="stock_qty" label="库存量" width="100">
          <template #default="{ row }">{{ row.stock_qty }} {{ row.unit }}</template>
        </el-table-column>
        <el-table-column prop="expire_date" label="有效期" width="100" />
        <el-table-column label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">
              {{ row.status === 1 ? '启用' : '禁用' }}
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
    <el-dialog v-model="formVisible" :title="isEdit ? '编辑试剂' : '新增试剂'" width="600px">
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="100px">
        <el-form-item label="试剂名称" prop="name">
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="编号" prop="code">
          <el-input v-model="form.code" :disabled="isEdit" />
        </el-form-item>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="规格">
              <el-input v-model="form.spec" placeholder="500mL, AR级" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="单位">
              <el-input v-model="form.unit" placeholder="瓶, mL, g" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="生产厂商">
          <el-input v-model="form.manufacturer" />
        </el-form-item>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="批号">
              <el-input v-model="form.batch_no" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="库存量">
              <el-input-number v-model="form.stock_qty" :min="0" :precision="2" style="width:100%" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="有效期">
          <el-date-picker v-model="form.expire_date" type="date" placeholder="选择日期" style="width:100%" value-format="YYYY-MM-DD" />
        </el-form-item>
        <el-form-item label="状态">
          <el-switch v-model="form.status" :active-value="1" :inactive-value="0" />
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
import { getReagents, getReagent, createReagent, updateReagent, deleteReagent } from '@/api/baseData'
import type { Reagent } from '@/api/baseData'

const loading = ref(false)
const list = ref<Reagent[]>([])
const keyword = ref('')

onMounted(() => loadData())

async function loadData() {
  loading.value = true
  try {
    const params: Record<string, string> = {}
    if (keyword.value) params.keyword = keyword.value
    const res = await getReagents(params)
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
  spec: '',
  manufacturer: '',
  batch_no: '',
  stock_qty: 0,
  unit: '',
  expire_date: '',
  status: 1,
})

const formRules: FormRules = {
  name: [{ required: true, message: '请输入试剂名称', trigger: 'blur' }],
  code: [{ required: true, message: '请输入编号', trigger: 'blur' }],
}

function openCreate() {
  isEdit.value = false
  form.id = 0
  form.name = ''
  form.code = ''
  form.spec = ''
  form.manufacturer = ''
  form.batch_no = ''
  form.stock_qty = 0
  form.unit = ''
  form.expire_date = ''
  form.status = 1
  formVisible.value = true
}

async function openEdit(row: Reagent) {
  isEdit.value = true
  const res = await getReagent(row.id)
  const d = res.data
  form.id = d.id
  form.name = d.name
  form.code = d.code
  form.spec = d.spec
  form.manufacturer = d.manufacturer
  form.batch_no = d.batch_no
  form.stock_qty = d.stock_qty
  form.unit = d.unit
  form.expire_date = d.expire_date
  form.status = d.status
  formVisible.value = true
}

async function handleSave() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  saving.value = true
  try {
    if (isEdit.value) {
      await updateReagent(form.id, {
        name: form.name,
        spec: form.spec,
        manufacturer: form.manufacturer,
        batch_no: form.batch_no,
        stock_qty: form.stock_qty,
        unit: form.unit,
        expire_date: form.expire_date || undefined,
        status: form.status,
      })
      ElMessage.success('更新成功')
    } else {
      await createReagent({
        name: form.name,
        code: form.code,
        spec: form.spec,
        manufacturer: form.manufacturer,
        batch_no: form.batch_no,
        stock_qty: form.stock_qty,
        unit: form.unit,
        expire_date: form.expire_date || undefined,
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
  await deleteReagent(id)
  ElMessage.success('删除成功')
  await loadData()
}
</script>