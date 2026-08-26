package workflow

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
)

// Engine is the core workflow state machine engine.
type Engine struct {
	db       *gorm.DB
	nodeDefs []NodeDefinition
	nodeMap  map[string]NodeDefinition
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

// StartInstance creates a new process instance and its first task.
// Returns the instance ID.
func (e *Engine) StartInstance(businessType string, businessID uint, title string, createdBy uint) (uint, error) {
	if len(e.nodeDefs) == 0 {
		return 0, errors.New("no workflow definition loaded")
	}
	firstNode := e.nodeDefs[0]

	instance := struct {
		BusinessType string
		BusinessID   uint
		Title        string
		CurrentNode  string
		Status       string
		CreatedBy    uint
	}{
		BusinessType: businessType,
		BusinessID:   businessID,
		Title:        title,
		CurrentNode:  firstNode.Code,
		Status:       InstanceStatusRunning,
		CreatedBy:    createdBy,
	}

	if err := e.db.Exec(`INSERT INTO process_instances
		(business_type, business_id, title, current_node, status, created_by, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, NOW(), NOW())`,
		instance.BusinessType, instance.BusinessID, instance.Title,
		instance.CurrentNode, instance.Status, instance.CreatedBy).
		Error; err != nil {
		return 0, fmt.Errorf("create process instance: %w", err)
	}

	// Retrieve the ID of the newly created instance
	var instanceID uint
	if err := e.db.Raw(`SELECT id FROM process_instances
		WHERE business_type=? AND business_id=? ORDER BY created_at DESC LIMIT 1`,
		businessType, businessID).Scan(&instanceID).Error; err != nil {
		return 0, fmt.Errorf("get instance id: %w", err)
	}

	// Create the first task
	if err := e.createTask(instanceID, firstNode.Code, firstNode.Name, 0); err != nil {
		return 0, fmt.Errorf("create first task: %w", err)
	}

	return instanceID, nil
}

// ApproveTask approves a pending task, creating the next node's task.
// If the approved node is the terminal node, the instance is marked completed.
func (e *Engine) ApproveTask(taskID uint, userID uint, comment string) error {
	return e.db.Transaction(func(tx *gorm.DB) error {
		// Fetch the task with optimistic lock check
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
			Version     int
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
		if err := tx.Exec(`UPDATE process_tasks SET status=?, comment=?, updated_at=NOW(), version=version+1
			WHERE id=? AND version=?`,
			TaskStatusCompleted, comment, taskID, task.Version).Error; err != nil {
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
	})
}

// RejectTask rejects a pending task, creating a new pending task for the
// rejection target node (typically the previous node in the workflow).
func (e *Engine) RejectTask(taskID uint, userID uint, comment string) error {
	return e.db.Transaction(func(tx *gorm.DB) error {
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
		if err := tx.Exec(`UPDATE process_tasks SET status=?, comment=?, updated_at=NOW(), version=version+1
			WHERE id=? AND version=?`,
			TaskStatusRejected, comment, taskID, task.Version).Error; err != nil {
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
	})
}

// GetPendingTasksByUser returns all pending tasks assigned to a specific user.
func (e *Engine) GetPendingTasksByUser(userID uint) ([]map[string]interface{}, error) {
	return e.queryTasks(`
		SELECT pt.id, pt.node_code, pt.node_name, pt.created_at,
			pi.title, pi.business_type, pi.business_id, pi.id as process_instance_id
		FROM process_tasks pt
		JOIN process_instances pi ON pi.id = pt.process_instance_id
		WHERE pt.status = 'pending' AND pt.assignee_user_id = ?
		ORDER BY pt.created_at DESC`, userID)
}

// GetPendingTasksByDept returns all pending tasks assigned to a department.
func (e *Engine) GetPendingTasksByDept(deptID uint) ([]map[string]interface{}, error) {
	return e.queryTasks(`
		SELECT pt.id, pt.node_code, pt.node_name, pt.created_at,
			pi.title, pi.business_type, pi.business_id, pi.id as process_instance_id
		FROM process_tasks pt
		JOIN process_instances pi ON pi.id = pt.process_instance_id
		WHERE pt.status = 'pending' AND pt.assignee_dept_id = ?
		ORDER BY pt.created_at DESC`, deptID)
}

// GetProcessHistory returns the full task history for a process instance.
func (e *Engine) GetProcessHistory(instanceID uint) ([]map[string]interface{}, error) {
	return e.queryTasks(`
		SELECT pt.id, pt.node_code, pt.node_name, pt.status, pt.comment,
			pt.assignee_user_id, pt.created_at, pt.updated_at
		FROM process_tasks pt
		WHERE pt.process_instance_id = ?
		ORDER BY pt.id ASC`, instanceID)
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
	ID               uint
	ProcessInstanceID uint
	NodeCode         string
	Status           string
	Version          int
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

	// Map department code to actual dept_id
	// We store the dept_code as a reference; the actual dept_id is resolved at creation
	_ = deptID // deptID is kept for future refinement

	return tx.Exec(`INSERT INTO process_tasks
		(process_instance_id, node_code, node_name, assignee_dept_id,
		 status, created_at, updated_at, version)
		VALUES (?, ?, ?,
			(SELECT id FROM depts WHERE code = ?),
			'pending', NOW(), NOW(), 0)`,
		instanceID, nodeCode, nodeName, def.DeptCode).Error
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