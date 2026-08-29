import request from './request'

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

/* ---- ContractReview (合同评审) ---- */
export interface ContractReview {
  id: number
  task_order_id: number
  review_result: string
  review_comment: string
  contract_file_path: string
  created_at: string
  updated_at: string
}

export function getContractReviewList(params?: { task_order_id?: string }) {
  return request.get('/business/contract-reviews', { params })
}
export function getContractReview(id: number) {
  return request.get(`/business/contract-reviews/${id}`)
}
export function createContractReview(data: {
  task_order_id: number
  review_result?: string
  review_comment?: string
  contract_file_path?: string
}) {
  return request.post('/business/contract-reviews', data)
}
export function updateContractReview(id: number, data: Partial<ContractReview>) {
  return request.put(`/business/contract-reviews/${id}`, data)
}
export function deleteContractReview(id: number) {
  return request.delete(`/business/contract-reviews/${id}`)
}
export function approveContractReview(id: number, data: {
  task_id: number
  review_result?: string
  review_comment?: string
  contract_file_path?: string
}) {
  return request.post(`/business/contract-reviews/${id}/approve`, data)
}
export function rejectContractReview(id: number, data: {
  task_id: number
  review_comment: string
}) {
  return request.post(`/business/contract-reviews/${id}/reject`, data)
}

/* ---- QCTask (质控任务) ---- */
export interface QCTask {
  id: number
  task_order_id: number
  qc_type: string
  qc_details: string
  created_at: string
  updated_at: string
}

export function getQCTaskList(params?: { task_order_id?: string }) {
  return request.get('/business/qc-tasks', { params })
}
export function getQCTask(id: number) {
  return request.get(`/business/qc-tasks/${id}`)
}
export function createQCTask(data: {
  task_order_id: number
  qc_type?: string
  qc_details?: string
}) {
  return request.post('/business/qc-tasks', data)
}
export function updateQCTask(id: number, data: Partial<QCTask>) {
  return request.put(`/business/qc-tasks/${id}`, data)
}
export function deleteQCTask(id: number) {
  return request.delete(`/business/qc-tasks/${id}`)
}
export function approveQCTask(id: number, data: {
  task_id: number
  qc_type?: string
  qc_details?: string
  comment?: string
}) {
  return request.post(`/business/qc-tasks/${id}/approve`, data)
}
export function rejectQCTask(id: number, data: {
  task_id: number
  comment: string
}) {
  return request.post(`/business/qc-tasks/${id}/reject`, data)
}

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
}) {
  return request.post(`/business/sampling-schedules/${id}/reject`, data)
}

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
}) {
  return request.post(`/business/field-sampling/${id}/reject`, data)
}

/* ---- SampleReceiving (样品接收) ---- */
export interface SampleReceiving {
  id: number
  task_order_id: number
  sample_condition: string
  sample_codes: string
  receiving_record_path: string
  created_at: string
  updated_at: string
}

export function getSampleReceivingList(params?: { task_order_id?: string }) {
  return request.get('/business/sample-receiving', { params })
}
export function getSampleReceiving(id: number) {
  return request.get(`/business/sample-receiving/${id}`)
}
export function createSampleReceiving(data: {
  task_order_id: number
  sample_condition?: string
  sample_codes?: string
  receiving_record_path?: string
}) {
  return request.post('/business/sample-receiving', data)
}
export function updateSampleReceiving(id: number, data: Partial<SampleReceiving>) {
  return request.put(`/business/sample-receiving/${id}`, data)
}
export function deleteSampleReceiving(id: number) {
  return request.delete(`/business/sample-receiving/${id}`)
}
export function approveSampleReceiving(id: number, data: {
  task_id: number
  sample_condition?: string
  sample_codes?: string
  receiving_record_path?: string
  comment?: string
}) {
  return request.post(`/business/sample-receiving/${id}/approve`, data)
}
export function rejectSampleReceiving(id: number, data: {
  task_id: number
  comment: string
}) {
  return request.post(`/business/sample-receiving/${id}/reject`, data)
}

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
}) {
  return request.post(`/business/data-entry/${id}/reject`, data)
}

/* ---- DataReview (数据复核, Node 9) ---- */
export interface DataReview {
  id: number
  task_order_id: number
  review_result: string
  review_comment: string
  issues_found: string
  created_at: string
  updated_at: string
}

export function getDataReviewList(params?: { task_order_id?: string }) {
  return request.get('/business/data-review', { params })
}
export function getDataReview(id: number) {
  return request.get(`/business/data-review/${id}`)
}
export function createDataReview(data: {
  task_order_id: number
  review_result?: string
  review_comment?: string
  issues_found?: string
}) {
  return request.post('/business/data-review', data)
}
export function updateDataReview(id: number, data: Partial<DataReview>) {
  return request.put(`/business/data-review/${id}`, data)
}
export function deleteDataReview(id: number) {
  return request.delete(`/business/data-review/${id}`)
}
export function approveDataReview(id: number, data: {
  task_id: number
  review_result?: string
  review_comment?: string
  issues_found?: string
}) {
  return request.post(`/business/data-review/${id}/approve`, data)
}
export function rejectDataReview(id: number, data: {
  task_id: number
  comment: string
}) {
  return request.post(`/business/data-review/${id}/reject`, data)
}

