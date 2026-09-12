import request from "../request"

/* ---- ReportSign (报告签发, Node 14) ---- */
export interface ReportSign {
  id: number
  task_order_id: number
  report_no: string
  report_title: string
  prepare_opinion: string
  review_opinion: string
  audit_opinion: string
  raw_records: string
  sign_result: string
  sign_comment: string
  signer_name: string
  sign_date: string
  sign_stamp: string
  created_at: string
  updated_at: string
}

export function getReportSignList(params?: { task_order_id?: string }) {
  return request.get('/business/report-sign', { params })
}
export function getReportSign(id: number) {
  return request.get(`/business/report-sign/${id}`)
}
export function createReportSign(data: {
  task_order_id: number
  sign_result?: string
  sign_comment?: string
  signer_name?: string
  sign_date?: string
  sign_stamp?: string
}) {
  return request.post('/business/report-sign', data)
}
export function updateReportSign(id: number, data: Partial<ReportSign>) {
  return request.put(`/business/report-sign/${id}`, data)
}
export function deleteReportSign(id: number) {
  return request.delete(`/business/report-sign/${id}`)
}
export function approveReportSign(id: number, data: {
  task_id: number
  report_no?: string
  sign_result?: string
  sign_comment?: string
  signer_name?: string
  sign_date?: string
  sign_stamp?: string
}) {
  return request.post(`/business/report-sign/${id}/approve`, data)
}
export function rejectReportSign(id: number, data: {
  task_id: number
  comment: string
  reject_target?: string
}) {
  return request.post(`/business/report-sign/${id}/reject`, data)
}

