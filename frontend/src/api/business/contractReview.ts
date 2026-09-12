import request from "../request"

/* ---- ContractReview (合同评审) ---- */
export interface ContractReview {
  id: number
  task_order_id: number
  review_result: string
  review_comment: string
  contract_file_path: string
  created_at: string
  updated_at: string
}

export function getContractReviewList(params?: { task_order_id?: string }) {
  return request.get('/business/contract-reviews', { params })
}
export function getContractReview(id: number) {
  return request.get(`/business/contract-reviews/${id}`)
}
export function createContractReview(data: {
  task_order_id: number
  review_result?: string
  review_comment?: string
  contract_file_path?: string
}) {
  return request.post('/business/contract-reviews', data)
}
export function updateContractReview(id: number, data: Partial<ContractReview>) {
  return request.put(`/business/contract-reviews/${id}`, data)
}
export function deleteContractReview(id: number) {
  return request.delete(`/business/contract-reviews/${id}`)
}
export function approveContractReview(id: number, data: {
  task_id: number
  review_result?: string
  review_comment?: string
  contract_file_path?: string
}) {
  return request.post(`/business/contract-reviews/${id}/approve`, data)
}
export function rejectContractReview(id: number, data: {
  task_id: number
  review_comment: string
  reject_target?: string
}) {
  return request.post(`/business/contract-reviews/${id}/reject`, data)
}

