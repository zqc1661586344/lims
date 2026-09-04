package service

import (
	"fmt"
	"lims-backend/internal/model"
	"lims-backend/internal/workflow"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// BusinessService wraps the workflow engine for Phase 6 business flow operations.
// It acts as the single Facade between business logic and the pure state-machine Engine,
// ensuring that every workflow transition (approve/reject) atomically advances the
// process state machine AND guarantees the target node's business table has at least
// one row (created via FirstOrCreate semantics).
type BusinessService struct {
	logger *zap.Logger
	engine *workflow.Engine
	db     *gorm.DB
}

// NewBusinessService creates a new business service.
func NewBusinessService(logger *zap.Logger, db *gorm.DB) *BusinessService {
	return &BusinessService{
		logger: logger,
		engine: workflow.DefaultEngine(db),
		db:     db,
	}
}

// StartWorkflow starts a new process instance for a business entity.
func (s *BusinessService) StartWorkflow(businessType string, businessID uint, title string, createdBy uint) (uint, error) {
	return s.engine.StartInstance(businessType, businessID, title, createdBy)
}

// ApproveTask approves a workflow task by its process_tasks.id.
// This is the unified low-level approve entry: it atomically advances the
// workflow state machine AND ensures the NEXT node's business table has a row.
func (s *BusinessService) ApproveTask(taskID uint, userID uint, comment string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		ctx, err := s.resolveTaskContext(tx, taskID)
		if err != nil {
			return err
		}
		nodeMap := workflow.BuildNodeMap()
		nextNode := nodeMap[ctx.NodeCode].NextNode

		if err := s.engine.ApproveTaskWithTx(tx, taskID, userID, comment); err != nil {
			return err
		}

		if nextNode != "" && nextNode != workflow.NodeTaskCreate {
			if err := s.ensureBusinessRecord(tx, nextNode, ctx.BusinessID); err != nil {
				return err
			}
		}
		return nil
	})
}

// RejectTask rejects a workflow task by its process_tasks.id.
// Same unified entry pattern: advances the workflow backward and ensures the
// reject-target node's business table has a row.
func (s *BusinessService) RejectTask(taskID uint, userID uint, comment string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		ctx, err := s.resolveTaskContext(tx, taskID)
		if err != nil {
			return err
		}
		nodeMap := workflow.BuildNodeMap()
		rejectTarget := nodeMap[ctx.NodeCode].RejectTarget
		if rejectTarget == "" {
			rejectTarget = workflow.FindPreviousNode(ctx.NodeCode)
		}

		if err := s.engine.RejectTaskWithTx(tx, taskID, userID, comment); err != nil {
			return err
		}

		if rejectTarget != "" && rejectTarget != workflow.NodeTaskCreate {
			if err := s.ensureBusinessRecord(tx, rejectTarget, ctx.BusinessID); err != nil {
				return err
			}
		}
		return nil
	})
}

// ApproveTaskByOrder resolves the current pending task for a task_order_id,
// then delegates to the unified ApproveTask entry.
func (s *BusinessService) ApproveTaskByOrder(orderID uint, userID uint, comment string) error {
	taskID, err := s.engine.GetPendingTaskIDByOrder(orderID)
	if err != nil {
		return err
	}
	if taskID == 0 {
		return fmt.Errorf("未找到该委托当前待办的任务(task_order_id=%d)", orderID)
	}
	return s.ApproveTask(taskID, userID, comment)
}

// RejectTaskByOrder resolves the current pending task for a task_order_id,
// then delegates to the unified RejectTask entry.
func (s *BusinessService) RejectTaskByOrder(orderID uint, userID uint, comment string) error {
	taskID, err := s.engine.GetPendingTaskIDByOrder(orderID)
	if err != nil {
		return err
	}
	if taskID == 0 {
		return fmt.Errorf("未找到该委托当前待办的任务(task_order_id=%d)", orderID)
	}
	return s.RejectTask(taskID, userID, comment)
}

