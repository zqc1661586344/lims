package service

import (
	"lims-backend/internal/workflow"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// WorkflowService wraps the workflow engine for workflow-level (taskID-based)
// operations. ApproveTask / RejectTask delegate to BusinessService so that
// every transition — regardless of caller entry point — atomically advances
// the process state machine AND ensures the target node's business record.
type WorkflowService struct {
	logger   *zap.Logger
	engine   *workflow.Engine
	business *BusinessService
}

// NewWorkflowService creates a new workflow service.
func NewWorkflowService(logger *zap.Logger, db *gorm.DB) *WorkflowService {
	return &WorkflowService{
		logger:   logger,
		engine:   workflow.DefaultEngine(db),
		business: NewBusinessService(logger, db),
	}
}

// StartInstance starts a new workflow process.
func (s *WorkflowService) StartInstance(businessType string, businessID uint, title string, createdBy uint) (uint, error) {
	return s.engine.StartInstance(businessType, businessID, title, createdBy)
}

// ApproveTask delegates to BusinessService.ApproveTask to guarantee business
// record creation for the next node.
func (s *WorkflowService) ApproveTask(taskID uint, userID uint, userDeptID uint, comment string) error {
	return s.business.ApproveTask(taskID, userID, userDeptID, comment)
}

// RejectTask delegates to BusinessService.RejectTask to guarantee business
// record creation for the reject-target node.
func (s *WorkflowService) RejectTask(taskID uint, userID uint, userDeptID uint, comment string, rejectTarget ...string) error {
	return s.business.RejectTask(taskID, userID, userDeptID, comment, rejectTarget...)
}

// GetPendingTasks returns pending tasks.
//   - isAdmin=true  → all departments' pending tasks (cross-department admin view)
//   - otherwise if deptID is set  → that department's pending tasks
//   - otherwise                  → current user's pending tasks
func (s *WorkflowService) GetPendingTasks(deptID *uint, userID uint, isAdmin bool) ([]workflow.PendingTaskDTO, error) {
	if isAdmin {
		return s.engine.GetAllPendingTasks()
	}
	if deptID != nil && *deptID > 0 {
		return s.engine.GetPendingTasksByDept(*deptID)
	}
	return s.engine.GetPendingTasksByUser(userID)
}

func (s *WorkflowService) GetProcessHistory(instanceID uint) ([]workflow.ProcessHistoryDTO, error) {
	return s.engine.GetProcessHistory(instanceID)
}

func (s *WorkflowService) GetInstance(instanceID uint) (*workflow.ProcessInstanceDTO, error) {
	return s.engine.GetInstance(instanceID)
}

func (s *WorkflowService) GetNodeDefinitions() []workflow.NodeDefinition {
	return workflow.GetDefinition()
}

func (s *WorkflowService) GetProgressByBusiness(businessType string, businessID uint, userID uint, userDeptID uint, isAdmin bool) (*workflow.ProgressDTO, error) {
	return s.engine.GetProgressByBusiness(businessType, businessID, userID, userDeptID, isAdmin)
}
