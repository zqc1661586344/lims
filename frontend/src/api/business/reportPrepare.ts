import request from "../request"

/* ---- ReportPrepare (报告编制, Node 11) ---- */
export interface ReportPrepare {
  id: number
  task_order_id: number
  report_no: string
  prepare_opinion: string
  report_title: string
  report_content: string
  report_file: string
  attachments: string
  created_at: string
  updated_at: string
}

export function getReportPrepareList(params?: { task_order_id?: string }) {
  return request.get('/business/report-prepare', { params })
}
export function getReportPrepare(id: number) {
  return request.get(`/business/report-prepare/${id}`)
}
export function createReportPrepare(data: {
  task_order_id: number
  report_no?: string
  prepare_opinion?: string
  report_title?: string
  report_content?: string
  report_file?: string
  attachments?: string
}) {
  return request.post('/business/report-prepare', data)
}
export function updateReportPrepare(id: number, data: Partial<ReportPrepare>) {
  return request.put(`/business/report-prepare/${id}`, data)
}
export function deleteReportPrepare(id: number) {
  return request.delete(`/business/report-prepare/${id}`)
}
export function approveReportPrepare(id: number, data: {
  task_id: number
  report_no?: string
  prepare_opinion?: string
  report_title?: string
  report_content?: string
  report_file?: string
  attachments?: string
}) {
  return request.post(`/business/report-prepare/${id}/approve`, data)
}
export function rejectReportPrepare(id: number, data: {
  task_id: number
  comment: string
  reject_target?: string
}) {
  return request.post(`/business/report-prepare/${id}/reject`, data)
}

