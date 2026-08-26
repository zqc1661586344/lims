import request from './request'

/* ---- Users ---- */
export interface User {
  id: number
  username: string
  real_name: string
  email: string
  phone: string
  dept_id: number | null
  status: number
  is_admin: boolean
  created_at: string
  dept?: { id: number; name: string }
  roles?: { id: number; name: string }[]
}

export function getUsers(params?: { dept_id?: string; keyword?: string }) {
  return request.get('/system/users', { params })
}

export function getUser(id: number) {
  return request.get(`/system/users/${id}`)
}

export function createUser(data: {
  username: string
  password: string
  real_name?: string
  email?: string
  phone?: string
  dept_id?: number | null
}) {
  return request.post('/system/users', data)
}

export function updateUser(id: number, data: Partial<User>) {
  return request.put(`/system/users/${id}`, data)
}

export function deleteUser(id: number) {
  return request.delete(`/system/users/${id}`)
}

export function updateUserRoles(id: number, role_ids: number[]) {
  return request.put(`/system/users/${id}/roles`, { role_ids })
}

/* ---- Departments ---- */
export interface Dept {
  id: number
  name: string
  code: string
  sort: number
  status: number
  parent_id: number | null
  children?: Dept[]
}

export function getDepts() {
  return request.get('/system/depts')
}

export function createDept(data: { name: string; code: string; sort?: number; parent_id?: number | null }) {
  return request.post('/system/depts', data)
}

export function updateDept(id: number, data: Partial<Dept>) {
  return request.put(`/system/depts/${id}`, data)
}

export function deleteDept(id: number) {
  return request.delete(`/system/depts/${id}`)
}

/* ---- Roles ---- */
export interface Role {
  id: number
  name: string
  code: string
  status: number
  remark: string
  permissions?: PermissionNode[]
}

export function getRoles() {
  return request.get('/system/roles')
}

export function getRole(id: number) {
  return request.get(`/system/roles/${id}`)
}

export function createRole(data: { name: string; code: string; remark?: string }) {
  return request.post('/system/roles', data)
}

export function updateRole(id: number, data: Partial<Role>) {
  return request.put(`/system/roles/${id}`, data)
}

export function deleteRole(id: number) {
  return request.delete(`/system/roles/${id}`)
}

export function updateRolePermissions(id: number, permission_ids: number[]) {
  return request.put(`/system/roles/${id}/permissions`, { permission_ids })
}

/* ---- Permissions ---- */
export interface PermissionNode {
  id: number
  name: string
  code: string
  type: string
  parent_id: number | null
  path: string
  icon: string
  sort: number
  children?: PermissionNode[]
}

export function getPermissions() {
  return request.get('/system/permissions')
}