import request from '../request'

export interface SamplingSheetTemplate {
  id: number
  code: string
  name: string
  sample_type: string
  node_code: string
  structure: any
  editable_ranges: string[]
  readonly_ranges: string[]
  version: number
  status: number
  created_at?: string
  updated_at?: string
}

export interface SamplingSheet {
  id: number
  task_order_id: number
  sampling_point: string
  template_id: number | null
  node_code: string
  sheet_data: any
  formula_results: any
  status: number
  created_by: number | null
  updated_by: number | null
  template?: SamplingSheetTemplate
  created_at?: string
  updated_at?: string
}

export function listSamplingSheetTemplates(params?: { name?: string; sample_type?: string; node_code?: string; status?: string }) {
  return request.get('/sampling-sheets/templates', { params })
}

export function getSamplingSheetTemplate(id: number) {
  return request.get(`/sampling-sheets/templates/${id}`)
}

export function createSamplingSheetTemplate(data: Partial<SamplingSheetTemplate>) {
  return request.post('/sampling-sheets/templates', data)
}

export function updateSamplingSheetTemplate(id: number, data: Partial<SamplingSheetTemplate>) {
  return request.put(`/sampling-sheets/templates/${id}`, data)
}

export function deleteSamplingSheetTemplate(id: number) {
  return request.delete(`/sampling-sheets/templates/${id}`)
}

export function getSamplingSheetTemplateBySampleType(sampleType: string) {
  return request.get(`/sampling-sheets/templates/by-sample-type/${sampleType}`)
}

export function listSamplingSheets(params?: { task_order_id?: string; node_code?: string; status?: string }) {
  return request.get('/sampling-sheets', { params })
}

export function getSamplingSheet(id: number) {
  return request.get(`/sampling-sheets/${id}`)
}

export function createSamplingSheet(data: {
  task_order_id: number
  sampling_point?: string
  sample_type?: string
  template_id?: number | null
  node_code?: string
  sheet_data?: any
}) {
  return request.post('/sampling-sheets', data)
}

export function updateSamplingSheet(id: number, data: {
  sampling_point?: string
  sheet_data?: any
  formula_results?: any
  status?: number
  node_code?: string
}) {
  return request.put(`/sampling-sheets/${id}`, data)
}

export function deleteSamplingSheet(id: number) {
  return request.delete(`/sampling-sheets/${id}`)
}