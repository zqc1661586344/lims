import request from './request'

export interface PendingTask {
  id: number
  node_code: string
  node_name: string
  title: string
  business_type: string
  business_id: number
  process_instance_id: number
  assignee_dept_id: number
  created_at: string
}

export function getPendingTasksByDept() {
  return request.get('/workflow/tasks/pending')
}
export function getPendingTasksByUser() {
  return request.get('/workflow/tasks/pending/user')
}
export function approveTask(id: number, data: { comment?: string }) {
  return request.post(`/workflow/tasks/${id}/approve`, data)
}
export function rejectTask(id: number, data: { comment: string }) {
  return request.post(`/workflow/tasks/${id}/reject`, data)
}
export function getProcessHistory(instanceId: number) {
  return request.get(`/workflow/instances/${instanceId}/history`)
}
export function getProcessInstance(instanceId: number) {
  return request.get(`/workflow/instances/${instanceId}`)
}
export function getNodeDefinitions() {
  return request.get('/workflow/nodes')
}