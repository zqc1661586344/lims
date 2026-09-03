import request from "../request"

/* ---- DataAudit (数据审核, Node 10) ---- */
export interface DataAudit {
  id: number
  task_order_id: number
  audit_result: string
  audit_comment: string
  issue_list: string
  created_at: string
  updated_at: string
}

export function getDataAuditList(params?: { task_order_id?: string }) {
  return request.get('/business/data-audit', { params })
}
export function getDataAudit(id: number) {
  return request.get(`/business/data-audit/${id}`)
}
export function createDataAudit(data: {
  task_order_id: number
  audit_result?: string
  audit_comment?: string
  issue_list?: string
}) {
  return request.post('/business/data-audit', data)
}
export function updateDataAudit(id: number, data: Partial<DataAudit>) {
  return request.put(`/business/data-audit/${id}`, data)
}
export function deleteDataAudit(id: number) {
  return request.delete(`/business/data-audit/${id}`)
}
export function approveDataAudit(id: number, data: {
  task_id: number
  audit_result?: string
  audit_comment?: string
  issue_list?: string
}) {
  return request.post(`/business/data-audit/${id}/approve`, data)
}
export function rejectDataAudit(id: number, data: {
  task_id: number
  comment: string
}) {
  return request.post(`/business/data-audit/${id}/reject`, data)
}

