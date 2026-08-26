package model

import "time"

// TaskOrder 任务委托（节点1：任务创建）
type TaskOrder struct {
	ID                uint      `gorm:"primarykey" json:"id"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
	OrderNo           string    `gorm:"size:50;uniqueIndex;not null" json:"order_no"`
	CustomerName      string    `gorm:"size:200;not null" json:"customer_name"`
	ProjectName       string    `gorm:"size:200;not null" json:"project_name"`
	SampleType        string    `gorm:"size:100" json:"sample_type"`
	TestItems         string    `gorm:"type:jsonb" json:"test_items"`        // JSON: [{test_item_id, name, standard, method}]
	Status            int       `gorm:"default:0" json:"status"`             // 0=草稿, 1=已提交, 2=流程中, 3=已完成
	ProcessInstanceID *uint     `json:"process_instance_id"`                 // 关联流程实例（提交后生成）
	CreatedBy         *uint     `gorm:"index" json:"created_by"`
}

func (TaskOrder) TableName() string { return "task_orders" }

// ContractReview 合同评审（节点2）
type ContractReview struct {
	ID               uint      `gorm:"primarykey" json:"id"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	TaskOrderID      uint      `gorm:"uniqueIndex;not null" json:"task_order_id"`
	ReviewResult     string    `gorm:"size:20;not null" json:"review_result"`  // 通过/驳回
	ReviewComment    string    `gorm:"type:text" json:"review_comment"`
	ContractFilePath string    `gorm:"size:500" json:"contract_file_path"`
}

func (ContractReview) TableName() string { return "contract_reviews" }

// QCTask 质控任务（节点3）
type QCTask struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	TaskOrderID uint      `gorm:"uniqueIndex;not null" json:"task_order_id"`
	QCType      string    `gorm:"size:100" json:"qc_type"`
	QCDetails   string    `gorm:"type:jsonb" json:"qc_details"` // JSON: [{item_id, method, standard}]
}

func (QCTask) TableName() string { return "qc_tasks" }

// SamplingSchedule 采样调度（节点4）
type SamplingSchedule struct {
	ID             uint      `gorm:"primarykey" json:"id"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	TaskOrderID    uint      `gorm:"uniqueIndex;not null" json:"task_order_id"`
	SamplingTeam   string    `gorm:"size:200" json:"sampling_team"`
	SamplingPoints string    `gorm:"type:jsonb" json:"sampling_points"`  // JSON: [{location, type, count}]
	EquipmentList  string    `gorm:"type:jsonb" json:"equipment_list"`   // JSON: [{equipment_id, name, model}]
}

func (SamplingSchedule) TableName() string { return "sampling_schedules" }

// FieldSamplingRecord 现场采样（节点5）
type FieldSamplingRecord struct {
	ID                     uint      `gorm:"primarykey" json:"id"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
	TaskOrderID            uint      `gorm:"uniqueIndex;not null" json:"task_order_id"`
	SamplePhotos           string    `gorm:"type:jsonb" json:"sample_photos"`             // JSON: [{url, description}]
	EquipmentCalRecords    string    `gorm:"type:jsonb" json:"equipment_cal_records"`      // JSON: [{equipment_id, cal_result}]
	SamplingRecordFilePath string    `gorm:"size:500" json:"sampling_record_file_path"`
}

func (FieldSamplingRecord) TableName() string { return "field_sampling_records" }

// SampleReceiving 样品接收（节点6）
type SampleReceiving struct {
	ID                 uint      `gorm:"primarykey" json:"id"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
	TaskOrderID        uint      `gorm:"uniqueIndex;not null" json:"task_order_id"`
	SampleCondition    string    `gorm:"size:200" json:"sample_condition"`
	SampleCodes        string    `gorm:"type:jsonb" json:"sample_codes"`          // JSON: [{code, name, status}]
	ReceivingRecordPath string   `gorm:"size:500" json:"receiving_record_path"`
}

func (SampleReceiving) TableName() string { return "sample_receivings" }

// TaskAssign 任务分配（节点7，Phase 7 完整实现，预留模型）
type TaskAssign struct {
	ID            uint      `gorm:"primarykey" json:"id"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	TaskOrderID   uint      `gorm:"uniqueIndex;not null" json:"task_order_id"`
	AssignedTo    string    `gorm:"size:200" json:"assigned_to"`
	TestItemList  string    `gorm:"type:jsonb" json:"test_item_list"` // JSON: [{test_item_id, name}]
}

