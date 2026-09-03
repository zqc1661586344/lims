import request from "../request"

/* ---- DataReview (数据复核, Node 9) ---- */
export interface DataReview {
  id: number
  task_order_id: number
  review_result: string
  review_comment: string
  issues_found: string
  created_at: string
  updated_at: string
}

export function getDataReviewList(params?: { task_order_id?: string }) {
  return request.get('/business/data-review', { params })
}
export function getDataReview(id: number) {
  return request.get(`/business/data-review/${id}`)
}
export function createDataReview(data: {
  task_order_id: number
  review_result?: string
  review_comment?: string
  issues_found?: string
}) {
  return request.post('/business/data-review', data)
}
export function updateDataReview(id: number, data: Partial<DataReview>) {
  return request.put(`/business/data-review/${id}`, data)
}
export function deleteDataReview(id: number) {
  return request.delete(`/business/data-review/${id}`)
}
export function approveDataReview(id: number, data: {
  task_id: number
  review_result?: string
  review_comment?: string
  issues_found?: string
}) {
  return request.post(`/business/data-review/${id}/approve`, data)
}
export function rejectDataReview(id: number, data: {
  task_id: number
  comment: string
}) {
  return request.post(`/business/data-review/${id}/reject`, data)
}

