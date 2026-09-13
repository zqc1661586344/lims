import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

interface JwtPayload {
  exp?: number
  iat?: number
  user_id?: number
  username?: string
  is_admin?: boolean
  permissions?: string[]
}

function decodeJwtPayload(token: string): JwtPayload | null {
  try {
    const parts = token.split('.')
    if (parts.length !== 3) return null
    const base64 = parts[1].replace(/-/g, '+').replace(/_/g, '/')
    const padded = base64 + '='.repeat((4 - (base64.length % 4)) % 4)
    const json = atob(padded)
    return JSON.parse(json)
  } catch {
    return null
  }
}

export const useUserStore = defineStore('user', () => {
  const token = ref(localStorage.getItem('token') || '')
  const username = ref(localStorage.getItem('username') || '')
  const isAdmin = ref(localStorage.getItem('is_admin') === 'true')
  const permissions = ref<string[]>(JSON.parse(localStorage.getItem('permissions') || '[]'))

  const isLoggedIn = computed(() => {
    if (!token.value) return false
    const payload = decodeJwtPayload(token.value)
    if (!payload?.exp) return true
    return payload.exp * 1000 > Date.now()
  })

  const isExpired = computed(() => {
    if (!token.value) return true
    const payload = decodeJwtPayload(token.value)
    if (!payload?.exp) return false
    return payload.exp * 1000 <= Date.now()
  })

  function hasPermission(code: string): boolean {
    if (isAdmin.value) return true
    const list = permissions.value
    return list.includes(code) || list.includes('*')
  }

  function hasAnyPermission(codes: string[]): boolean {
    if (isAdmin.value) return true
    if (codes.length === 0) return true
    const list = permissions.value
    if (list.includes('*')) return true
    return codes.some((c) => list.includes(c))
  }

  function hasAllPermissions(codes: string[]): boolean {
    if (isAdmin.value) return true
    if (codes.length === 0) return true
    const list = permissions.value
    if (list.includes('*')) return true
    return codes.every((c) => list.includes(c))
  }

  function setAuth(data: { token: string; username: string; is_admin: boolean; permissions: string[] }) {
    token.value = data.token
    username.value = data.username
    isAdmin.value = data.is_admin
    permissions.value = data.permissions
    localStorage.setItem('token', data.token)
    localStorage.setItem('username', data.username)
    localStorage.setItem('is_admin', String(data.is_admin))
    localStorage.setItem('permissions', JSON.stringify(data.permissions))
  }

  function clearAuth() {
    token.value = ''
    username.value = ''
    isAdmin.value = false
    permissions.value = []
    localStorage.clear()
    sessionStorage.clear()
  }

  return { token, username, isAdmin, permissions, isLoggedIn, isExpired, hasPermission, hasAnyPermission, hasAllPermissions, setAuth, clearAuth }
})