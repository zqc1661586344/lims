import { useUserStore } from '@/stores/user'

export function hasPermission(code: string): boolean {
  try {
    const store = useUserStore()
    return store.hasPermission(code)
  } catch {
    const perms = localStorage.getItem('permissions')
    if (!perms) return false
    const list: string[] = JSON.parse(perms)
    return list.includes(code) || list.includes('*')
  }
}

export function isAdmin(): boolean {
  try {
    return useUserStore().isAdmin
  } catch {
    return localStorage.getItem('is_admin') === 'true'
  }
}