<template>
  <div class="user-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>用户管理</span>
          <el-button type="primary" @click="openCreate">新增用户</el-button>
        </div>
      </template>

      <el-table :data="users" stripe v-loading="loading">
        <el-table-column prop="id" label="ID" width="60" />
        <el-table-column prop="username" label="用户名" min-width="120" />
        <el-table-column prop="real_name" label="姓名" min-width="100" />
        <el-table-column prop="email" label="邮箱" min-width="160" />
        <el-table-column prop="phone" label="电话" width="120" />
        <el-table-column label="部门" width="100">
          <template #default="{ row }">{{ row.dept?.name || '-' }}</template>
        </el-table-column>
        <el-table-column label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="is_admin" label="管理员" width="80">
          <template #default="{ row }">
            <el-tag :type="row.is_admin ? 'warning' : 'info'" size="small">{{ row.is_admin ? '是' : '否' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="240" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="openEdit(row)">编辑</el-button>
            <el-button size="small" @click="openRoles(row)">角色</el-button>
            <el-popconfirm title="确认删除？" @confirm="handleDelete(row.id)">
              <template #reference>
                <el-button size="small" type="danger">删除</el-button>
              </template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- User form dialog -->
    <el-dialog v-model="formVisible" :title="isEdit ? '编辑用户' : '新增用户'" width="500px">
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="80px">
        <el-form-item label="用户名" prop="username">
          <el-input v-model="form.username" :disabled="isEdit" />
        </el-form-item>
        <el-form-item label="密码" prop="password" v-if="!isEdit">
          <el-input v-model="form.password" type="password" show-password />
        </el-form-item>
        <el-form-item label="姓名" prop="real_name">
          <el-input v-model="form.real_name" />
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="form.email" />
        </el-form-item>
        <el-form-item label="电话" prop="phone">
          <el-input v-model="form.phone" />
        </el-form-item>
        <el-form-item label="部门">
          <el-select v-model="form.dept_id" placeholder="选择部门" clearable style="width:100%">
            <el-option v-for="d in depts" :key="d.id" :label="d.name" :value="d.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-switch v-model="form.status" :active-value="1" :inactive-value="0" />
        </el-form-item>
        <el-form-item label="管理员">
          <el-switch v-model="form.is_admin" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="formVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSave">保存</el-button>
      </template>
    </el-dialog>

    <!-- Role assignment dialog -->
    <el-dialog v-model="roleVisible" title="分配角色" width="400px">
      <el-checkbox-group v-model="selectedRoles">
        <el-checkbox v-for="r in roles" :key="r.id" :label="r.id" :value="r.id">{{ r.name }}</el-checkbox>
      </el-checkbox-group>
      <template #footer>
        <el-button @click="roleVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSaveRoles">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { getUsers, getUser, createUser, updateUser, deleteUser, updateUserRoles, getDepts, getRoles } from '@/api/system'
import type { User, Dept, Role } from '@/api/system'

const loading = ref(false)
const users = ref<User[]>([])
const depts = ref<Dept[]>([])
const roles = ref<Role[]>([])

onMounted(async () => {
  await loadUsers()
  const [deptRes, roleRes] = await Promise.all([getDepts(), getRoles()])
  depts.value = deptRes.data
  roles.value = roleRes.data
})

async function loadUsers() {
  loading.value = true
  try {
    const res = await getUsers()
    users.value = res.data?.items ?? res.data
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
  username: '',
  password: '',
  real_name: '',
  email: '',
  phone: '',
  dept_id: null as number | null,
  status: 1,
  is_admin: false,
})

const formRules: FormRules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
}

function openCreate() {
  isEdit.value = false
  form.id = 0
  form.username = ''
  form.password = ''
  form.real_name = ''
  form.email = ''
  form.phone = ''
  form.dept_id = null
  form.status = 1
  form.is_admin = false
  formVisible.value = true
}

async function openEdit(row: User) {
  isEdit.value = true
  const res = await getUser(row.id)
  const u = res.data
  form.id = u.id
  form.username = u.username
  form.password = ''
  form.real_name = u.real_name
  form.email = u.email
  form.phone = u.phone
  form.dept_id = u.dept_id
  form.status = u.status
  form.is_admin = u.is_admin
  formVisible.value = true
}

async function handleSave() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  saving.value = true
  try {
    if (isEdit.value) {
      await updateUser(form.id, {
        real_name: form.real_name,
        email: form.email,
        phone: form.phone,
        dept_id: form.dept_id,
        status: form.status,
        is_admin: form.is_admin,
      })
      ElMessage.success('更新成功')
    } else {
      await createUser({
        username: form.username,
        password: form.password,
        real_name: form.real_name,
        email: form.email,
        phone: form.phone,
        dept_id: form.dept_id,
      })
      ElMessage.success('创建成功')
    }
    formVisible.value = false
    await loadUsers()
  } finally {
    saving.value = false
  }
}

async function handleDelete(id: number) {
  await deleteUser(id)
  ElMessage.success('删除成功')
  await loadUsers()
}

// --- Role dialog ---
const roleVisible = ref(false)
const selectedRoles = ref<number[]>([])
const currentUserId = ref(0)

async function openRoles(row: User) {
  currentUserId.value = row.id
  selectedRoles.value = row.roles?.map(r => r.id) || []
  roleVisible.value = true
}

async function handleSaveRoles() {
  await updateUserRoles(currentUserId.value, selectedRoles.value)
  ElMessage.success('角色分配成功')
  roleVisible.value = false
  await loadUsers()
}
</script>