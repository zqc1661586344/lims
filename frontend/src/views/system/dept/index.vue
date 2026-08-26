<template>
  <div class="dept-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>部门管理</span>
          <el-button type="primary" @click="openCreate">新增部门</el-button>
        </div>
      </template>

      <el-table :data="depts" stripe row-key="id" v-loading="loading" default-expand-all>
        <el-table-column prop="name" label="部门名称" min-width="150" />
        <el-table-column prop="code" label="编码" width="150" />
        <el-table-column prop="sort" label="排序" width="60" />
        <el-table-column label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="openEdit(row)">编辑</el-button>
            <el-button size="small" type="primary" @click="openCreate(row)">新增子部门</el-button>
            <el-popconfirm title="确认删除？" @confirm="handleDelete(row.id)">
              <template #reference>
                <el-button size="small" type="danger">删除</el-button>
              </template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- Dept form dialog -->
    <el-dialog v-model="formVisible" :title="isEdit ? '编辑部门' : '新增部门'" width="450px">
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="80px">
        <el-form-item label="上级部门">
          <el-tree-select
            v-model="form.parent_id"
            :data="depts"
            :props="{ label: 'name', value: 'id', children: 'children' }"
            placeholder="顶级部门"
            clearable
            check-strictly
            style="width:100%"
          />
        </el-form-item>
        <el-form-item label="部门名称" prop="name">
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="编码" prop="code">
          <el-input v-model="form.code" />
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="form.sort" :min="0" style="width:100%" />
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
import { getDepts, createDept, updateDept, deleteDept } from '@/api/system'
import type { Dept } from '@/api/system'

const loading = ref(false)
const depts = ref<Dept[]>([])

onMounted(async () => {
  await loadDepts()
})

async function loadDepts() {
  loading.value = true
  try {
    const res = await getDepts()
    depts.value = res.data
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
  parent_id: null as number | null,
  name: '',
  code: '',
  sort: 0,
  status: 1,
})

const formRules: FormRules = {
  name: [{ required: true, message: '请输入部门名称', trigger: 'blur' }],
  code: [{ required: true, message: '请输入部门编码', trigger: 'blur' }],
}

function openCreate(parent?: Dept) {
  isEdit.value = false
  form.id = 0
  form.parent_id = parent?.id || null
  form.name = ''
  form.code = ''
  form.sort = 0
  form.status = 1
  formVisible.value = true
}

function openEdit(row: Dept) {
  isEdit.value = true
  form.id = row.id
  form.parent_id = row.parent_id
  form.name = row.name
  form.code = row.code
  form.sort = row.sort
  form.status = row.status
  formVisible.value = true
}

async function handleSave() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  saving.value = true
  try {
    if (isEdit.value) {
      await updateDept(form.id, {
        parent_id: form.parent_id,
        name: form.name,
        code: form.code,
        sort: form.sort,
        status: form.status,
      })
      ElMessage.success('更新成功')
    } else {
      await createDept({
        name: form.name,
        code: form.code,
        sort: form.sort,
        parent_id: form.parent_id,
      })
      ElMessage.success('创建成功')
    }
    formVisible.value = false
    await loadDepts()
  } finally {
    saving.value = false
  }
}

async function handleDelete(id: number) {
  await deleteDept(id)
  ElMessage.success('删除成功')
  await loadDepts()
}
</script>