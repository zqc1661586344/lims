/**
 * Check if the user has a specific permission code.
 * In Phase 2 this is a simple localStorage check.
 * Enhanced in later phases with actual permission tree.
 */
export function hasPermission(code: string): boolean {
  const perms = localStorage.getItem('permissions')
  if (!perms) return false
  const list: string[] = JSON.parse(perms)
  return list.includes(code) || list.includes('*')
}

/**
 * Check if the current user is admin.
 */
export function isAdmin(): boolean {
  return localStorage.getItem('is_admin') === 'true'
}