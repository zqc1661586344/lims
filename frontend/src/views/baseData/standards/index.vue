<template>
  <div class="standards-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>检测标准管理</span>
          <div>
            <el-input v-model="keyword" placeholder="搜索名称/编号" clearable style="width:200px;margin-right:8px" @clear="loadData" @keyup.enter="loadData" />
            <el-button type="primary" @click="openCreate">新增标准</el-button>
          </div>
        </div>
      </template>

      <el-table :data="standards" stripe v-loading="loading">
        <el-table-column prop="id" label="ID" width="60" />
        <el-table-column prop="name" label="标准名称" min-width="200" />
        <el-table-column prop="code" label="编号" width="120" />
        <el-table-column prop="issuer" label="发布机构" min-width="140" />
        <el-table-column prop="version" label="版本" width="70" />
        <el-table-column prop="publish_date" label="发布日期" width="120" />
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
    <el-dialog v-model="formVisible" :title="isEdit ? '编辑检测标准' : '新增检测标准'" width="600px">
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="100px">
        <el-form-item label="标准名称" prop="name">
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="编号" prop="code">
          <el-input v-model="form.code" :disabled="isEdit" />
        </el-form-item>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="发布机构">
              <el-input v-model="form.issuer" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="版本号">
              <el-input v-model="form.version" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="发布日期">
          <el-date-picker v-model="form.publish_date" type="date" placeholder="选择日期" style="width:100%" value-format="YYYY-MM-DD" />
        </el-form-item>
        <el-form-item label="文件路径">
          <el-input v-model="form.file_path" placeholder="MinIO文件路径" />
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
import { getTestStandards, getTestStandard, createTestStandard, updateTestStandard, deleteTestStandard } from '@/api/baseData'
import type { TestStandard } from '@/api/baseData'

const loading = ref(false)
const standards = ref<TestStandard[]>([])
const keyword = ref('')

onMounted(() => loadData())

async function loadData() {
  loading.value = true
  try {
    const params: Record<string, string> = {}
    if (keyword.value) params.keyword = keyword.value
    const res = await getTestStandards(params)
    standards.value = res.data
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
  issuer: '',
  version: '',
  publish_date: '',
  file_path: '',
  status: 1,
})

const formRules: FormRules = {
  name: [{ required: true, message: '请输入标准名称', trigger: 'blur' }],
  code: [{ required: true, message: '请输入编号', trigger: 'blur' }],
}

function openCreate() {
  isEdit.value = false
  form.id = 0
  form.name = ''
  form.code = ''
  form.issuer = ''
  form.version = ''
  form.publish_date = ''
  form.file_path = ''
  form.status = 1
  formVisible.value = true
}

async function openEdit(row: TestStandard) {
  isEdit.value = true
  const res = await getTestStandard(row.id)
  const d = res.data
  form.id = d.id
  form.name = d.name
  form.code = d.code
  form.issuer = d.issuer
  form.version = d.version
  form.publish_date = d.publish_date
  form.file_path = d.file_path
  form.status = d.status
  formVisible.value = true
}

async function handleSave() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  saving.value = true
  try {
    if (isEdit.value) {
      await updateTestStandard(form.id, {
        name: form.name,
        issuer: form.issuer,
        version: form.version,
        publish_date: form.publish_date || undefined,
        file_path: form.file_path,
        status: form.status,
      })
      ElMessage.success('更新成功')
    } else {
      await createTestStandard({
        name: form.name,
        code: form.code,
        issuer: form.issuer,
        version: form.version,
        publish_date: form.publish_date || undefined,
        file_path: form.file_path,
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
  await deleteTestStandard(id)
  ElMessage.success('删除成功')
  await loadData()
}
</script>