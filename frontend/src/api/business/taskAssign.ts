import request from "../request"

/* ---- TaskAssign (任务分配, Node 7) ---- */
export interface TaskAssign {
  id: number
  task_order_id: number
  assigned_to: string
  test_item_list: string
  created_at: string
  updated_at: string
}

export function getTaskAssignList(params?: { task_order_id?: string }) {
  return request.get('/business/task-assign', { params })
}
export function getTaskAssign(id: number) {
  return request.get(`/business/task-assign/${id}`)
}
export function createTaskAssign(data: {
  task_order_id: number
  assigned_to?: string
  test_item_list?: string
}) {
  return request.post('/business/task-assign', data)
}
export function updateTaskAssign(id: number, data: Partial<TaskAssign>) {
  return request.put(`/business/task-assign/${id}`, data)
}
export function deleteTaskAssign(id: number) {
  return request.delete(`/business/task-assign/${id}`)
}
export function approveTaskAssign(id: number, data: {
  task_id: number
  assigned_to?: string
  test_item_list?: string
  comment?: string
}) {
  return request.post(`/business/task-assign/${id}/approve`, data)
}
export function rejectTaskAssign(id: number, data: {
  task_id: number
  comment: string
}) {
  return request.post(`/business/task-assign/${id}/reject`, data)
}

