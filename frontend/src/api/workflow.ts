import request from './request'

export interface PendingTask {
  id: number
  node_code: string
  node_name: string
  next_node?: string
  next_node_name?: string
  next_dept_name?: string
  title: string
  business_type: string
  business_id: number
  process_instance_id: number
  assignee_dept_id: number
  dept_name?: string
  created_at: string
}

export interface TimelineNode {
  code: string
  name: string
  time: string
  status: string
  active: boolean
  operator: string
  dept: string
  comment: string
}

export interface NodeDefinition {
  code: string
  name: string
  dept_code: string
  can_reject: boolean
  reject_target: string
  next_node: string
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
export function rejectTask(id: number, data: { comment: string; reject_target?: string }) {
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

export interface ProgressNode {
  code: string
  name: string
  dept_code: string
  index: number
  status: 'completed' | 'current' | 'pending' | 'rejected' | 'skipped'
  can_reject: boolean
  has_task: boolean
  comment?: string
}

export interface WorkflowProgress {
  instance_id: number
  current_node: string
  current_node_index: number
  total_nodes: number
  status: string
  title: string
  created_at: string
  nodes: ProgressNode[]
}

export function getWorkflowProgress(businessType: string, businessId: number) {
  return request.get(`/workflow/progress/${businessType}/${businessId}`)
}