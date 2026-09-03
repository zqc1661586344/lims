package workflow

import (
	"errors"
	"fmt"
	"lims-backend/internal/model"
	"sync"

	"gorm.io/gorm"
)

// Engine is the core workflow state machine engine.
type Engine struct {
	db       *gorm.DB
	nodeDefs []NodeDefinition
	nodeMap  map[string]NodeDefinition
}

var (
	defaultEngineOnce sync.Once
	defaultEngine     *Engine
)

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
func (e *Engine) ApproveTask(taskID uint, userID uint, comment string) error {
	return e.db.Transaction(func(tx *gorm.DB) error {
		return e.ApproveTaskWithTx(tx, taskID, userID, comment)
	})
}

// ApproveTaskWithTx is the transaction-aware core of ApproveTask.
// The caller is responsible for starting and committing the transaction.
func (e *Engine) ApproveTaskWithTx(tx *gorm.DB, taskID uint, userID uint, comment string) error {
	task, err := e.getTaskForUpdate(tx, taskID)
	if err != nil {
		return err
	}
	if task.Status != TaskStatusPending {
		return ErrTaskAlreadyCompleted
	}

	// Fetch the instance
	var instance struct {
		ID          uint
		CurrentNode string
		Status      string
	}
	if err := tx.Raw(`SELECT id, current_node, status FROM process_instances
		WHERE id=? FOR UPDATE`, task.ProcessInstanceID).Scan(&instance).Error; err != nil {
		return fmt.Errorf("get instance: %w", err)
	}
	if instance.Status != InstanceStatusRunning {
		return ErrInstanceNotRunning
	}

	nodeDef, ok := e.nodeMap[task.NodeCode]
	if !ok {
		return ErrUnknownNode
	}

	// Mark current task as completed
	if err := tx.Exec(`UPDATE process_tasks SET status=?, comment=?, updated_at=NOW()
		WHERE id=?`,
		TaskStatusCompleted, comment, taskID).Error; err != nil {
		return fmt.Errorf("update task: %w", err)
	}

	nextNode := nodeDef.NextNode
	if nextNode == "" {
		// Terminal node — complete the instance
		if err := tx.Exec(`UPDATE process_instances SET status=?, current_node=?, updated_at=NOW()
			WHERE id=?`, InstanceStatusCompleted, task.NodeCode, instance.ID).Error; err != nil {
			return fmt.Errorf("complete instance: %w", err)
		}
	} else {
		// Create the next task
		nextDef, ok := e.nodeMap[nextNode]
		if !ok {
			return ErrUnknownNode
		}
		if err := e.createTaskWithTx(tx, instance.ID, nextNode, nextDef.Name, 0); err != nil {
			return err
		}
		// Advance the instance current node
		if err := tx.Exec(`UPDATE process_instances SET current_node=?, updated_at=NOW()
			WHERE id=?`, nextNode, instance.ID).Error; err != nil {
			return fmt.Errorf("advance instance: %w", err)
		}
	}
	return nil
}

// RejectTask rejects a pending task, creating a new pending task for the
// rejection target node (typically the previous node in the workflow).
func (e *Engine) RejectTask(taskID uint, userID uint, comment string) error {
	return e.db.Transaction(func(tx *gorm.DB) error {
		return e.RejectTaskWithTx(tx, taskID, userID, comment)
	})
}

// RejectTaskWithTx is the transaction-aware core of RejectTask.
// The caller is responsible for starting and committing the transaction.
func (e *Engine) RejectTaskWithTx(tx *gorm.DB, taskID uint, userID uint, comment string) error {
	task, err := e.getTaskForUpdate(tx, taskID)
	if err != nil {
		return err
	}
	if task.Status != TaskStatusPending {
		return ErrTaskAlreadyCompleted
	}

	nodeDef, ok := e.nodeMap[task.NodeCode]
	if !ok {
		return ErrUnknownNode
	}
	if !nodeDef.CanReject {
		return ErrCannotRejectFinalNode
	}

	rejectTarget := nodeDef.RejectTarget
	if rejectTarget == "" {
		return ErrCannotRejectFromFirstNode
	}

	// Fetch the instance
	var instanceID uint
	if err := tx.Raw(`SELECT id FROM process_instances WHERE id=? FOR UPDATE`,
		task.ProcessInstanceID).Scan(&instanceID).Error; err != nil {
		return fmt.Errorf("get instance: %w", err)
	}

	// Mark current task as rejected
	if err := tx.Exec(`UPDATE process_tasks SET status=?, comment=?, updated_at=NOW()
		WHERE id=?`,
		TaskStatusRejected, comment, taskID).Error; err != nil {
		return fmt.Errorf("update task rejected: %w", err)
	}

	// Create a new pending task for the reject target node
	targetDef, ok := e.nodeMap[rejectTarget]
	if !ok {
		return ErrUnknownNode
	}
	if err := e.createTaskWithTx(tx, instanceID, rejectTarget, targetDef.Name, 0); err != nil {
		return err
	}

	// Update instance current node back to reject target
	if err := tx.Exec(`UPDATE process_instances SET current_node=?, updated_at=NOW()
		WHERE id=?`, rejectTarget, instanceID).Error; err != nil {
		return fmt.Errorf("revert instance node: %w", err)
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

// --- internal helpers ---

type taskRow struct {
	ID                uint
	ProcessInstanceID uint
	NodeCode          string
	Status            string
	Version           int
}

func (e *Engine) getTaskForUpdate(tx *gorm.DB, taskID uint) (*taskRow, error) {
	var row taskRow
	if err := tx.Raw(`SELECT id, process_instance_id, node_code, status, version
		FROM process_tasks WHERE id=? FOR UPDATE`, taskID).Scan(&row).Error; err != nil {
		return nil, fmt.Errorf("get task: %w", err)
	}
	if row.ID == 0 {
		return nil, fmt.Errorf("task not found: %d", taskID)
	}
	return &row, nil
}

func (e *Engine) createTask(instanceID uint, nodeCode, nodeName string, deptID uint) error {
	return e.createTaskWithTx(e.db, instanceID, nodeCode, nodeName, deptID)
}

func (e *Engine) createTaskWithTx(tx *gorm.DB, instanceID uint, nodeCode, nodeName string, deptID uint) error {
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

	return tx.Create(&model.ProcessTask{
		ProcessInstanceID: instanceID,
		NodeCode:          nodeCode,
		NodeName:          nodeName,
		AssigneeDeptID:    &resolvedDeptID,
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
