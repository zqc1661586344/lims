package service

import (
	"lims-backend/internal/workflow"

	"gorm.io/gorm"
)

// BusinessService wraps the workflow engine for Phase 6 business flow operations.
type BusinessService struct {
	engine *workflow.Engine
}

// NewBusinessService creates a new business service.
func NewBusinessService(db *gorm.DB) *BusinessService {
	return &BusinessService{
		engine: workflow.NewEngine(db),
	}
}

// StartWorkflow starts a new process instance for a business entity.
func (s *BusinessService) StartWorkflow(businessType string, businessID uint, title string, createdBy uint) (uint, error) {
	return s.engine.StartInstance(businessType, businessID, title, createdBy)
}

// ApproveTask approves a pending task.
func (s *BusinessService) ApproveTask(taskID uint, userID uint, comment string) error {
	return s.engine.ApproveTask(taskID, userID, comment)
}

// RejectTask rejects a pending task.
func (s *BusinessService) RejectTask(taskID uint, userID uint, comment string) error {
	return s.engine.RejectTask(taskID, userID, comment)
}

// GetPendingTasks returns pending tasks for a user or their department.
func (s *BusinessService) GetPendingTasks(deptID *uint, userID uint) ([]map[string]interface{}, error) {
	if deptID != nil && *deptID > 0 {
		return s.engine.GetPendingTasksByDept(*deptID)
	}
	return s.engine.GetPendingTasksByUser(userID)
}

// GetProcessHistory returns the task history for a process instance.
func (s *BusinessService) GetProcessHistory(instanceID uint) ([]map[string]interface{}, error) {
	return s.engine.GetProcessHistory(instanceID)
}

// GetInstance returns a process instance by ID.
func (s *BusinessService) GetInstance(instanceID uint) (map[string]interface{}, error) {
	return s.engine.GetInstance(instanceID)
}

// GetNodeDefinitions returns all workflow node definitions.
func (s *BusinessService) GetNodeDefinitions() []workflow.NodeDefinition {
	return workflow.GetDefinition()
}