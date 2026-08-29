package service

import (
	"fmt"
	"lims-backend/internal/workflow"

	"gorm.io/gorm"
)

// BusinessService wraps the workflow engine for Phase 6 business flow operations.
type BusinessService struct {
	engine *workflow.Engine
	db     *gorm.DB
}

// NewBusinessService creates a new business service.
func NewBusinessService(db *gorm.DB) *BusinessService {
	return &BusinessService{
		engine: workflow.NewEngine(db),
		db:     db,
	}
}

// StartWorkflow starts a new process instance for a business entity.
func (s *BusinessService) StartWorkflow(businessType string, businessID uint, title string, createdBy uint) (uint, error) {
	return s.engine.StartInstance(businessType, businessID, title, createdBy)
}

// ApproveTask approves a workflow task by its process_tasks.id.
func (s *BusinessService) ApproveTask(taskID uint, userID uint, comment string) error {
	return s.engine.ApproveTask(taskID, userID, comment)
}

// RejectTask rejects a workflow task by its process_tasks.id.
func (s *BusinessService) RejectTask(taskID uint, userID uint, comment string) error {
	return s.engine.RejectTask(taskID, userID, comment)
}

// ApproveTaskByOrder approves the current pending task for a task order.
// The business frontend only knows the task_order_id, not the workflow task id,
// so we resolve the active pending task before approving.
func (s *BusinessService) ApproveTaskByOrder(orderID uint, userID uint, comment string) error {
	taskID, err := s.engine.GetPendingTaskIDByOrder(orderID)
	if err != nil {
		return err
	}
	if taskID == 0 {
		return fmt.Errorf("未找到该委托当前待办的任务(task_order_id=%d)", orderID)
	}
	return s.engine.ApproveTask(taskID, userID, comment)
}

// RejectTaskByOrder rejects the current pending task for a task order.
func (s *BusinessService) RejectTaskByOrder(orderID uint, userID uint, comment string) error {
	taskID, err := s.engine.GetPendingTaskIDByOrder(orderID)
	if err != nil {
		return err
	}
	if taskID == 0 {
		return fmt.Errorf("未找到该委托当前待办的任务(task_order_id=%d)", orderID)
	}
	return s.engine.RejectTask(taskID, userID, comment)
}