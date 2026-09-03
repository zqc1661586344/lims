import request from "../request"

/* ---- ReportPrint (报告打印发放, Node 15) ---- */
export interface ReportPrint {
  id: number
  task_order_id: number
  print_count: number
  print_result: string
  print_comment: string
  recipient_name: string
  recipient_date: string
  delivery_method: string
  tracking_no: string
  created_at: string
  updated_at: string
}

export function getReportPrintList(params?: { task_order_id?: string }) {
  return request.get('/business/report-print', { params })
}
export function getReportPrint(id: number) {
  return request.get(`/business/report-print/${id}`)
}
export function createReportPrint(data: {
  task_order_id: number
  print_count?: number
  print_result?: string
  print_comment?: string
  recipient_name?: string
  recipient_date?: string
  delivery_method?: string
  tracking_no?: string
}) {
  return request.post('/business/report-print', data)
}
export function updateReportPrint(id: number, data: Partial<ReportPrint>) {
  return request.put(`/business/report-print/${id}`, data)
}
export function deleteReportPrint(id: number) {
  return request.delete(`/business/report-print/${id}`)
}
export function approveReportPrint(id: number, data: {
  task_id: number
  print_count?: number
  print_result?: string
  print_comment?: string
  recipient_name?: string
  recipient_date?: string
  delivery_method?: string
  tracking_no?: string
}) {
  return request.post(`/business/report-print/${id}/approve`, data)
}
export function rejectReportPrint(id: number, data: {
  task_id: number
  comment: string
}) {
  return request.post(`/business/report-print/${id}/reject`, data)
}

