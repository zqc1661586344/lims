import request from "../request"

/* ---- SamplingSchedule (采样调度) ---- */
export interface SamplingSchedule {
  id: number
  task_order_id: number
  sampling_team: string
  sampling_points: string
  equipment_list: string
  created_at: string
  updated_at: string
}

export function getSamplingScheduleList(params?: { task_order_id?: string }) {
  return request.get('/business/sampling-schedules', { params })
}
export function getSamplingSchedule(id: number) {
  return request.get(`/business/sampling-schedules/${id}`)
}
export function createSamplingSchedule(data: {
  task_order_id: number
  sampling_team?: string
  sampling_points?: string
  equipment_list?: string
}) {
  return request.post('/business/sampling-schedules', data)
}
export function updateSamplingSchedule(id: number, data: Partial<SamplingSchedule>) {
  return request.put(`/business/sampling-schedules/${id}`, data)
}
export function deleteSamplingSchedule(id: number) {
  return request.delete(`/business/sampling-schedules/${id}`)
}
export function approveSamplingSchedule(id: number, data: {
  task_id: number
  sampling_team?: string
  sampling_points?: string
  equipment_list?: string
  comment?: string
}) {
  return request.post(`/business/sampling-schedules/${id}/approve`, data)
}
export function rejectSamplingSchedule(id: number, data: {
  task_id: number
  comment: string
  reject_target?: string
}) {
  return request.post(`/business/sampling-schedules/${id}/reject`, data)
}

