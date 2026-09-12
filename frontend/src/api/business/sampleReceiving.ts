import request from "../request"

/* ---- SampleReceiving (样品接收) ---- */
export interface SampleReceiving {
  id: number
  task_order_id: number
  sample_condition: string
  sample_codes: string
  receiving_record_path: string
  created_at: string
  updated_at: string
}

export function getSampleReceivingList(params?: { task_order_id?: string }) {
  return request.get('/business/sample-receiving', { params })
}
export function getSampleReceiving(id: number) {
  return request.get(`/business/sample-receiving/${id}`)
}
export function createSampleReceiving(data: {
  task_order_id: number
  sample_condition?: string
  sample_codes?: string
  receiving_record_path?: string
}) {
  return request.post('/business/sample-receiving', data)
}
export function updateSampleReceiving(id: number, data: Partial<SampleReceiving>) {
  return request.put(`/business/sample-receiving/${id}`, data)
}
export function deleteSampleReceiving(id: number) {
  return request.delete(`/business/sample-receiving/${id}`)
}
export function approveSampleReceiving(id: number, data: {
  task_id: number
  sample_condition?: string
  sample_codes?: string
  receiving_record_path?: string
  comment?: string
}) {
  return request.post(`/business/sample-receiving/${id}/approve`, data)
}
export function rejectSampleReceiving(id: number, data: {
  task_id: number
  comment: string
  reject_target?: string
}) {
  return request.post(`/business/sample-receiving/${id}/reject`, data)
}

