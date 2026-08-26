<template>
  <div class="role-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>角色管理</span>
          <el-button type="primary" @click="openCreate">新增角色</el-button>
        </div>
      </template>

      <el-table :data="roles" stripe v-loading="loading">
        <el-table-column prop="name" label="角色名称" min-width="150" />
        <el-table-column prop="code" label="编码" width="150" />
        <el-table-column prop="remark" label="备注" min-width="200" show-overflow-tooltip />
        <el-table-column label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="240" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="openEdit(row)">编辑</el-button>
            <el-button size="small" type="primary" @click="openPermissions(row)">权限</el-button>
            <el-popconfirm title="确认删除？" @confirm="handleDelete(row.id)">
              <template #reference>
                <el-button size="small" type="danger">删除</el-button>
              </template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- Role form dialog -->
    <el-dialog v-model="formVisible" :title="isEdit ? '编辑角色' : '新增角色'" width="450px">
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="80px">
        <el-form-item label="角色名称" prop="name">
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="编码" prop="code">
          <el-input v-model="form.code" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="form.remark" type="textarea" :rows="2" />
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

    <!-- Permission assignment dialog -->
    <el-dialog v-model="permVisible" title="分配权限" width="400px">
      <el-tree
        ref="treeRef"
        :data="permissions"
        node-key="id"
        show-checkbox
        default-expand-all
        :props="{ label: 'name', children: 'children' }"
      />
      <template #footer>
        <el-button @click="permVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSavePermissions">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { nextTick, onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import type { ElTree } from 'element-plus'
import { getRoles, getRole, createRole, updateRole, deleteRole, getPermissions, updateRolePermissions } from '@/api/system'
import type { Role, PermissionNode } from '@/api/system'

const loading = ref(false)
const roles = ref<Role[]>([])
const permissions = ref<PermissionNode[]>([])

onMounted(async () => {
  await Promise.all([loadRoles(), loadPermissions()])
})

async function loadRoles() {
  loading.value = true
  try {
    const res = await getRoles()
    roles.value = res.data
  } finally {
    loading.value = false
  }
}

async function loadPermissions() {
  const res = await getPermissions()
  permissions.value = res.data
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
  remark: '',
  status: 1,
})

const formRules: FormRules = {
  name: [{ required: true, message: '请输入角色名称', trigger: 'blur' }],
  code: [{ required: true, message: '请输入角色编码', trigger: 'blur' }],
}

function openCreate() {
  isEdit.value = false
  form.id = 0
  form.name = ''
  form.code = ''
  form.remark = ''
  form.status = 1
  formVisible.value = true
}

async function openEdit(row: Role) {
  isEdit.value = true
  const res = await getRole(row.id)
  const r = res.data
  form.id = r.id
  form.name = r.name
  form.code = r.code
  form.remark = r.remark
  form.status = r.status
  formVisible.value = true
}

async function handleSave() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  saving.value = true
  try {
    if (isEdit.value) {
      await updateRole(form.id, {
        name: form.name,
        code: form.code,
        remark: form.remark,
        status: form.status,
      })
      ElMessage.success('更新成功')
    } else {
      await createRole({
        name: form.name,
        code: form.code,
        remark: form.remark,
      })
      ElMessage.success('创建成功')
    }
    formVisible.value = false
    await loadRoles()
  } finally {
    saving.value = false
  }
}

async function handleDelete(id: number) {
  await deleteRole(id)
  ElMessage.success('删除成功')
  await loadRoles()
}

// --- Permission dialog ---
const permVisible = ref(false)
const treeRef = ref<InstanceType<typeof ElTree>>()
const currentRoleId = ref(0)

async function openPermissions(row: Role) {
  currentRoleId.value = row.id
  const res = await getRole(row.id)
  const r = res.data as Role
  // Set checked keys from the role's current permissions
  const permIds = r.permissions?.map(p => p.id) || []
  await nextTick()
  treeRef.value?.setCheckedKeys(permIds)
  permVisible.value = true
}

async function handleSavePermissions() {
  const checkedKeys = (treeRef.value?.getCheckedKeys() || []) as number[]
  const halfCheckedKeys = (treeRef.value?.getHalfCheckedKeys() || []) as number[]
  await updateRolePermissions(currentRoleId.value, [...checkedKeys, ...halfCheckedKeys])
  ElMessage.success('权限分配成功')
  permVisible.value = false
}
</script>