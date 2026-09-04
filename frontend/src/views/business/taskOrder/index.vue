<template>
  <div class="task-order-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>任务委托管理</span>
          <div>
            <el-input v-model="keyword" placeholder="搜索单号/客户/项目" clearable style="width:220px;margin-right:8px" @clear="loadData" @keyup.enter="loadData" />
            <el-button type="primary" @click="openCreate">新增委托</el-button>
          </div>
        </div>
      </template>

      <el-table :data="items" stripe v-loading="loading">
        <el-table-column prop="order_no" label="委托单号" width="160" />
        <el-table-column prop="customer_name" label="客户名称" min-width="140" />
        <el-table-column prop="project_name" label="项目名称" min-width="160" />
        <el-table-column prop="sample_type" label="样品类型" width="100" />
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag v-if="row.status === 0" type="info" size="small">草稿</el-tag>
            <el-tag v-else-if="row.status === 1" type="warning" size="small">已提交</el-tag>
            <el-tag v-else-if="row.status === 2" type="primary" size="small">流程中</el-tag>
            <el-tag v-else type="success" size="small">已完成</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="170" />
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="openEdit(row)">编辑</el-button>
            <el-button v-if="row.status === 0" size="small" type="primary" @click="handleSubmit(row)">提交</el-button>
            <el-popconfirm v-if="row.status === 0" title="确认删除？" @confirm="handleDelete(row.id)">
              <template #reference>
                <el-button size="small" type="danger">删除</el-button>
              </template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- Form dialog -->
    <el-dialog v-model="formVisible" :title="isEdit ? '编辑委托' : '新增委托'" width="650px">
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="100px">
        <el-form-item label="委托单号" prop="order_no">
          <el-input v-model="form.order_no" :disabled="isEdit" />
        </el-form-item>
        <el-form-item label="客户名称" prop="customer_name">
          <el-input v-model="form.customer_name" />
        </el-form-item>
        <el-form-item label="项目名称" prop="project_name">
          <el-input v-model="form.project_name" />
        </el-form-item>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="样品类型">
              <el-input v-model="form.sample_type" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="检测项目">
              <el-select v-model="form.test_items" multiple filterable placeholder="选择检测项目" style="width:100%">
                <el-option v-for="item in testItems" :key="item.id" :label="item.name" :value="JSON.stringify({test_item_id:item.id,name:item.name})" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
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
import { getTaskOrderList, getTaskOrder, createTaskOrder, updateTaskOrder, deleteTaskOrder, submitTaskOrder } from '@/api/business'
import type { TaskOrder } from '@/api/business'
import { getTestItems, type TestItem } from '@/api/baseData'

const loading = ref(false)
const items = ref<TaskOrder[]>([])
const keyword = ref('')
const testItems = ref<TestItem[]>([])

onMounted(async () => {
  await loadData()
  try {
    const res = await getTestItems()
    testItems.value = res.data
  } catch {
    testItems.value = []
  }
})

async function loadData() {
  loading.value = true
  try {
    const params: Record<string, string> = {}
    if (keyword.value) params.keyword = keyword.value
    const res = await getTaskOrderList(params)
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
  order_no: '',
  customer_name: '',
  project_name: '',
  sample_type: '',
  test_items: [] as string[],
})

const formRules: FormRules = {
  order_no: [{ required: true, message: '请输入委托单号', trigger: 'blur' }],
  customer_name: [{ required: true, message: '请输入客户名称', trigger: 'blur' }],
  project_name: [{ required: true, message: '请输入项目名称', trigger: 'blur' }],
}

function openCreate() {
  isEdit.value = false
  form.id = 0
  form.order_no = ''
  form.customer_name = ''
  form.project_name = ''
  form.sample_type = ''
  form.test_items = []
  formVisible.value = true
}

async function openEdit(row: TaskOrder) {
  if (row.status !== 0) {
    ElMessage.warning('已提交的委托不可编辑')
    return
  }
  isEdit.value = true
  const res = await getTaskOrder(row.id)
  const d = res.data
  form.id = d.id
  form.order_no = d.order_no
  form.customer_name = d.customer_name
  form.project_name = d.project_name
  form.sample_type = d.sample_type
  form.test_items = d.test_items ? JSON.parse(d.test_items) : []
  formVisible.value = true
}

async function handleSave() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  saving.value = true
  try {
    const data = {
      order_no: form.order_no,
      customer_name: form.customer_name,
      project_name: form.project_name,
      sample_type: form.sample_type,
      test_items: JSON.stringify(form.test_items),
    }
    console.log('[taskOrder handleSave] form.test_items =', form.test_items)
    console.log('[taskOrder handleSave] data.test_items =', data.test_items)
    if (isEdit.value) {
      await updateTaskOrder(form.id, data)
      ElMessage.success('更新成功')
    } else {
      await createTaskOrder(data)
      ElMessage.success('创建成功')
    }
    formVisible.value = false
    await loadData()
  } finally {
    saving.value = false
  }
}

async function handleSubmit(row: TaskOrder) {
  await submitTaskOrder(row.id)
  ElMessage.success('提交成功，流程已启动')
  await loadData()
}

async function handleDelete(id: number) {
  await deleteTaskOrder(id)
  ElMessage.success('删除成功')
  await loadData()
}
</script>