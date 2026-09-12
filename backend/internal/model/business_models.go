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
	TestItems         string    `gorm:"type:jsonb" json:"test_items"`
	Status            int       `gorm:"default:0" json:"status"`
	ProcessInstanceID *uint     `gorm:"constraint:OnDelete:SET NULL;references:process_instances(id)" json:"process_instance_id"`
	CreatedBy         *uint     `gorm:"index;constraint:OnDelete:SET NULL;references:users(id)" json:"created_by"`
}

func (TaskOrder) TableName() string { return "task_orders" }

// ContractReview 合同评审（节点2）
type ContractReview struct {
	ID               uint      `gorm:"primarykey" json:"id"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	TaskOrderID      uint      `gorm:"uniqueIndex;not null;constraint:OnDelete:CASCADE;references:task_orders(id)" json:"task_order_id"`
	ReviewResult     string    `gorm:"size:20;not null" json:"review_result"`
	ReviewComment    string    `gorm:"type:text" json:"review_comment"`
	ContractFilePath string    `gorm:"size:500" json:"contract_file_path"`
}

func (ContractReview) TableName() string { return "contract_reviews" }

// QCTask 质控任务（节点3）
type QCTask struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	TaskOrderID uint      `gorm:"uniqueIndex;not null;constraint:OnDelete:CASCADE;references:task_orders(id)" json:"task_order_id"`
	QCType      string    `gorm:"size:100" json:"qc_type"`
	QCDetails   string    `gorm:"type:jsonb" json:"qc_details"`
}

func (QCTask) TableName() string { return "qc_tasks" }

// SamplingSchedule 采样调度（节点4）
type SamplingSchedule struct {
	ID             uint      `gorm:"primarykey" json:"id"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	TaskOrderID    uint      `gorm:"uniqueIndex;not null;constraint:OnDelete:CASCADE;references:task_orders(id)" json:"task_order_id"`
	SamplingTeam   string    `gorm:"size:200" json:"sampling_team"`
	SamplingPoints string    `gorm:"type:jsonb" json:"sampling_points"`
	EquipmentList  string    `gorm:"type:jsonb" json:"equipment_list"`
}

func (SamplingSchedule) TableName() string { return "sampling_schedules" }

// FieldSamplingRecord 现场采样（节点5）
type FieldSamplingRecord struct {
	ID                     uint      `gorm:"primarykey" json:"id"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
	TaskOrderID            uint      `gorm:"uniqueIndex;not null;constraint:OnDelete:CASCADE;references:task_orders(id)" json:"task_order_id"`
	SamplePhotos           string    `gorm:"type:jsonb" json:"sample_photos"`
	EquipmentCalRecords    string    `gorm:"type:jsonb" json:"equipment_cal_records"`
	SamplingRecordFilePath string    `gorm:"size:500" json:"sampling_record_file_path"`
}

func (FieldSamplingRecord) TableName() string { return "field_sampling_records" }

// SampleReceiving 样品接收（节点6）
type SampleReceiving struct {
	ID                  uint      `gorm:"primarykey" json:"id"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
	TaskOrderID         uint      `gorm:"uniqueIndex;not null;constraint:OnDelete:CASCADE;references:task_orders(id)" json:"task_order_id"`
	SampleCondition     string    `gorm:"size:200" json:"sample_condition"`
	SampleCodes         string    `gorm:"type:jsonb" json:"sample_codes"`
	ReceivingRecordPath string    `gorm:"size:500" json:"receiving_record_path"`
}

func (SampleReceiving) TableName() string { return "sample_receivings" }

// TaskAssign 任务分配（节点7）
type TaskAssign struct {
	ID             uint      `gorm:"primarykey" json:"id"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	TaskOrderID    uint      `gorm:"uniqueIndex;not null;constraint:OnDelete:CASCADE;references:task_orders(id)" json:"task_order_id"`
	AssignedTo     string    `gorm:"size:200" json:"assigned_to"`
	AssigneeUserID *uint     `gorm:"index;constraint:OnDelete:SET NULL;references:users(id)" json:"assignee_user_id"`
	TestItemList   string    `gorm:"type:jsonb" json:"test_item_list"`
}

func (TaskAssign) TableName() string { return "task_assigns" }

// DataEntry 数据录入（节点8）
type DataEntry struct {
	ID           uint      `gorm:"primarykey" json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	TaskOrderID  uint      `gorm:"index;not null;constraint:OnDelete:CASCADE;references:task_orders(id)" json:"task_order_id"`
	TestItemID   uint      `json:"test_item_id"`
	OriginalData string    `gorm:"type:jsonb" json:"original_data"`
	RawRecordID  *uint     `json:"raw_record_id"`
}

func (DataEntry) TableName() string { return "data_entries" }

// DataReview 数据复核（节点9）
type DataReview struct {
	ID            uint      `gorm:"primarykey" json:"id"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	TaskOrderID   uint      `gorm:"uniqueIndex;not null;constraint:OnDelete:CASCADE;references:task_orders(id)" json:"task_order_id"`
	ReviewResult  string    `gorm:"size:20;not null" json:"review_result"`
	ReviewComment string    `gorm:"type:text" json:"review_comment"`
	IssuesFound   string    `gorm:"type:jsonb" json:"issues_found"`
}

func (DataReview) TableName() string { return "data_reviews" }

// DataAudit 数据审核（节点10）
type DataAudit struct {
	ID           uint      `gorm:"primarykey" json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	TaskOrderID  uint      `gorm:"uniqueIndex;not null;constraint:OnDelete:CASCADE;references:task_orders(id)" json:"task_order_id"`
	AuditResult  string    `gorm:"size:20;not null" json:"audit_result"`
	AuditComment string    `gorm:"type:text" json:"audit_comment"`
	IssueList    string    `gorm:"type:jsonb" json:"issue_list"`
}

