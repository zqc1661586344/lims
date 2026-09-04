import request from '../request'

export interface LabSheetTemplate {
  id: number
  code: string
  name: string
  test_item_id: number
  node_code: string
  structure: any
  editable_ranges: string[]
  readonly_ranges: string[]
  version: number
  status: number
}

export interface LabSheet {
  id: number
  task_order_id: number
  test_item_id: number
  template_id: number | null
  node_code: string
  sheet_data: any
  formula_results: any
  status: number
  created_by: number | null
  updated_by: number | null
  template?: LabSheetTemplate
  created_at: string
  updated_at: string
}

export function listLabSheetTemplates(params?: { name?: string; test_item_id?: string; node_code?: string; status?: string }) {
  return request.get('/lab-sheets/templates', { params })
}

export function getLabSheetTemplate(id: number) {
  return request.get(`/lab-sheets/templates/${id}`)
}

export function createLabSheetTemplate(data: Partial<LabSheetTemplate>) {
  return request.post('/lab-sheets/templates', data)
}

export function updateLabSheetTemplate(id: number, data: Partial<LabSheetTemplate>) {
  return request.put(`/lab-sheets/templates/${id}`, data)
}

export function deleteLabSheetTemplate(id: number) {
  return request.delete(`/lab-sheets/templates/${id}`)
}

export function getLabSheetTemplateByItem(testItemID: number) {
  return request.get(`/lab-sheets/templates/by-item/${testItemID}`)
}

export function listLabSheets(params?: { task_order_id?: string; test_item_id?: string; node_code?: string; status?: string }) {
  return request.get('/lab-sheets', { params })
}

export function getLabSheet(id: number) {
  return request.get(`/lab-sheets/${id}`)
}

export function createLabSheet(data: {
  task_order_id: number
  test_item_id: number
  template_id?: number | null
  node_code?: string
  sheet_data?: any
}) {
  return request.post('/lab-sheets', data)
}

export function updateLabSheet(id: number, data: {
  sheet_data?: any
  formula_results?: any
  status?: number
  node_code?: string
}) {
  return request.put(`/lab-sheets/${id}`, data)
}

export function deleteLabSheet(id: number) {
  return request.delete(`/lab-sheets/${id}`)
}