package service

import (
	"lims-backend/internal/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ApplyTaskOrderScope restricts a GORM query to rows whose associated task_order
// the current user is allowed to see.
//
// Access rules (admin bypasses all):
//  1. task_orders.created_by = current user
//  2. OR current user's department has an assigned task (any node) in the
//     process instance linked to that task_order
//
// Models are expected to expose their FK to task_orders via the column name
// "task_order_id" (which every business node table uses).
func ApplyTaskOrderScope(db *gorm.DB, c *gin.Context) *gorm.DB {
	if middleware.IsAdmin(c) {
		return db
	}

	userID := middleware.GetUserID(c)
	userDeptID := middleware.GetDeptIDVal(c)

	return db.Where(
		`task_order_id IN (
			SELECT to2.id FROM task_orders to2
			WHERE to2.created_by = ?
			   OR to2.id IN (
					SELECT pi2.business_id
					FROM process_instances pi2
					JOIN process_tasks pt2 ON pt2.process_instance_id = pi2.id
					WHERE pi2.business_type = 'task_order'
					  AND pt2.assignee_dept_id = ?
			   )
		)`,
		userID, userDeptID,
	)
}

func ApplyTaskOrderSelfScope(db *gorm.DB, c *gin.Context) *gorm.DB {
	if middleware.IsAdmin(c) {
		return db
	}

	userID := middleware.GetUserID(c)
	userDeptID := middleware.GetDeptIDVal(c)

	return db.Where(
		`id IN (
			SELECT to2.id FROM task_orders to2
			WHERE to2.created_by = ?
			   OR to2.id IN (
					SELECT pi2.business_id
					FROM process_instances pi2
					JOIN process_tasks pt2 ON pt2.process_instance_id = pi2.id
					WHERE pi2.business_type = 'task_order'
					  AND pt2.assignee_dept_id = ?
			   )
		)`,
		userID, userDeptID,
	)
}

// CanAccessTaskOrder checks whether the current user may view the given task_order.
// Used as a cheap existence check before returning a 403 on Get requests.
func CanAccessTaskOrder(db *gorm.DB, c *gin.Context, taskOrderID uint) bool {
	if middleware.IsAdmin(c) {
		return true
	}

	userID := middleware.GetUserID(c)
	userDeptID := middleware.GetDeptIDVal(c)

	var cnt int64
	db.Raw(`
		SELECT COUNT(*) FROM task_orders to2
		WHERE to2.id = ? AND (
			to2.created_by = ? OR to2.id IN (
				SELECT pi2.business_id
				FROM process_instances pi2
				JOIN process_tasks pt2 ON pt2.process_instance_id = pi2.id
				WHERE pi2.business_type = 'task_order'
				  AND pt2.assignee_dept_id = ?
			)
		)`, taskOrderID, userID, userDeptID).Scan(&cnt)

	return cnt > 0
}
