package service

import (
	"encoding/json"
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
func (s *BusinessService) ApproveTask(taskID uint, userID uint, userDeptID uint, comment string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		ctx, err := s.resolveTaskContext(tx, taskID)
		if err != nil {
			return err
		}
		nodeMap := workflow.BuildNodeMap()
		nextNode := nodeMap[ctx.NodeCode].NextNode

		if err := s.engine.ApproveTaskWithTx(tx, taskID, userID, userDeptID, comment); err != nil {
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

func (s *BusinessService) RejectTask(taskID uint, userID uint, userDeptID uint, comment string, rejectTarget ...string) error {
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

		if err := s.engine.RejectTaskWithTx(tx, taskID, userID, userDeptID, comment); err != nil {
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

func (s *BusinessService) ApproveTaskByOrder(orderID uint, userID uint, userDeptID uint, comment string) error {
	taskID, err := s.engine.GetPendingTaskIDByOrder(orderID)
	if err != nil {
		return err
	}
	if taskID == 0 {
		return fmt.Errorf("未找到该委托当前待办的任务(task_order_id=%d)", orderID)
	}
	return s.ApproveTask(taskID, userID, userDeptID, comment)
}

func (s *BusinessService) RejectTaskByOrder(orderID uint, userID uint, userDeptID uint, comment string, rejectTarget ...string) error {
	taskID, err := s.engine.GetPendingTaskIDByOrder(orderID)
	if err != nil {
		return err
	}
	if taskID == 0 {
		return fmt.Errorf("未找到该委托当前待办的任务(task_order_id=%d)", orderID)
	}
	return s.RejectTask(taskID, userID, userDeptID, comment, rejectTarget...)
}

func (s *BusinessService) ApproveWithBusiness(orderID uint, userID uint, userDeptID uint, comment string, save func(tx *gorm.DB) error) error {
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

		if err := s.engine.ApproveTaskWithTx(tx, taskID, userID, userDeptID, comment); err != nil {
			return err
		}
		if nextNode != "" && nextNode != workflow.NodeTaskCreate {
			return s.ensureBusinessRecord(tx, nextNode, ctx.BusinessID)
		}
		return nil
	})
}

func (s *BusinessService) RejectWithBusiness(orderID uint, userID uint, userDeptID uint, comment string, save func(tx *gorm.DB) error) error {
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

		if err := s.engine.RejectTaskWithTx(tx, taskID, userID, userDeptID, comment); err != nil {
			return err
		}
		if rejectTarget != "" && rejectTarget != workflow.NodeTaskCreate {
			return s.ensureBusinessRecord(tx, rejectTarget, ctx.BusinessID)
		}
		return nil
	})
}

func (s *BusinessService) CheckInstanceRunning(tx *gorm.DB, businessType string, businessID uint) error {
	var count int64
	if err := tx.Raw(`SELECT COUNT(*) FROM process_instances
		WHERE business_type = ? AND business_id = ? AND status = 'running'`,
		businessType, businessID).Scan(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("该记录关联运行中的流程，无法删除")
	}
	return nil
}

func (s *BusinessService) CheckNodeNotAdvanced(tx *gorm.DB, businessType string, businessID uint, nodeCode string) error {
	var currentNode string
	if err := tx.Raw(`SELECT current_node FROM process_instances
		WHERE business_type = ? AND business_id = ? AND status = 'running'
		LIMIT 1`, businessType, businessID).Scan(&currentNode).Error; err != nil {
		return err
	}
	if currentNode != "" && currentNode != nodeCode {
		return workflow.ErrTaskAlreadyApproved
	}
	return nil
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
		var order model.TaskOrder
		if err := tx.Select("id, test_items").First(&order, orderID).Error; err != nil {
			return err
		}

		testItemIDs, err := parseTestItemIDs(order.TestItems)
		if err != nil {
			s.logger.Warn("NodeDataEntry: parse test_items failed", zap.Uint("order_id", orderID), zap.Error(err))
		}

		if len(testItemIDs) == 0 {
			s.logger.Warn("NodeDataEntry: no test items defined on task order, skipping data entry creation",
				zap.Uint("order_id", orderID))
			if err := tx.Where("task_order_id = ? AND test_item_id = 0", orderID).Delete(&model.DataEntry{}).Error; err != nil {
				s.logger.Warn("NodeDataEntry: clean up old test_item_id=0 entries failed", zap.Error(err))
			}
			return nil
		}

		if err := tx.Where("task_order_id = ? AND test_item_id = 0", orderID).Delete(&model.DataEntry{}).Error; err != nil {
			s.logger.Warn("NodeDataEntry: clean up old test_item_id=0 entries failed", zap.Error(err))
		}

		for _, tid := range testItemIDs {
			var count int64
			if err := tx.Model(&model.DataEntry{}).
				Where("task_order_id = ? AND test_item_id = ?", orderID, tid).
				Count(&count).Error; err != nil {
				return err
			}
			if count > 0 {
				continue
			}
			if err := tx.Create(&model.DataEntry{
				TaskOrderID:  orderID,
				TestItemID:   tid,
				OriginalData: "{}",
			}).Error; err != nil {
				return err
			}
		}
		return s.ensureLabSheetsForOrder(tx, orderID, testItemIDs)
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

// parseTestItemIDs decodes TaskOrder.TestItems which is stored as a JSON array of
// JSON-encoded strings (frontend uses JSON.stringify per item). Example stored value:
//
//	["{\"test_item_id\":1,\"name\":\"pH\"}", "{\"test_item_id\":2,\"name\":\"COD\"}"]
func parseTestItemIDs(raw string) ([]uint, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	var outer []string
	if err := json.Unmarshal([]byte(raw), &outer); err != nil {
		var outer2 []map[string]interface{}
		if err2 := json.Unmarshal([]byte(raw), &outer2); err2 == nil {
			for _, m := range outer2 {
				if v, ok := m["test_item_id"].(float64); ok && uint(v) > 0 {
				}
			}
		}
		return nil, fmt.Errorf("outer: %w", err)
	}
	seen := make(map[uint]struct{})
	var ids []uint
	for _, s := range outer {
		var obj struct {
			TestItemID uint `json:"test_item_id"`
		}
		if err := json.Unmarshal([]byte(s), &obj); err != nil {
			continue
		}
		if obj.TestItemID > 0 {
			if _, dup := seen[obj.TestItemID]; !dup {
				seen[obj.TestItemID] = struct{}{}
				ids = append(ids, obj.TestItemID)
			}
		}
	}
	return ids, nil
}

// ensureLabSheetsForOrder creates LabSheet instances for every test item declared on
// the TaskOrder, using the latest template (if any) as the starting structure.
// Called automatically when the workflow advances to node_data_entry.
func (s *BusinessService) ensureLabSheetsForOrder(tx *gorm.DB, orderID uint, testItemIDs []uint) error {
	for _, it := range testItemIDs {
		if it == 0 {
			continue
		}
		var existing int64
		if err := tx.Model(&model.LabSheet{}).
			Where("task_order_id = ? AND test_item_id = ? AND node_code = ?", orderID, it, workflow.NodeDataEntry).
			Count(&existing).Error; err != nil {
			return err
		}
		if existing > 0 {
			continue
		}

		var tpl model.LabSheetTemplate
		var templateID *uint
		var sheetData json.RawMessage = []byte("{}")
		if err := tx.Where("test_item_id = ? AND status = 1", it).
			Order("version DESC").First(&tpl).Error; err == nil {
			templateID = &tpl.ID
			if len(tpl.Structure) > 0 {
				sheetData = tpl.Structure
			}
		}

		if err := tx.Create(&model.LabSheet{
			TaskOrderID: orderID,
			TestItemID:  it,
			TemplateID:  templateID,
			NodeCode:    workflow.NodeDataEntry,
			SheetData:   sheetData,
			Status:      0,
		}).Error; err != nil {
			return err
		}
	}
	return nil
}
