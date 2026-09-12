package model

import "time"

// ProcessInstance represents a workflow process instance.
type ProcessInstance struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	BusinessType string `gorm:"size:50;index;not null" json:"business_type"`
	BusinessID   uint   `gorm:"index;not null" json:"business_id"`
	Title        string `gorm:"size:200;not null" json:"title"`
	CurrentNode  string `gorm:"size:50;not null" json:"current_node"`
	Status       string `gorm:"size:20;default:running" json:"status"`
	CreatedBy    uint   `gorm:"index;not null;constraint:OnDelete:RESTRICT;references:users(id)" json:"created_by"`

	Tasks []ProcessTask `gorm:"foreignKey:ProcessInstanceID;constraint:OnDelete:CASCADE" json:"tasks,omitempty"`
}

func (ProcessInstance) TableName() string { return "process_instances" }

// ProcessTask represents a single task/node execution within a process instance.
type ProcessTask struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	ProcessInstanceID uint   `gorm:"index;not null;constraint:OnDelete:CASCADE;references:process_instances(id)" json:"process_instance_id"`
	NodeCode          string `gorm:"size:50;not null" json:"node_code"`
	NodeName          string `gorm:"size:100;not null" json:"node_name"`
	AssigneeDeptID    *uint  `gorm:"index;constraint:OnDelete:SET NULL;references:depts(id)" json:"assignee_dept_id"`
	AssigneeUserID    *uint  `gorm:"index;constraint:OnDelete:SET NULL;references:users(id)" json:"assignee_user_id"`
	Status            string `gorm:"size:20;default:pending" json:"status"`
	Comment           string `gorm:"size:500" json:"comment"`
	OutputDocPath     string `gorm:"size:500" json:"output_doc_path"`

	Version int `gorm:"default:0" json:"version"`
}

func (ProcessTask) TableName() string { return "process_tasks" }