// ApproveWithBusiness executes a business save callback and workflow approval
// inside a single transaction to guarantee data consistency.
// The callback is responsible for writing the CURRENT node's business record;
// the unified ApproveTask takes care of the ENGINE advance + NEXT node's record.
func (s *BusinessService) ApproveWithBusiness(orderID uint, userID uint, comment string, save func(tx *gorm.DB) error) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if save != nil {
			if err := save(tx); err != nil {
				return err
			}
		}
		taskID, err := s.engine.GetPendingTaskIDByOrderTx(tx, orderID)
		if err != nil {
			return err
		}
		if taskID == 0 {
			return fmt.Errorf("未找到该委托当前待办的任务(task_order_id=%d)", orderID)
		}

		ctx, err := s.resolveTaskContext(tx, taskID)
		if err != nil {
			return err
		}
		nodeMap := workflow.BuildNodeMap()
		nextNode := nodeMap[ctx.NodeCode].NextNode

		if err := s.engine.ApproveTaskWithTx(tx, taskID, userID, comment); err != nil {
			return err
		}
		if nextNode != "" && nextNode != workflow.NodeTaskCreate {
			return s.ensureBusinessRecord(tx, nextNode, ctx.BusinessID)
		}
		return nil
	})
}

// RejectWithBusiness mirrors ApproveWithBusiness for the reject path.
func (s *BusinessService) RejectWithBusiness(orderID uint, userID uint, comment string, save func(tx *gorm.DB) error) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if save != nil {
			if err := save(tx); err != nil {
				return err
			}
		}
		taskID, err := s.engine.GetPendingTaskIDByOrderTx(tx, orderID)
		if err != nil {
			return err
		}
		if taskID == 0 {
			return fmt.Errorf("未找到该委托当前待办的任务(task_order_id=%d)", orderID)
		}

		ctx, err := s.resolveTaskContext(tx, taskID)
		if err != nil {
			return err
		}
		nodeMap := workflow.BuildNodeMap()
		rejectTarget := nodeMap[ctx.NodeCode].RejectTarget
		if rejectTarget == "" {
			rejectTarget = workflow.FindPreviousNode(ctx.NodeCode)
		}

		if err := s.engine.RejectTaskWithTx(tx, taskID, userID, comment); err != nil {
			return err
		}
		if rejectTarget != "" && rejectTarget != workflow.NodeTaskCreate {
			return s.ensureBusinessRecord(tx, rejectTarget, ctx.BusinessID)
		}
		return nil
	})
}

// taskContext holds the minimal workflow/instance context resolved from a taskID.
type taskContext struct {
	NodeCode     string
	InstanceID   uint
	BusinessType string
	BusinessID   uint
}

// resolveTaskContext looks up node_code + business context for a given taskID
// via a simple JOIN across process_tasks / process_instances.
func (s *BusinessService) resolveTaskContext(tx *gorm.DB, taskID uint) (*taskContext, error) {
	var ctx taskContext
	err := tx.Raw(`
		SELECT pt.node_code, pi.id AS instance_id, pi.business_type, pi.business_id
		FROM process_tasks pt
		JOIN process_instances pi ON pi.id = pt.process_instance_id
		WHERE pt.id = ? AND pt.status = 'pending'
	`, taskID).Scan(&ctx).Error
	if err != nil {
		return nil, err
	}
	if ctx.InstanceID == 0 {
		return nil, fmt.Errorf("流程任务不存在或已完成 (task_id=%d)", taskID)
	}
	return &ctx, nil
}

