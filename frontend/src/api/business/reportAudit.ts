import request from "../request"

/* ---- ReportAudit (报告审核, Node 13) ---- */
export interface ReportAudit {
  id: number
  task_order_id: number
  audit_result: string
  audit_comment: string
  audit_issues: string
  created_at: string
  updated_at: string
}

export function getReportAuditList(params?: { task_order_id?: string }) {
  return request.get('/business/report-audit', { params })
}
export function getReportAudit(id: number) {
  return request.get(`/business/report-audit/${id}`)
}
export function createReportAudit(data: {
  task_order_id: number
  audit_result?: string
  audit_comment?: string
  audit_issues?: string
}) {
  return request.post('/business/report-audit', data)
}
export function updateReportAudit(id: number, data: Partial<ReportAudit>) {
  return request.put(`/business/report-audit/${id}`, data)
}
export function deleteReportAudit(id: number) {
  return request.delete(`/business/report-audit/${id}`)
}
export function approveReportAudit(id: number, data: {
  task_id: number
  audit_result?: string
  audit_comment?: string
  audit_issues?: string
}) {
  return request.post(`/business/report-audit/${id}/approve`, data)
}
export function rejectReportAudit(id: number, data: {
  task_id: number
  comment: string
}) {
  return request.post(`/business/report-audit/${id}/reject`, data)
}