/* ---- DataAudit (数据审核, Node 10) ---- */
export interface DataAudit {
  id: number
  task_order_id: number
  audit_result: string
  audit_comment: string
  issue_list: string
  created_at: string
  updated_at: string
}

export function getDataAuditList(params?: { task_order_id?: string }) {
  return request.get('/business/data-audit', { params })
}
export function getDataAudit(id: number) {
  return request.get(`/business/data-audit/${id}`)
}
export function createDataAudit(data: {
  task_order_id: number
  audit_result?: string
  audit_comment?: string
  issue_list?: string
}) {
  return request.post('/business/data-audit', data)
}
export function updateDataAudit(id: number, data: Partial<DataAudit>) {
  return request.put(`/business/data-audit/${id}`, data)
}
export function deleteDataAudit(id: number) {
  return request.delete(`/business/data-audit/${id}`)
}
export function approveDataAudit(id: number, data: {
  task_id: number
  audit_result?: string
  audit_comment?: string
  issue_list?: string
}) {
  return request.post(`/business/data-audit/${id}/approve`, data)
}
export function rejectDataAudit(id: number, data: {
  task_id: number
  comment: string
}) {
  return request.post(`/business/data-audit/${id}/reject`, data)
}

/* ---- ReportPrepare (报告编制, Node 11) ---- */
export interface ReportPrepare {
  id: number
  task_order_id: number
  report_no: string
  prepare_opinion: string
  report_title: string
  report_content: string
  report_file: string
  attachments: string
  created_at: string
  updated_at: string
}

export function getReportPrepareList(params?: { task_order_id?: string }) {
  return request.get('/business/report-prepare', { params })
}
export function getReportPrepare(id: number) {
  return request.get(`/business/report-prepare/${id}`)
}
export function createReportPrepare(data: {
  task_order_id: number
  report_no?: string
  prepare_opinion?: string
  report_title?: string
  report_content?: string
  report_file?: string
  attachments?: string
}) {
  return request.post('/business/report-prepare', data)
}
export function updateReportPrepare(id: number, data: Partial<ReportPrepare>) {
  return request.put(`/business/report-prepare/${id}`, data)
}
export function deleteReportPrepare(id: number) {
  return request.delete(`/business/report-prepare/${id}`)
}
export function approveReportPrepare(id: number, data: {
  task_id: number
  report_no?: string
  prepare_opinion?: string
  report_title?: string
  report_content?: string
  report_file?: string
  attachments?: string
}) {
  return request.post(`/business/report-prepare/${id}/approve`, data)
}
export function rejectReportPrepare(id: number, data: {
  task_id: number
  comment: string
}) {
  return request.post(`/business/report-prepare/${id}/reject`, data)
}

/* ---- ReportReview (报告复核, Node 12) ---- */
export interface ReportReview {
  id: number
  task_order_id: number
  review_result: string
  review_comment: string
  reviewed_items: string
  created_at: string
  updated_at: string
}

export function getReportReviewList(params?: { task_order_id?: string }) {
  return request.get('/business/report-review', { params })
}
export function getReportReview(id: number) {
  return request.get(`/business/report-review/${id}`)
}
export function createReportReview(data: {
  task_order_id: number
  review_result?: string
  review_comment?: string
  reviewed_items?: string
}) {
  return request.post('/business/report-review', data)
}
export function updateReportReview(id: number, data: Partial<ReportReview>) {
  return request.put(`/business/report-review/${id}`, data)
}
export function deleteReportReview(id: number) {
  return request.delete(`/business/report-review/${id}`)
}
export function approveReportReview(id: number, data: {
  task_id: number
  review_result?: string
  review_comment?: string
  reviewed_items?: string
}) {
  return request.post(`/business/report-review/${id}/approve`, data)
}
export function rejectReportReview(id: number, data: {
  task_id: number
  comment: string
}) {
  return request.post(`/business/report-review/${id}/reject`, data)
}

/* ---- ReportAudit (报告审核, Node 13) ---- */
export interface ReportAudit {
  id: number
  task_order_id: number
  audit_result: string
  audit_comment: string
  audit_issues: string
  created_at: string
  updated_at: string
}

export function getReportAuditList(params?: { task_order_id?: string }) {
  return request.get('/business/report-audit', { params })
}
export function getReportAudit(id: number) {
  return request.get(`/business/report-audit/${id}`)
}
export function createReportAudit(data: {
  task_order_id: number
  audit_result?: string
  audit_comment?: string
  audit_issues?: string
}) {
  return request.post('/business/report-audit', data)
}
export function updateReportAudit(id: number, data: Partial<ReportAudit>) {
  return request.put(`/business/report-audit/${id}`, data)
}
export function deleteReportAudit(id: number) {
  return request.delete(`/business/report-audit/${id}`)
}
export function approveReportAudit(id: number, data: {
  task_id: number
  audit_result?: string
  audit_comment?: string
  audit_issues?: string
}) {
  return request.post(`/business/report-audit/${id}/approve`, data)
}
export function rejectReportAudit(id: number, data: {
  task_id: number
  comment: string
}) {
  return request.post(`/business/report-audit/${id}/reject`, data)
}

