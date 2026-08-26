package workflow

// Node codes for the 16-node LIMS workflow.
// The naming follows: node_<stage>_<action>
const (
	NodeTaskCreate       = "node_task_create"        // 1  - 业务室 - 任务创建
	NodeContractReview   = "node_contract_review"    // 2  - 技术室 - 合同评审
	NodeQCTask           = "node_qc_task"            // 3  - 报告室 - 质控任务
	NodeSamplingSchedule = "node_sampling_schedule"  // 4  - 现场室 - 采样调度
	NodeFieldSampling    = "node_field_sampling"     // 5  - 现场室 - 现场采样
	NodeSampleReceiving  = "node_sample_receiving"   // 6  - 样品室 - 样品接收
	NodeTaskAssign       = "node_task_assign"        // 7  - 实验室 - 任务分配
	NodeDataEntry        = "node_data_entry"         // 8  - 实验室 - 数据录入
	NodeDataReview       = "node_data_review"        // 9  - 实验室 - 数据复核
	NodeDataAudit        = "node_data_audit"         // 10 - 实验室 - 数据审核
	NodeReportPrepare    = "node_report_prepare"     // 11 - 报告室 - 报告编制
	NodeReportReview     = "node_report_review"      // 12 - 实验室 - 报告复核
	NodeReportAudit      = "node_report_audit"       // 13 - 质控室 - 报告审核
	NodeReportSign       = "node_report_sign"        // 14 - 技术室 - 报告签发
	NodeReportPrint      = "node_report_print"       // 15 - 业务室 - 报告发放
	NodeProjectArchive   = "node_project_archive"    // 16 - 报告室 - 项目归档
)

// Department codes mapped from the 7 predefined departments.
const (
	DeptBusiness = "dept_business" // 业务室
	DeptTech     = "dept_tech"     // 技术室
	DeptReport   = "dept_report"   // 报告室
	DeptField    = "dept_field"    // 现场室
	DeptSample   = "dept_sample"   // 样品室
	DeptLab      = "dept_lab"      // 实验室
	DeptQC       = "dept_qc"       // 质控室
)

// Task statuses.
const (
	TaskStatusPending   = "pending"
	TaskStatusCompleted = "completed"
	TaskStatusRejected  = "rejected"
	TaskStatusSkipped   = "skipped"
)

// Instance statuses.
const (
	InstanceStatusRunning    = "running"
	InstanceStatusCompleted  = "completed"
	InstanceStatusTerminated = "terminated"
)

// AllNodeCodes returns all 16 node codes in order.
func AllNodeCodes() []string {
	return []string{
		NodeTaskCreate,
		NodeContractReview,
		NodeQCTask,
		NodeSamplingSchedule,
		NodeFieldSampling,
		NodeSampleReceiving,
		NodeTaskAssign,
		NodeDataEntry,
		NodeDataReview,
		NodeDataAudit,
		NodeReportPrepare,
		NodeReportReview,
		NodeReportAudit,
		NodeReportSign,
		NodeReportPrint,
		NodeProjectArchive,
	}
}