import request from "../request"

/* ---- QCTask (质控任务) ---- */
export interface QCTask {
  id: number
  task_order_id: number
  qc_type: string
  qc_details: string
  created_at: string
  updated_at: string
}

export function getQCTaskList(params?: { task_order_id?: string }) {
  return request.get('/business/qc-tasks', { params })
}
export function getQCTask(id: number) {
  return request.get(`/business/qc-tasks/${id}`)
}
export function createQCTask(data: {
  task_order_id: number
  qc_type?: string
  qc_details?: string
}) {
  return request.post('/business/qc-tasks', data)
}
export function updateQCTask(id: number, data: Partial<QCTask>) {
  return request.put(`/business/qc-tasks/${id}`, data)
}
export function deleteQCTask(id: number) {
  return request.delete(`/business/qc-tasks/${id}`)
}
export function approveQCTask(id: number, data: {
  task_id: number
  qc_type?: string
  qc_details?: string
  comment?: string
}) {
  return request.post(`/business/qc-tasks/${id}/approve`, data)
}
export function rejectQCTask(id: number, data: {
  task_id: number
  comment: string
}) {
  return request.post(`/business/qc-tasks/${id}/reject`, data)
}