func (DataAudit) TableName() string { return "data_audits" }

// ReportPrepare 报告编制（节点11）
type ReportPrepare struct {
	ID             uint      `gorm:"primarykey" json:"id"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	TaskOrderID    uint      `gorm:"uniqueIndex;not null;constraint:OnDelete:CASCADE;references:task_orders(id)" json:"task_order_id"`
	ReportNo       string    `gorm:"size:100" json:"report_no"`
	PrepareOpinion string    `gorm:"type:text" json:"prepare_opinion"`
	ReportTitle    string    `gorm:"size:200;not null" json:"report_title"`
	ReportContent  string    `gorm:"type:jsonb" json:"report_content"`
	ReportFile     string    `gorm:"size:500" json:"report_file"`
	Attachments    string    `gorm:"type:jsonb" json:"attachments"`
}

func (ReportPrepare) TableName() string { return "report_prepares" }

// ReportReview 报告复核（节点12）
type ReportReview struct {
	ID            uint      `gorm:"primarykey" json:"id"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	TaskOrderID   uint      `gorm:"uniqueIndex;not null;constraint:OnDelete:CASCADE;references:task_orders(id)" json:"task_order_id"`
	ReviewResult  string    `gorm:"size:20;not null" json:"review_result"`
	ReviewComment string    `gorm:"type:text" json:"review_comment"`
	ReviewedItems string    `gorm:"type:jsonb" json:"reviewed_items"`
}

func (ReportReview) TableName() string { return "report_reviews" }

// ReportAudit 报告审核（节点13）
type ReportAudit struct {
	ID           uint      `gorm:"primarykey" json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	TaskOrderID  uint      `gorm:"uniqueIndex;not null;constraint:OnDelete:CASCADE;references:task_orders(id)" json:"task_order_id"`
	AuditResult  string    `gorm:"size:20;not null" json:"audit_result"`
	AuditComment string    `gorm:"type:text" json:"audit_comment"`
	AuditIssues  string    `gorm:"type:jsonb" json:"audit_issues"`
}

func (ReportAudit) TableName() string { return "report_audits" }

// ReportSign 报告签发（节点14）
type ReportSign struct {
	ID             uint       `gorm:"primarykey" json:"id"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	TaskOrderID    uint       `gorm:"uniqueIndex;not null;constraint:OnDelete:CASCADE;references:task_orders(id)" json:"task_order_id"`
	ReportNo       string     `gorm:"size:100" json:"report_no"`
	ReportTitle    string     `gorm:"size:200" json:"report_title"`
	PrepareOpinion string     `gorm:"type:text" json:"prepare_opinion"`
	ReviewOpinion  string     `gorm:"type:text" json:"review_opinion"`
	AuditOpinion   string     `gorm:"type:text" json:"audit_opinion"`
	RawRecords     string     `gorm:"type:jsonb" json:"raw_records"`
	SignResult     string     `gorm:"size:20;not null" json:"sign_result"`
	SignComment    string     `gorm:"type:text" json:"sign_comment"`
	SignerName     string     `gorm:"size:100" json:"signer_name"`
	SignDate       *time.Time `json:"sign_date"`
	SignStamp      string     `gorm:"size:500" json:"sign_stamp"`
}

func (ReportSign) TableName() string { return "report_signs" }

// ReportPrint 报告打印发放（节点15）
type ReportPrint struct {
	ID             uint       `gorm:"primarykey" json:"id"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	TaskOrderID    uint       `gorm:"uniqueIndex;not null;constraint:OnDelete:CASCADE;references:task_orders(id)" json:"task_order_id"`
	PrintCount     int        `gorm:"default:1" json:"print_count"`
	PrintResult    string     `gorm:"size:20;not null" json:"print_result"`
	PrintComment   string     `gorm:"type:text" json:"print_comment"`
	RecipientName  string     `gorm:"size:100" json:"recipient_name"`
	RecipientDate  *time.Time `json:"recipient_date"`
	DeliveryMethod string     `gorm:"size:50" json:"delivery_method"`
	TrackingNo     string     `gorm:"size:100" json:"tracking_no"`
}

func (ReportPrint) TableName() string { return "report_prints" }

// ProjectArchive 项目归档（节点16，终节点）
type ProjectArchive struct {
	ID              uint       `gorm:"primarykey" json:"id"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	TaskOrderID     uint       `gorm:"uniqueIndex;not null;constraint:OnDelete:CASCADE;references:task_orders(id)" json:"task_order_id"`
	ArchiveNo       string     `gorm:"size:100;uniqueIndex" json:"archive_no"`
	ArchiveLocation string     `gorm:"size:200" json:"archive_location"`
	ArchiveDate     *time.Time `json:"archive_date"`
	ArchiveFiles    string     `gorm:"type:jsonb" json:"archive_files"`
	ArchiveComment  string     `gorm:"type:text" json:"archive_comment"`
	RetentionPeriod int        `gorm:"default:36" json:"retention_period"`
}

func (ProjectArchive) TableName() string { return "project_archives" }
