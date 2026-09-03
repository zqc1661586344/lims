import request from "../request"

/* ---- TaskOrder (任务委托) ---- */
export interface TaskOrder {
  id: number
  order_no: string
  customer_name: string
  project_name: string
  sample_type: string
  test_items: string
  status: number // 0=草稿, 1=已提交, 2=流程中, 3=已完成
  process_instance_id: number | null
  created_by: number | null
  created_at: string
  updated_at: string
}

export function getTaskOrderList(params?: { keyword?: string }) {
  return request.get('/business/task-orders', { params })
}
export function getTaskOrder(id: number) {
  return request.get(`/business/task-orders/${id}`)
}
export function createTaskOrder(data: {
  order_no: string
  customer_name: string
  project_name: string
  sample_type?: string
  test_items?: string
}) {
  return request.post('/business/task-orders', data)
}
export function updateTaskOrder(id: number, data: Partial<TaskOrder>) {
  return request.put(`/business/task-orders/${id}`, data)
}
export function deleteTaskOrder(id: number) {
  return request.delete(`/business/task-orders/${id}`)
}
export function submitTaskOrder(id: number) {
  return request.post(`/business/task-orders/${id}/submit`)
}