func (TaskAssign) TableName() string { return "task_assigns" }

// DataEntry 数据录入（节点8，Phase 7 完整实现，预留模型）
type DataEntry struct {
	ID           uint      `gorm:"primarykey" json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	TaskOrderID  uint      `gorm:"index;not null" json:"task_order_id"`
	TestItemID   uint      `json:"test_item_id"`
	OriginalData string    `gorm:"type:jsonb" json:"original_data"` // JSON: 原始数据
	RawRecordID  *uint     `json:"raw_record_id"`                   // 关联原始记录（Phase 7）
}

func (DataEntry) TableName() string { return "data_entries" }

// DataReview 数据复核（节点9）
type DataReview struct {
	ID            uint      `gorm:"primarykey" json:"id"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	TaskOrderID   uint      `gorm:"uniqueIndex;not null" json:"task_order_id"`
	ReviewResult  string    `gorm:"size:20;not null" json:"review_result"`  // 通过/驳回
	ReviewComment string    `gorm:"type:text" json:"review_comment"`
	IssuesFound   string    `gorm:"type:jsonb" json:"issues_found"`         // JSON: [{issue, severity}]
}

func (DataReview) TableName() string { return "data_reviews" }

// DataAudit 数据审核（节点10）
type DataAudit struct {
	ID           uint      `gorm:"primarykey" json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	TaskOrderID  uint      `gorm:"uniqueIndex;not null" json:"task_order_id"`
	AuditResult  string    `gorm:"size:20;not null" json:"audit_result"`   // 通过/驳回
	AuditComment string    `gorm:"type:text" json:"audit_comment"`
	IssueList    string    `gorm:"type:jsonb" json:"issue_list"`            // JSON: [{issue, resolution}]
}

func (DataAudit) TableName() string { return "data_audits" }

// ReportPrepare 报告编制（节点11）
type ReportPrepare struct {
	ID             uint      `gorm:"primarykey" json:"id"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	TaskOrderID    uint      `gorm:"uniqueIndex;not null" json:"task_order_id"`
	ReportTitle    string    `gorm:"size:200;not null" json:"report_title"`
	ReportContent  string    `gorm:"type:jsonb" json:"report_content"`    // JSON: 报告内容
	ReportFile     string    `gorm:"size:500" json:"report_file"`         // 报告文件路径
	Attachments    string    `gorm:"type:jsonb" json:"attachments"`       // JSON: [{name, url}]
}

func (ReportPrepare) TableName() string { return "report_prepares" }

// ReportReview 报告复核（节点12）
type ReportReview struct {
	ID             uint      `gorm:"primarykey" json:"id"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	TaskOrderID    uint      `gorm:"uniqueIndex;not null" json:"task_order_id"`
	ReviewResult   string    `gorm:"size:20;not null" json:"review_result"`  // 通过/驳回
	ReviewComment  string    `gorm:"type:text" json:"review_comment"`
	ReviewedItems  string    `gorm:"type:jsonb" json:"reviewed_items"`       // JSON: [{item, result}]
}

func (ReportReview) TableName() string { return "report_reviews" }

// ReportAudit 报告审核（节点13）
type ReportAudit struct {
	ID           uint      `gorm:"primarykey" json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	TaskOrderID  uint      `gorm:"uniqueIndex;not null" json:"task_order_id"`
	AuditResult  string    `gorm:"size:20;not null" json:"audit_result"`   // 通过/驳回
	AuditComment string    `gorm:"type:text" json:"audit_comment"`
	AuditIssues  string    `gorm:"type:jsonb" json:"audit_issues"`         // JSON: [{issue, severity, action}]
}

func (ReportAudit) TableName() string { return "report_audits" }