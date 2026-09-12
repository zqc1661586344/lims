import request from "../request"

/* ---- FieldSamplingRecord (现场采样) ---- */
export interface FieldSamplingRecord {
  id: number
  task_order_id: number
  sample_photos: string
  equipment_cal_records: string
  sampling_record_file_path: string
  created_at: string
  updated_at: string
}

export function getFieldSamplingList(params?: { task_order_id?: string }) {
  return request.get('/business/field-sampling', { params })
}
export function getFieldSampling(id: number) {
  return request.get(`/business/field-sampling/${id}`)
}
export function createFieldSampling(data: {
  task_order_id: number
  sample_photos?: string
  equipment_cal_records?: string
  sampling_record_file_path?: string
}) {
  return request.post('/business/field-sampling', data)
}
export function updateFieldSampling(id: number, data: Partial<FieldSamplingRecord>) {
  return request.put(`/business/field-sampling/${id}`, data)
}
export function deleteFieldSampling(id: number) {
  return request.delete(`/business/field-sampling/${id}`)
}
export function approveFieldSampling(id: number, data: {
  task_id: number
  sample_photos?: string
  equipment_cal_records?: string
  sampling_record_file_path?: string
  comment?: string
}) {
  return request.post(`/business/field-sampling/${id}/approve`, data)
}
export function rejectFieldSampling(id: number, data: {
  task_id: number
  comment: string
  reject_target?: string
}) {
  return request.post(`/business/field-sampling/${id}/reject`, data)
}