/* ---- ReportSign (报告签发, Node 14) ---- */
export interface ReportSign {
  id: number
  task_order_id: number
  report_no: string
  report_title: string
  prepare_opinion: string
  review_opinion: string
  audit_opinion: string
  raw_records: string
  sign_result: string
  sign_comment: string
  signer_name: string
  sign_date: string
  sign_stamp: string
  created_at: string
  updated_at: string
}

export function getReportSignList(params?: { task_order_id?: string }) {
  return request.get('/business/report-sign', { params })
}
export function getReportSign(id: number) {
  return request.get(`/business/report-sign/${id}`)
}
export function createReportSign(data: {
  task_order_id: number
  sign_result?: string
  sign_comment?: string
  signer_name?: string
  sign_date?: string
  sign_stamp?: string
}) {
  return request.post('/business/report-sign', data)
}
export function updateReportSign(id: number, data: Partial<ReportSign>) {
  return request.put(`/business/report-sign/${id}`, data)
}
export function deleteReportSign(id: number) {
  return request.delete(`/business/report-sign/${id}`)
}
export function approveReportSign(id: number, data: {
  task_id: number
  sign_result?: string
  sign_comment?: string
  signer_name?: string
  sign_date?: string
  sign_stamp?: string
}) {
  return request.post(`/business/report-sign/${id}/approve`, data)
}
export function rejectReportSign(id: number, data: {
  task_id: number
  comment: string
}) {
  return request.post(`/business/report-sign/${id}/reject`, data)
}

/* ---- ReportPrint (报告打印发放, Node 15) ---- */
export interface ReportPrint {
  id: number
  task_order_id: number
  print_count: number
  print_result: string
  print_comment: string
  recipient_name: string
  recipient_date: string
  delivery_method: string
  tracking_no: string
  created_at: string
  updated_at: string
}

export function getReportPrintList(params?: { task_order_id?: string }) {
  return request.get('/business/report-print', { params })
}
export function getReportPrint(id: number) {
  return request.get(`/business/report-print/${id}`)
}
export function createReportPrint(data: {
  task_order_id: number
  print_count?: number
  print_result?: string
  print_comment?: string
  recipient_name?: string
  recipient_date?: string
  delivery_method?: string
  tracking_no?: string
}) {
  return request.post('/business/report-print', data)
}
export function updateReportPrint(id: number, data: Partial<ReportPrint>) {
  return request.put(`/business/report-print/${id}`, data)
}
export function deleteReportPrint(id: number) {
  return request.delete(`/business/report-print/${id}`)
}
export function approveReportPrint(id: number, data: {
  task_id: number
  print_count?: number
  print_result?: string
  print_comment?: string
  recipient_name?: string
  recipient_date?: string
  delivery_method?: string
  tracking_no?: string
}) {
  return request.post(`/business/report-print/${id}/approve`, data)
}
export function rejectReportPrint(id: number, data: {
  task_id: number
  comment: string
}) {
  return request.post(`/business/report-print/${id}/reject`, data)
}

/* ---- ProjectArchive (项目归档, Node 16) ---- */
export interface ProjectArchive {
  id: number
  task_order_id: number
  archive_no: string
  archive_location: string
  archive_date: string
  archive_files: string
  archive_comment: string
  retention_period: number
  created_at: string
  updated_at: string
}

export function getProjectArchiveList(params?: { task_order_id?: string }) {
  return request.get('/business/project-archive', { params })
}
export function getProjectArchive(id: number) {
  return request.get(`/business/project-archive/${id}`)
}
export function createProjectArchive(data: {
  task_order_id: number
  archive_no?: string
  archive_location?: string
  archive_date?: string
  archive_files?: string
  archive_comment?: string
  retention_period?: number
}) {
  return request.post('/business/project-archive', data)
}
export function updateProjectArchive(id: number, data: Partial<ProjectArchive>) {
  return request.put(`/business/project-archive/${id}`, data)
}
export function deleteProjectArchive(id: number) {
  return request.delete(`/business/project-archive/${id}`)
}
export function approveProjectArchive(id: number, data: {
  task_id: number
  archive_no?: string
  archive_location?: string
  archive_date?: string
  archive_files?: string
  archive_comment?: string
  retention_period?: number
}) {
  return request.post(`/business/project-archive/${id}/approve`, data)
}
export function rejectProjectArchive(id: number, data: {
  task_id: number
  comment: string
}) {
  return request.post(`/business/project-archive/${id}/reject`, data)
}

/* ---- Workflow integration helpers ---- */
export interface PendingTask {
  id: number
  node_code: string
  node_name: string
  next_node?: string
  next_node_name?: string
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