// ensureBusinessRecord guarantees the business table for the given node_code
// has at least one row with the specified task_order_id. Uses Gorm's native
// FirstOrCreate for tables with unique task_order_id (all except data_entries).
// If a row already exists it is left untouched; missing rows are inserted.
// All jsonb-typed fields must be explicitly initialized to "{}" because
// PostgreSQL rejects Go's empty-string zero value for jsonb columns.
func (s *BusinessService) ensureBusinessRecord(tx *gorm.DB, nodeCode string, orderID uint) error {
	switch nodeCode {
	case workflow.NodeContractReview:
		var m model.ContractReview
		return tx.Where("task_order_id = ?", orderID).FirstOrCreate(&m, model.ContractReview{TaskOrderID: orderID}).Error
	case workflow.NodeQCTask:
		var m model.QCTask
		return tx.Where("task_order_id = ?", orderID).FirstOrCreate(&m, model.QCTask{
			TaskOrderID: orderID,
			QCDetails:   "{}",
		}).Error
	case workflow.NodeSamplingSchedule:
		var m model.SamplingSchedule
		return tx.Where("task_order_id = ?", orderID).FirstOrCreate(&m, model.SamplingSchedule{
			TaskOrderID:    orderID,
			SamplingPoints: "{}",
			EquipmentList:  "{}",
		}).Error
	case workflow.NodeFieldSampling:
		var m model.FieldSamplingRecord
		return tx.Where("task_order_id = ?", orderID).FirstOrCreate(&m, model.FieldSamplingRecord{
			TaskOrderID:         orderID,
			SamplePhotos:        "{}",
			EquipmentCalRecords: "{}",
		}).Error
	case workflow.NodeSampleReceiving:
		var m model.SampleReceiving
		return tx.Where("task_order_id = ?", orderID).FirstOrCreate(&m, model.SampleReceiving{
			TaskOrderID: orderID,
			SampleCodes: "{}",
		}).Error
	case workflow.NodeTaskAssign:
		var m model.TaskAssign
		return tx.Where("task_order_id = ?", orderID).FirstOrCreate(&m, model.TaskAssign{
			TaskOrderID:  orderID,
			TestItemList: "{}",
		}).Error
	case workflow.NodeDataEntry:
		var count int64
		if err := tx.Model(&model.DataEntry{}).Where("task_order_id = ?", orderID).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return nil
		}
		return tx.Create(&model.DataEntry{TaskOrderID: orderID, OriginalData: "{}"}).Error
	case workflow.NodeDataReview:
		var m model.DataReview
		return tx.Where("task_order_id = ?", orderID).FirstOrCreate(&m, model.DataReview{
			TaskOrderID: orderID,
			IssuesFound: "{}",
		}).Error
	case workflow.NodeDataAudit:
		var m model.DataAudit
		return tx.Where("task_order_id = ?", orderID).FirstOrCreate(&m, model.DataAudit{
			TaskOrderID: orderID,
			IssueList:   "{}",
		}).Error
	case workflow.NodeReportPrepare:
		var m model.ReportPrepare
		return tx.Where("task_order_id = ?", orderID).FirstOrCreate(&m, model.ReportPrepare{
			TaskOrderID:   orderID,
			ReportTitle:   "待编制",
			ReportContent: "{}",
			Attachments:   "{}",
		}).Error
	case workflow.NodeReportReview:
		var m model.ReportReview
		return tx.Where("task_order_id = ?", orderID).FirstOrCreate(&m, model.ReportReview{
			TaskOrderID:   orderID,
			ReviewedItems: "{}",
		}).Error
	case workflow.NodeReportAudit:
		var m model.ReportAudit
		return tx.Where("task_order_id = ?", orderID).FirstOrCreate(&m, model.ReportAudit{
			TaskOrderID: orderID,
			AuditIssues: "{}",
		}).Error
	case workflow.NodeReportSign:
		var m model.ReportSign
		return tx.Where("task_order_id = ?", orderID).FirstOrCreate(&m, model.ReportSign{
			TaskOrderID: orderID,
			RawRecords:  "{}",
		}).Error
	case workflow.NodeReportPrint:
		var m model.ReportPrint
		return tx.Where("task_order_id = ?", orderID).FirstOrCreate(&m, model.ReportPrint{TaskOrderID: orderID}).Error
	case workflow.NodeProjectArchive:
		var m model.ProjectArchive
		return tx.Where("task_order_id = ?", orderID).FirstOrCreate(&m, model.ProjectArchive{
			TaskOrderID:  orderID,
			ArchiveFiles: "{}",
		}).Error
	}
	return nil
}
