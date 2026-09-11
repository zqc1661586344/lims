import request from './request'

/* ---- Test Items (检测项目) ---- */
export interface TestItem {
  id: number
  name: string
  code: string
  category: string
  unit: string
  method: string
  standard_id: number | null
  price: number
  status: number
  created_at: string
  updated_at: string
  standard?: TestStandard
}

export function getTestItems(params?: {
  keyword?: string
  category?: string
  page?: number
  page_size?: number
}) {
  return request.get('/base-data/items', { params })
}

export function getTestItem(id: number) {
  return request.get(`/base-data/items/${id}`)
}

export function createTestItem(data: {
  name: string
  code: string
  category?: string
  unit?: string
  method?: string
  standard_id?: number | null
  price?: number
}) {
  return request.post('/base-data/items', data)
}

export function updateTestItem(id: number, data: Partial<TestItem>) {
  return request.put(`/base-data/items/${id}`, data)
}

export function deleteTestItem(id: number) {
  return request.delete(`/base-data/items/${id}`)
}

/* ---- Test Standards (检测标准) ---- */
export interface TestStandard {
  id: number
  name: string
  code: string
  issuer: string
  version: string
  publish_date: string
  file_path: string
  status: number
  created_at: string
  updated_at: string
}

export function getTestStandards(params?: {
  keyword?: string
  page?: number
  page_size?: number
}) {
  return request.get('/base-data/standards', { params })
}

export function getTestStandard(id: number) {
  return request.get(`/base-data/standards/${id}`)
}

export function createTestStandard(data: {
  name: string
  code: string
  issuer?: string
  version?: string
  publish_date?: string
  file_path?: string
}) {
  return request.post('/base-data/standards', data)
}

export function updateTestStandard(id: number, data: Partial<TestStandard>) {
  return request.put(`/base-data/standards/${id}`, data)
}

export function deleteTestStandard(id: number) {
  return request.delete(`/base-data/standards/${id}`)
}

/* ---- Equipment (仪器设备) ---- */
export interface Equipment {
  id: number
  name: string
  code: string
  model: string
  factory: string
  calibration_date: string
  next_cal_date: string
  status: number
  created_at: string
  updated_at: string
}

export function getEquipmentList(params?: {
  keyword?: string
  page?: number
  page_size?: number
}) {
  return request.get('/base-data/equipment', { params })
}

export function getEquipment(id: number) {
  return request.get(`/base-data/equipment/${id}`)
}

export function createEquipment(data: {
  name: string
  code: string
  model?: string
  factory?: string
  calibration_date?: string
  next_cal_date?: string
}) {
  return request.post('/base-data/equipment', data)
}

export function updateEquipment(id: number, data: Partial<Equipment>) {
  return request.put(`/base-data/equipment/${id}`, data)
}

export function deleteEquipment(id: number) {
  return request.delete(`/base-data/equipment/${id}`)
}

/* ---- Reagents (物资管理) ---- */
export interface Reagent {
  id: number
  name: string
  code: string
  spec: string
  manufacturer: string
  batch_no: string
  stock_qty: number
  unit: string
  expire_date: string
  status: number
  created_at: string
  updated_at: string
}

export function getReagents(params?: {
  keyword?: string
  page?: number
  page_size?: number
}) {
  return request.get('/base-data/reagents', { params })
}

export function getReagent(id: number) {
  return request.get(`/base-data/reagents/${id}`)
}

export function createReagent(data: {
  name: string
  code: string
  spec?: string
  manufacturer?: string
  batch_no?: string
  stock_qty?: number
  unit?: string
  expire_date?: string
}) {
  return request.post('/base-data/reagents', data)
}

export function updateReagent(id: number, data: Partial<Reagent>) {
  return request.put(`/base-data/reagents/${id}`, data)
}

export function deleteReagent(id: number) {
  return request.delete(`/base-data/reagents/${id}`)
}