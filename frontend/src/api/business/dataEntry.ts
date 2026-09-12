import request from "../request"

/* ---- DataEntry (数据录入, Node 8) ---- */
export interface DataEntry {
  id: number
  task_order_id: number
  test_item_id: number
  original_data: string
  raw_record_id: number | null
  created_at: string
  updated_at: string
}

export function getDataEntryList(params?: { task_order_id?: string; test_item_id?: string }) {
  return request.get('/business/data-entry', { params })
}
export function getDataEntry(id: number) {
  return request.get(`/business/data-entry/${id}`)
}
export function createDataEntry(data: {
  task_order_id: number
  test_item_id?: number
  original_data?: string
  raw_record_id?: number | null
}) {
  return request.post('/business/data-entry', data)
}
export function updateDataEntry(id: number, data: Partial<DataEntry>) {
  return request.put(`/business/data-entry/${id}`, data)
}
export function deleteDataEntry(id: number) {
  return request.delete(`/business/data-entry/${id}`)
}
export function approveDataEntry(id: number, data: {
  task_id: number
  comment?: string
}) {
  return request.post(`/business/data-entry/${id}/approve`, data)
}
export function rejectDataEntry(id: number, data: {
  task_id: number
  comment: string
  reject_target?: string
}) {
  return request.post(`/business/data-entry/${id}/reject`, data)
}

