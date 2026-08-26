<template>
  <div class="items-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>检测项目管理</span>
          <div>
            <el-input v-model="keyword" placeholder="搜索名称/编号" clearable style="width:200px;margin-right:8px" @clear="loadData" @keyup.enter="loadData" />
            <el-select v-model="categoryFilter" placeholder="分类筛选" clearable style="width:140px;margin-right:8px" @change="loadData">
              <el-option label="化学" value="化学" />
              <el-option label="物理" value="物理" />
              <el-option label="微生物" value="微生物" />
              <el-option label="力学" value="力学" />
            </el-select>
            <el-button type="primary" @click="openCreate">新增项目</el-button>
          </div>
        </div>
      </template>

      <el-table :data="items" stripe v-loading="loading">
        <el-table-column prop="id" label="ID" width="60" />
        <el-table-column prop="name" label="项目名称" min-width="160" />
        <el-table-column prop="code" label="编号" width="120" />
        <el-table-column prop="category" label="分类" width="80" />
        <el-table-column prop="unit" label="单位" width="60" />
        <el-table-column label="检测标准" min-width="140">
          <template #default="{ row }">{{ row.standard?.name || '-' }}</template>
        </el-table-column>
        <el-table-column prop="price" label="价格(元)" width="100">
          <template #default="{ row }">{{ row.price?.toFixed(2) ?? '0.00' }}</template>
        </el-table-column>
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
    <el-dialog v-model="formVisible" :title="isEdit ? '编辑检测项目' : '新增检测项目'" width="600px">
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="100px">
        <el-form-item label="项目名称" prop="name">
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="编号" prop="code">
          <el-input v-model="form.code" :disabled="isEdit" />
        </el-form-item>
        <el-form-item label="分类" prop="category">
          <el-select v-model="form.category" placeholder="选择分类" clearable style="width:100%">
            <el-option label="化学" value="化学" />
            <el-option label="物理" value="物理" />
            <el-option label="微生物" value="微生物" />
            <el-option label="力学" value="力学" />
            <el-option label="其他" value="其他" />
          </el-select>
        </el-form-item>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="单位" prop="unit">
              <el-input v-model="form.unit" placeholder="mg/L, %等" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="价格(元)" prop="price">
              <el-input-number v-model="form.price" :min="0" :precision="2" style="width:100%" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="检测方法" prop="method">
          <el-input v-model="form.method" type="textarea" :rows="2" />
        </el-form-item>
        <el-form-item label="关联标准">
          <el-select v-model="form.standard_id" placeholder="选择检测标准" clearable filterable style="width:100%">
            <el-option v-for="s in standards" :key="s.id" :label="s.name" :value="s.id" />
          </el-select>
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
import { getTestItems, getTestItem, createTestItem, updateTestItem, deleteTestItem, getTestStandards } from '@/api/baseData'
import type { TestItem, TestStandard } from '@/api/baseData'

const loading = ref(false)
const items = ref<TestItem[]>([])
const standards = ref<TestStandard[]>([])
const keyword = ref('')
const categoryFilter = ref('')

onMounted(async () => {
  await loadData()
  const res = await getTestStandards()
  standards.value = res.data
})

async function loadData() {
  loading.value = true
  try {
    const params: Record<string, string> = {}
    if (keyword.value) params.keyword = keyword.value
    if (categoryFilter.value) params.category = categoryFilter.value
    const res = await getTestItems(params)
    items.value = res.data
  } finally {
    loading.value = false
  }
}

// --- Form dialog ---
const formVisible = ref(false)
const isEdit = ref(false)
const formRef = ref<FormInstance>()
const saving = ref(false)

const form = reactive({
  id: 0,
  name: '',
  code: '',
  category: '',
  unit: '',
  method: '',
  standard_id: null as number | null,
  price: 0,
  status: 1,
})

const formRules: FormRules = {
  name: [{ required: true, message: '请输入项目名称', trigger: 'blur' }],
  code: [{ required: true, message: '请输入编号', trigger: 'blur' }],
}

function openCreate() {
  isEdit.value = false
  form.id = 0
  form.name = ''
  form.code = ''
  form.category = ''
  form.unit = ''
  form.method = ''
  form.standard_id = null
  form.price = 0
  form.status = 1
  formVisible.value = true
}

async function openEdit(row: TestItem) {
  isEdit.value = true
  const res = await getTestItem(row.id)
  const d = res.data
  form.id = d.id
  form.name = d.name
  form.code = d.code
  form.category = d.category
  form.unit = d.unit
  form.method = d.method
  form.standard_id = d.standard_id
  form.price = d.price
  form.status = d.status
  formVisible.value = true
}

async function handleSave() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  saving.value = true
  try {
    if (isEdit.value) {
      await updateTestItem(form.id, {
        name: form.name,
        category: form.category,
        unit: form.unit,
        method: form.method,
        standard_id: form.standard_id,
        price: form.price,
        status: form.status,
      })
      ElMessage.success('更新成功')
    } else {
      await createTestItem({
        name: form.name,
        code: form.code,
        category: form.category,
        unit: form.unit,
        method: form.method,
        standard_id: form.standard_id,
        price: form.price,
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
  await deleteTestItem(id)
  ElMessage.success('删除成功')
  await loadData()
}
</script>