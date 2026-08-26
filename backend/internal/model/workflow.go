package model

import "time"

// ProcessInstance represents a workflow process instance.
type ProcessInstance struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	BusinessType string `gorm:"size:50;index;not null" json:"business_type"` // e.g. "task_order"
	BusinessID   uint   `gorm:"index;not null" json:"business_id"`
	Title        string `gorm:"size:200;not null" json:"title"`
	CurrentNode  string `gorm:"size:50;not null" json:"current_node"`       // current active node code
	Status       string `gorm:"size:20;default:running" json:"status"`      // running / completed / terminated
	CreatedBy    uint   `gorm:"index;not null" json:"created_by"`

	// Associations
	Tasks []ProcessTask `gorm:"foreignKey:ProcessInstanceID" json:"tasks,omitempty"`
}

func (ProcessInstance) TableName() string { return "process_instances" }

// ProcessTask represents a single task/node execution within a process instance.
type ProcessTask struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	ProcessInstanceID uint   `gorm:"index;not null" json:"process_instance_id"`
	NodeCode          string `gorm:"size:50;not null" json:"node_code"`
	NodeName          string `gorm:"size:100;not null" json:"node_name"`
	AssigneeDeptID    *uint  `gorm:"index" json:"assignee_dept_id"`
	AssigneeUserID    *uint  `gorm:"index" json:"assignee_user_id"`
	Status            string `gorm:"size:20;default:pending" json:"status"` // pending / completed / rejected / skipped
	Comment           string `gorm:"size:500" json:"comment"`
	OutputDocPath     string `gorm:"size:500" json:"output_doc_path"` // document produced at this node

	// Version for optimistic locking
	Version int `gorm:"default:0" json:"version"`
}

func (ProcessTask) TableName() string { return "process_tasks" }