package workflow

import (
	"errors"
	"fmt"
	"lims-backend/internal/model"
	"sync"

	"gorm.io/gorm"
)

// BusinessStatusSyncFn is a callback the engine invokes inside its own
// transaction after each workflow transition, so the business layer can
// keep its own business-table status column (e.g. task_orders.status) in
// sync without the engine knowing anything about business schema.
//
// event values:
//
//	"advance"  — task approved, instance advanced to next running node
//	"complete" — last node approved, instance finished
//	"reject"   — task rejected, instance rewound to earlier node
type BusinessStatusSyncFn func(tx *gorm.DB, businessType string, businessID uint, event string) error

// Engine is the core workflow state machine engine.
type Engine struct {
	db       *gorm.DB
	nodeDefs []NodeDefinition
	nodeMap  map[string]NodeDefinition
	syncBiz  BusinessStatusSyncFn
}

// SetBusinessSync registers the business-layer callback that keeps business
// table status columns aligned with workflow transitions. Safe to call at any
// time; the latest registration wins.
func (e *Engine) SetBusinessSync(fn BusinessStatusSyncFn) {
	e.syncBiz = fn
}

var (
	defaultEngineOnce sync.Once
	defaultEngine     *Engine
	ErrOptimisticLock = errors.New("optimistic lock conflict — please retry")
)

// sodCheckNodes maps nodes that enforce Segregation of Duties (CNAS
// requirement: reviewer != performer) to their immediately-preceding
// node whose assignee must differ from the current operator.
var sodCheckNodes = map[string]string{
	NodeDataReview:   NodeDataEntry,
	NodeDataAudit:    NodeDataReview,
	NodeReportReview: NodeReportPrepare,
	NodeReportAudit:  NodeReportReview,
}

// NewEngine creates a new workflow engine with the given GORM DB.
func NewEngine(db *gorm.DB) *Engine {
	defs := GetDefinition()
	return &Engine{
		db:       db,
		nodeDefs: defs,
		nodeMap:  BuildNodeMap(),
	}
}

// DefaultEngine returns a process-wide singleton Engine.
// The first call lazily initializes it; subsequent calls return the same instance.
func DefaultEngine(db *gorm.DB) *Engine {
	defaultEngineOnce.Do(func() {
		defaultEngine = NewEngine(db)
	})
	return defaultEngine
}

// StartInstance creates a new process instance and its first task.
// Returns the instance ID.
func (e *Engine) StartInstance(businessType string, businessID uint, title string, createdBy uint) (uint, error) {
	if len(e.nodeDefs) == 0 {
		return 0, errors.New("no workflow definition loaded")
	}
	firstNode := e.nodeDefs[0]

	instance := &model.ProcessInstance{
		BusinessType: businessType,
		BusinessID:   businessID,
		Title:        title,
		CurrentNode:  firstNode.Code,
		Status:       InstanceStatusRunning,
		CreatedBy:    createdBy,
	}

	if err := e.db.Create(instance).Error; err != nil {
		return 0, fmt.Errorf("create process instance: %w", err)
	}

	if err := e.createTask(instance.ID, firstNode.Code, firstNode.Name, 0); err != nil {
		return 0, fmt.Errorf("create first task: %w", err)
	}

	return instance.ID, nil
}

// ApproveTask approves a pending task, creating the next node's task.
// If the approved node is the terminal node, the instance is marked completed.
func (e *Engine) ApproveTask(taskID uint, userID uint, userDeptID uint, comment string) error {
	return e.db.Transaction(func(tx *gorm.DB) error {
		return e.ApproveTaskWithTx(tx, taskID, userID, userDeptID, comment)
	})
}

