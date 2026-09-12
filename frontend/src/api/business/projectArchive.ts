import request from "../request"

/* ---- ProjectArchive (项目归档, Node 16) ---- */
export interface ProjectArchive {
  id: number
  task_order_id: number
  archive_no: string
  archive_location: string
  archive_date: string
  archive_files: string
  archive_comment: string
  retention_period: number
  created_at: string
  updated_at: string
}

export function getProjectArchiveList(params?: { task_order_id?: string }) {
  return request.get('/business/project-archive', { params })
}
export function getProjectArchive(id: number) {
  return request.get(`/business/project-archive/${id}`)
}
export function createProjectArchive(data: {
  task_order_id: number
  archive_no?: string
  archive_location?: string
  archive_date?: string
  archive_files?: string
  archive_comment?: string
  retention_period?: number
}) {
  return request.post('/business/project-archive', data)
}
export function updateProjectArchive(id: number, data: Partial<ProjectArchive>) {
  return request.put(`/business/project-archive/${id}`, data)
}
export function deleteProjectArchive(id: number) {
  return request.delete(`/business/project-archive/${id}`)
}
export function approveProjectArchive(id: number, data: {
  task_id: number
  archive_no?: string
  archive_location?: string
  archive_date?: string
  archive_files?: string
  archive_comment?: string
  retention_period?: number
}) {
  return request.post(`/business/project-archive/${id}/approve`, data)
}
export function rejectProjectArchive(id: number, data: {
  task_id: number
  comment: string
  reject_target?: string
}) {
  return request.post(`/business/project-archive/${id}/reject`, data)
}
