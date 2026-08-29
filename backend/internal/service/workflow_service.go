package service

import (
	"lims-backend/internal/workflow"

	"gorm.io/gorm"
)

// WorkflowService wraps the workflow engine with business logic.
type WorkflowService struct {
	engine *workflow.Engine
}

// NewWorkflowService creates a new workflow service.
func NewWorkflowService(db *gorm.DB) *WorkflowService {
	return &WorkflowService{
		engine: workflow.NewEngine(db),
	}
}

// StartInstance starts a new workflow process.
func (s *WorkflowService) StartInstance(businessType string, businessID uint, title string, createdBy uint) (uint, error) {
	return s.engine.StartInstance(businessType, businessID, title, createdBy)
}

// ApproveTask approves a pending task.
func (s *WorkflowService) ApproveTask(taskID uint, userID uint, comment string) error {
	return s.engine.ApproveTask(taskID, userID, comment)
}

// RejectTask rejects a pending task.
func (s *WorkflowService) RejectTask(taskID uint, userID uint, comment string) error {
	return s.engine.RejectTask(taskID, userID, comment)
}

// GetPendingTasks returns pending tasks.
//   - isAdmin=true  → all departments' pending tasks (cross-department admin view)
//   - otherwise, if deptID is set  → that department's pending tasks
//   - otherwise                  → current user's pending tasks
func (s *WorkflowService) GetPendingTasks(deptID *uint, userID uint, isAdmin bool) ([]map[string]interface{}, error) {
	if isAdmin {
		return s.engine.GetAllPendingTasks()
	}
	if deptID != nil && *deptID > 0 {
		return s.engine.GetPendingTasksByDept(*deptID)
	}
	return s.engine.GetPendingTasksByUser(userID)
}

// GetProcessHistory returns the task history for a process instance.
func (s *WorkflowService) GetProcessHistory(instanceID uint) ([]map[string]interface{}, error) {
	return s.engine.GetProcessHistory(instanceID)
}

// GetInstance returns a process instance by ID.
func (s *WorkflowService) GetInstance(instanceID uint) (map[string]interface{}, error) {
	return s.engine.GetInstance(instanceID)
}

// GetNodeDefinitions returns all workflow node definitions.
func (s *WorkflowService) GetNodeDefinitions() []workflow.NodeDefinition {
	return workflow.GetDefinition()
}