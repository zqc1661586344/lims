import request from "../request"

/* ---- ReportReview (报告复核, Node 12) ---- */
export interface ReportReview {
  id: number
  task_order_id: number
  review_result: string
  review_comment: string
  reviewed_items: string
  created_at: string
  updated_at: string
}

export function getReportReviewList(params?: { task_order_id?: string }) {
  return request.get('/business/report-review', { params })
}
export function getReportReview(id: number) {
  return request.get(`/business/report-review/${id}`)
}
export function createReportReview(data: {
  task_order_id: number
  review_result?: string
  review_comment?: string
  reviewed_items?: string
}) {
  return request.post('/business/report-review', data)
}
export function updateReportReview(id: number, data: Partial<ReportReview>) {
  return request.put(`/business/report-review/${id}`, data)
}
export function deleteReportReview(id: number) {
  return request.delete(`/business/report-review/${id}`)
}
export function approveReportReview(id: number, data: {
  task_id: number
  review_result?: string
  review_comment?: string
  reviewed_items?: string
}) {
  return request.post(`/business/report-review/${id}/approve`, data)
}
export function rejectReportReview(id: number, data: {
  task_id: number
  comment: string
  reject_target?: string
}) {
  return request.post(`/business/report-review/${id}/reject`, data)
}

