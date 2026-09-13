package workflow

import "time"

// PendingTaskDTO is returned by GetPendingTasksByUser / ByDept / All.
type PendingTaskDTO struct {
	ID                uint      `json:"id"`
	NodeCode          string    `json:"node_code"`
	NodeName          string    `json:"node_name"`
	CreatedAt         time.Time `json:"created_at"`
	Title             string    `json:"title"`
	BusinessType      string    `json:"business_type"`
	BusinessID        uint      `json:"business_id"`
	ProcessInstanceID uint      `json:"process_instance_id"`

	AssigneeDeptID *uint  `json:"assignee_dept_id,omitempty"`
	DeptName       string `json:"dept_name,omitempty"`

	NextNode     string `json:"next_node,omitempty"`
	NextNodeName string `json:"next_node_name,omitempty"`
	NextDeptName string `json:"next_dept_name,omitempty"`
}

// ProcessHistoryDTO is returned by GetProcessHistory — shape that the
// frontend ProcessTimeline component expects.
type ProcessHistoryDTO struct {
	Name     string    `json:"name"`
	Status   string    `json:"status"`
	Comment  string    `json:"comment"`
	Operator string    `json:"operator"`
	Dept     string    `json:"dept"`
	Time     time.Time `json:"time"`
	Active   bool      `json:"active"`
	NodeCode string    `json:"node_code,omitempty"`
}

// ProcessInstanceDTO is returned by GetInstance.
type ProcessInstanceDTO struct {
	ID           uint      `json:"id"`
	BusinessType string    `json:"business_type"`
	BusinessID   uint      `json:"business_id"`
	Title        string    `json:"title"`
	CurrentNode  string    `json:"current_node"`
	Status       string    `json:"status"`
	CreatedBy    uint      `json:"created_by"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// NodeProgressDTO is one node within a ProgressDTO.
type NodeProgressDTO struct {
	Code      string `json:"code"`
	Name      string `json:"name"`
	DeptCode  string `json:"dept_code"`
	Index     int    `json:"index"`
	Status    string `json:"status"`
	CanReject bool   `json:"can_reject"`
	HasTask   bool   `json:"has_task"`
	Comment   string `json:"comment,omitempty"`
}

// ProgressDTO is returned by GetProgressByBusiness.
type ProgressDTO struct {
	InstanceID       uint              `json:"instance_id"`
	CurrentNode      string            `json:"current_node"`
	CurrentNodeIndex int               `json:"current_node_index"`
	TotalNodes       int               `json:"total_nodes"`
	Status           string            `json:"status"`
	Title            string            `json:"title"`
	CreatedAt        time.Time         `json:"created_at"`
	Nodes            []NodeProgressDTO `json:"nodes"`
}