func (e *Engine) ApproveTaskWithTx(tx *gorm.DB, taskID uint, userID uint, userDeptID uint, comment string) error {
	task, err := e.getTaskForUpdate(tx, taskID)
	if err != nil {
		return err
	}
	if task.Status != TaskStatusPending {
		return ErrTaskAlreadyCompleted
	}
	if task.AssigneeDeptID != userDeptID {
		return ErrDeptNotMatch
	}
	if err := e.checkSoD(tx, task.ProcessInstanceID, task.NodeCode, userID); err != nil {
		return err
	}

	var instance struct {
		ID           uint
		CurrentNode  string
		Status       string
		BusinessType string
		BusinessID   uint
	}
	if err := tx.Raw(`SELECT id, current_node, status, business_type, business_id
		FROM process_instances WHERE id=? FOR UPDATE`, task.ProcessInstanceID).Scan(&instance).Error; err != nil {
		return fmt.Errorf("get instance: %w", err)
	}
	if instance.Status != InstanceStatusRunning {
		return ErrInstanceNotRunning
	}

	nodeDef, ok := e.nodeMap[task.NodeCode]
	if !ok {
		return ErrUnknownNode
	}

	result := tx.Exec(`UPDATE process_tasks SET status=?, comment=?, updated_at=NOW(), version=version+1
		WHERE id=? AND version=?`,
		TaskStatusCompleted, comment, taskID, task.Version)
	if result.Error != nil {
		return fmt.Errorf("update task: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrOptimisticLock
	}

	nextNode := nodeDef.NextNode
	if nextNode == "" {
		if err := tx.Exec(`UPDATE process_instances SET status=?, current_node=?, updated_at=NOW()
			WHERE id=?`, InstanceStatusCompleted, task.NodeCode, instance.ID).Error; err != nil {
			return fmt.Errorf("complete instance: %w", err)
		}
		if e.syncBiz != nil {
			if err := e.syncBiz(tx, instance.BusinessType, instance.BusinessID, "complete"); err != nil {
				return fmt.Errorf("business sync complete: %w", err)
			}
		}
	} else {
		nextDef, ok := e.nodeMap[nextNode]
		if !ok {
			return ErrUnknownNode
		}
		if err := e.createTaskWithTx(tx, instance.ID, nextNode, nextDef.Name, 0); err != nil {
			return err
		}
		if err := tx.Exec(`UPDATE process_instances SET current_node=?, updated_at=NOW()
			WHERE id=?`, nextNode, instance.ID).Error; err != nil {
			return fmt.Errorf("advance instance: %w", err)
		}
		if e.syncBiz != nil {
			if err := e.syncBiz(tx, instance.BusinessType, instance.BusinessID, "advance"); err != nil {
				return fmt.Errorf("business sync advance: %w", err)
			}
		}
	}
	return nil
}

func (e *Engine) RejectTask(taskID uint, userID uint, userDeptID uint, comment string, rejectTarget ...string) error {
	return e.db.Transaction(func(tx *gorm.DB) error {
		return e.RejectTaskWithTx(tx, taskID, userID, userDeptID, comment, rejectTarget...)
	})
}

func (e *Engine) RejectTaskWithTx(tx *gorm.DB, taskID uint, userID uint, userDeptID uint, comment string, rejectTargetOverride ...string) error {
	task, err := e.getTaskForUpdate(tx, taskID)
	if err != nil {
		return err
	}
	if task.Status != TaskStatusPending {
		return ErrTaskAlreadyCompleted
	}
	if task.AssigneeDeptID != userDeptID {
		return ErrDeptNotMatch
	}
	if err := e.checkSoD(tx, task.ProcessInstanceID, task.NodeCode, userID); err != nil {
		return err
	}

	nodeDef, ok := e.nodeMap[task.NodeCode]
	if !ok {
		return ErrUnknownNode
	}
	if !nodeDef.CanReject {
		return ErrCannotRejectFinalNode
	}

	rejectTarget := nodeDef.RejectTarget
	if len(rejectTargetOverride) > 0 && rejectTargetOverride[0] != "" {
		rejectTarget = rejectTargetOverride[0]
	}
	if rejectTarget == "" {
		return ErrCannotRejectFromFirstNode
	}

	var instance struct {
		ID           uint
		BusinessType string
		BusinessID   uint
	}
	if err := tx.Raw(`SELECT id, business_type, business_id FROM process_instances WHERE id=? FOR UPDATE`,
		task.ProcessInstanceID).Scan(&instance).Error; err != nil {
		return fmt.Errorf("get instance: %w", err)
	}

	result := tx.Exec(`UPDATE process_tasks SET status=?, comment=?, updated_at=NOW(), version=version+1
		WHERE id=? AND version=?`,
		TaskStatusRejected, comment, taskID, task.Version)
	if result.Error != nil {
		return fmt.Errorf("update task rejected: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrOptimisticLock
	}

	targetDef, ok := e.nodeMap[rejectTarget]
	if !ok {
		return ErrUnknownNode
	}
	if err := e.createTaskWithTx(tx, instance.ID, rejectTarget, targetDef.Name, 0); err != nil {
		return err
	}

	if err := tx.Exec(`UPDATE process_instances SET current_node=?, updated_at=NOW()
		WHERE id=?`, rejectTarget, instance.ID).Error; err != nil {
		return fmt.Errorf("revert instance node: %w", err)
	}

	if e.syncBiz != nil {
		if err := e.syncBiz(tx, instance.BusinessType, instance.BusinessID, "reject"); err != nil {
			return fmt.Errorf("business sync reject: %w", err)
		}
	}
	return nil
}

// GetPendingTasksByUser returns all pending tasks assigned to a specific user.
func (e *Engine) GetPendingTasksByUser(userID uint) ([]map[string]interface{}, error) {
	tasks, err := e.queryTasks(`
		SELECT pt.id, pt.node_code, pt.node_name, pt.created_at,
			pi.title, pi.business_type, pi.business_id, pi.id as process_instance_id
		FROM process_tasks pt
		JOIN process_instances pi ON pi.id = pt.process_instance_id
		WHERE pt.status = 'pending' AND pt.assignee_user_id = ?
		ORDER BY pt.created_at DESC`, userID)
	e.attachNextNode(tasks)
	return tasks, err
}

// GetPendingTasksByDept returns all pending tasks assigned to a department.
func (e *Engine) GetPendingTasksByDept(deptID uint) ([]map[string]interface{}, error) {
	tasks, err := e.queryTasks(`
		SELECT pt.id, pt.node_code, pt.node_name, pt.created_at,
			pi.title, pi.business_type, pi.business_id, pi.id as process_instance_id,
			d.name as dept_name
		FROM process_tasks pt
		JOIN process_instances pi ON pi.id = pt.process_instance_id
		LEFT JOIN depts d ON d.id = pt.assignee_dept_id
		WHERE pt.status = 'pending' AND pt.assignee_dept_id = ?
		ORDER BY pt.created_at DESC`, deptID)
	e.attachNextNode(tasks)
	return tasks, err
}

// GetAllPendingTasks returns ALL pending tasks across every department.
// Used by admin (cross-department view). Non-admin users should never call this.
func (e *Engine) GetAllPendingTasks() ([]map[string]interface{}, error) {
	tasks, err := e.queryTasks(`
		SELECT pt.id, pt.node_code, pt.node_name, pt.created_at,
			pi.title, pi.business_type, pi.business_id, pi.id as process_instance_id,
			pt.assignee_dept_id, d.name as dept_name
		FROM process_tasks pt
		JOIN process_instances pi ON pi.id = pt.process_instance_id
		LEFT JOIN depts d ON d.id = pt.assignee_dept_id
		WHERE pt.status = 'pending'
		ORDER BY pt.created_at DESC`)
	e.attachNextNode(tasks)
	return tasks, err
}

// attachNextNode appends next_node / next_node_name / next_dept_name to each pending task row.
// The next node (from the workflow definition) is displayed on the frontend so
// operators know which node the task advances to after approval.
func (e *Engine) attachNextNode(tasks []map[string]interface{}) {
	deptNames := e.deptNameMap()
	for _, t := range tasks {
		code, _ := t["node_code"].(string)
		if def, ok := e.nodeMap[code]; ok && def.NextNode != "" {
			if next, ok2 := e.nodeMap[def.NextNode]; ok2 {
				t["next_node"] = next.Code
				t["next_node_name"] = next.Name
				if name, ok3 := deptNames[next.DeptCode]; ok3 {
					t["next_dept_name"] = name
				}
			}
		}
	}
}

// deptNameMap loads the depts table (code -> name) once per call so the next
// node's responsible department can be shown on the frontend.
func (e *Engine) deptNameMap() map[string]string {
	type deptRow struct {
		Code string
		Name string
	}
	var rows []deptRow
	e.db.Raw(`SELECT code, name FROM depts`).Scan(&rows)
	m := make(map[string]string, len(rows))
	for _, r := range rows {
		m[r.Code] = r.Name
	}
	return m
}

// getPendingTaskIDByBusiness resolves the currently-pending workflow task ID
// for a business entity (e.g. a task order). It joins the running process
// instance for that business and finds its active pending task.
// Returns 0 (no error) if no pending task is found.

// AssignTask assigns a pending task to a specific user.
func (e *Engine) AssignTask(taskID uint, userID uint) error {
	return e.db.Transaction(func(tx *gorm.DB) error {
		return e.AssignTaskWithTx(tx, taskID, userID)
	})
}

func (e *Engine) AssignTaskWithTx(tx *gorm.DB, taskID uint, userID uint) error {
	var task struct {
		ID     uint
		Status string
	}
	if err := tx.Raw(`SELECT id, status FROM process_tasks WHERE id=? FOR UPDATE`, taskID).Scan(&task).Error; err != nil {
		return fmt.Errorf("get task: %w", err)
	}
	if task.ID == 0 {
		return fmt.Errorf("task not found: %d", taskID)
	}
	if task.Status != TaskStatusPending {
		return ErrTaskAlreadyCompleted
	}
	if err := tx.Exec(`UPDATE process_tasks SET assignee_user_id=? WHERE id=?`, userID, taskID).Error; err != nil {
		return fmt.Errorf("assign task: %w", err)
	}
	return nil
}

func (e *Engine) getPendingTaskIDByBusiness(businessType string, businessID uint) (uint, error) {
	return e.getPendingTaskIDByBusinessWithTx(e.db, businessType, businessID)
}

func (e *Engine) getPendingTaskIDByBusinessWithTx(tx *gorm.DB, businessType string, businessID uint) (uint, error) {
	var taskID uint
	err := tx.Raw(`
		SELECT pt.id
		FROM process_tasks pt
		JOIN process_instances pi ON pi.id = pt.process_instance_id
		WHERE pi.business_type = ? AND pi.business_id = ?
			AND pi.status = 'running'
			AND pt.status = 'pending'
			AND pt.node_code = pi.current_node
		ORDER BY pt.id DESC
		LIMIT 1`, businessType, businessID).Scan(&taskID).Error
	if err != nil {
		return 0, fmt.Errorf("resolve pending task: %w", err)
	}
	return taskID, nil
}

// GetPendingTaskIDByOrder resolves the pending workflow task ID for a task order.
// Exposed for business appovals that only know the task_order_id.
func (e *Engine) GetPendingTaskIDByOrder(orderID uint) (uint, error) {
	return e.getPendingTaskIDByBusiness("task_order", orderID)
}

// GetPendingTaskIDByOrderTx is the transaction-aware version of GetPendingTaskIDByOrder.
func (e *Engine) GetPendingTaskIDByOrderTx(tx *gorm.DB, orderID uint) (uint, error) {
	return e.getPendingTaskIDByBusinessWithTx(tx, "task_order", orderID)
}

// GetProcessHistory returns the full task history for a process instance.
// The returned fields are aliased to match the frontend TimelineNode shape
// (name/time/status/active/operator/dept/comment) so no client-side mapping is needed.
func (e *Engine) GetProcessHistory(instanceID uint) ([]map[string]interface{}, error) {
	history, err := e.queryTasks(`
		SELECT pt.id, pt.node_code, pt.node_name AS name, pt.status, pt.comment,
			pt.assignee_user_id, u.username AS operator, d.name AS dept,
			pt.created_at, pi.current_node
		FROM process_tasks pt
		JOIN process_instances pi ON pi.id = pt.process_instance_id
		LEFT JOIN users u ON u.id = pt.assignee_user_id
		LEFT JOIN depts d ON d.id = pt.assignee_dept_id
		WHERE pt.process_instance_id = ?
		ORDER BY pt.id ASC`, instanceID)

	// Normalize into the shape the frontend ProcessTimeline expects.
	for _, row := range history {
		row["time"] = row["created_at"]
		delete(row, "created_at")
		row["active"] = row["current_node"] == row["node_code"]
		delete(row, "current_node")
		delete(row, "node_code")
		delete(row, "id")
		delete(row, "assignee_user_id")
	}
	return history, err
}

// GetInstance returns a process instance by ID.
func (e *Engine) GetInstance(instanceID uint) (map[string]interface{}, error) {
	var result map[string]interface{}
	rows, err := e.db.Raw(`SELECT id, business_type, business_id, title,
		current_node, status, created_by, created_at, updated_at
		FROM process_instances WHERE id=?`, instanceID).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		result = make(map[string]interface{})
		cols, _ := rows.Columns()
		vals := make([]interface{}, len(cols))
		for i := range vals {
			vals[i] = new(interface{})
		}
		rows.Scan(vals...)
		for i, col := range cols {
			result[col] = *(vals[i].(*interface{}))
		}
	}

	return result, nil
}

// GetProgressByBusiness returns the workflow progress for a given business object.
// Returns instance_id, current_node, overall status, and all 16 nodes with their
// execution status (completed/current/pending/rejected).
func (e *Engine) GetProgressByBusiness(businessType string, businessID uint) (map[string]interface{}, error) {
	var instance model.ProcessInstance
	if err := e.db.Where("business_type = ? AND business_id = ?", businessType, businessID).
		First(&instance).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	nodeDefs := GetDefinition()
	taskStatusMap := make(map[string]string)
	taskInfoMap := make(map[string]map[string]interface{})

	var tasks []model.ProcessTask
	e.db.Where("process_instance_id = ?", instance.ID).Find(&tasks)
	for _, t := range tasks {
		taskStatusMap[t.NodeCode] = t.Status
		taskInfoMap[t.NodeCode] = map[string]interface{}{
			"status":  t.Status,
			"comment": t.Comment,
		}
	}

	nodes := make([]map[string]interface{}, 0, len(nodeDefs))
	currentNodeIndex := -1
	for i, def := range nodeDefs {
		nodeStatus := "pending"
		if instance.CurrentNode == def.Code {
			nodeStatus = "current"
			currentNodeIndex = i
		} else if st, ok := taskStatusMap[def.Code]; ok {
			if st == "completed" {
				nodeStatus = "completed"
			} else if st == "rejected" {
				nodeStatus = "rejected"
			} else if st == "skipped" {
				nodeStatus = "skipped"
			}
		}

		node := map[string]interface{}{
			"code":       def.Code,
			"name":       def.Name,
			"dept_code":  def.DeptCode,
			"index":      i,
			"status":     nodeStatus,
			"can_reject": def.CanReject,
			"has_task":   taskStatusMap[def.Code] != "",
		}
		if info, ok := taskInfoMap[def.Code]; ok {
			if c, ok := info["comment"]; ok {
				node["comment"] = c
			}
		}
		nodes = append(nodes, node)
	}

	progress := map[string]interface{}{
		"instance_id":        instance.ID,
		"current_node":       instance.CurrentNode,
		"current_node_index": currentNodeIndex,
		"total_nodes":        len(nodeDefs),
		"status":             instance.Status,
		"title":              instance.Title,
		"created_at":         instance.CreatedAt,
		"nodes":              nodes,
	}
	return progress, nil
}

// --- internal helpers ---

type taskRow struct {
	ID                uint
	ProcessInstanceID uint
	NodeCode          string
	Status            string
	Version           int
	AssigneeDeptID    uint
}

func (e *Engine) getTaskForUpdate(tx *gorm.DB, taskID uint) (*taskRow, error) {
	var row taskRow
	if err := tx.Raw(`SELECT id, process_instance_id, node_code, status, version,
		COALESCE(assignee_dept_id, 0) as assignee_dept_id
		FROM process_tasks WHERE id=? FOR UPDATE`, taskID).Scan(&row).Error; err != nil {
		return nil, fmt.Errorf("get task: %w", err)
	}
	if row.ID == 0 {
		return nil, fmt.Errorf("task not found: %d", taskID)
	}
	return &row, nil
}

func (e *Engine) checkSoD(tx *gorm.DB, instanceID uint, nodeCode string, userID uint) error {
	prevNode, ok := sodCheckNodes[nodeCode]
	if !ok {
		return nil
	}

	var prevAssignee *uint
	if err := tx.Raw(`SELECT assignee_user_id FROM process_tasks
		WHERE process_instance_id = ? AND node_code = ? AND status = 'completed'
		ORDER BY id DESC LIMIT 1`, instanceID, prevNode).Scan(&prevAssignee).Error; err != nil {
		return fmt.Errorf("sod check: %w", err)
	}
	if prevAssignee != nil && *prevAssignee == userID {
		return ErrSoDViolation
	}
	return nil
}

func (e *Engine) createTask(instanceID uint, nodeCode, nodeName string, deptID uint, assigneeUserID ...uint) error {
	return e.createTaskWithTx(e.db, instanceID, nodeCode, nodeName, deptID)
}

func (e *Engine) createTaskWithTx(tx *gorm.DB, instanceID uint, nodeCode, nodeName string, deptID uint, assigneeUserID ...uint) error {
	// Get the responsible department from node definition
	def, ok := e.nodeMap[nodeCode]
	if !ok {
		return ErrUnknownNode
	}

	// Resolve department code to actual dept_id
	var resolvedDeptID uint
	if err := tx.Raw(`SELECT id FROM depts WHERE code = ?`, def.DeptCode).Scan(&resolvedDeptID).Error; err != nil {
		return fmt.Errorf("resolve dept: %w", err)
	}
	if resolvedDeptID == 0 {
		return fmt.Errorf("department %q not found in depts table (node=%s)", def.DeptCode, nodeCode)
	}

	var userIDPtr *uint
	if len(assigneeUserID) > 0 && assigneeUserID[0] != 0 {
		uid := assigneeUserID[0]
		userIDPtr = &uid
	}
	return tx.Create(&model.ProcessTask{
		ProcessInstanceID: instanceID,
		NodeCode:          nodeCode,
		NodeName:          nodeName,
		AssigneeDeptID:    &resolvedDeptID,
		AssigneeUserID:    userIDPtr,
		Status:            TaskStatusPending,
		Version:           0,
	}).Error
}

func (e *Engine) queryTasks(query string, args ...interface{}) ([]map[string]interface{}, error) {
	rows, err := e.db.Raw(query, args...).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []map[string]interface{}
	cols, _ := rows.Columns()
	for rows.Next() {
		vals := make([]interface{}, len(cols))
		for i := range vals {
			vals[i] = new(interface{})
		}
		rows.Scan(vals...)
		row := make(map[string]interface{})
		for i, col := range cols {
			row[col] = *(vals[i].(*interface{}))
		}
		results = append(results, row)
	}
	return results, nil
}
