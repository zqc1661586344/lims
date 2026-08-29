package handler

import (
	"fmt"
	"lims-backend/internal/middleware"
	"lims-backend/internal/service"
	"lims-backend/internal/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// WorkflowHandler handles workflow-related HTTP requests.
type WorkflowHandler struct {
	svc *service.WorkflowService
}

// NewWorkflowHandler creates a new workflow handler.
func NewWorkflowHandler(db *gorm.DB) *WorkflowHandler {
	return &WorkflowHandler{
		svc: service.NewWorkflowService(db),
	}
}

// StartInstance creates a new process instance.
func (h *WorkflowHandler) StartInstance(c *gin.Context) {
	var req struct {
		BusinessType string `json:"business_type" binding:"required"`
		BusinessID   uint   `json:"business_id" binding:"required"`
		Title        string `json:"title" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数错误: %v", err))
		return
	}

	userID := middleware.GetUserID(c)
	id, err := h.svc.StartInstance(req.BusinessType, req.BusinessID, req.Title, userID)
	if err != nil {
		utils.InternalError(c, fmt.Sprintf("启动流程失败: %v", err))
		return
	}
	utils.Created(c, gin.H{"id": id})
}

// GetInstance returns a process instance by ID.
func (h *WorkflowHandler) GetInstance(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	instance, err := h.svc.GetInstance(id)
	if err != nil {
		utils.InternalError(c, "查询流程实例失败")
		return
	}
	if instance == nil {
		utils.NotFound(c, "流程实例不存在")
		return
	}
	utils.Success(c, instance)
}

// GetPendingTasks returns pending tasks for the current user or their department.
// Admin (is_admin=true) sees all departments' pending tasks (cross-department view).
func (h *WorkflowHandler) GetPendingTasks(c *gin.Context) {
	userID := middleware.GetUserID(c)
	deptID := middleware.GetDeptID(c)
	isAdmin := middleware.IsAdmin(c)
	tasks, err := h.svc.GetPendingTasks(deptID, userID, isAdmin)
	if err != nil {
		utils.InternalError(c, "查询待办任务失败")
		return
	}
	utils.Success(c, tasks)
}

// GetPendingTasksByUser returns pending tasks only for the current user.
// Admin still gets cross-department tasks here so a single admin can drive the full flow.
func (h *WorkflowHandler) GetPendingTasksByUser(c *gin.Context) {
	userID := middleware.GetUserID(c)
	isAdmin := middleware.IsAdmin(c)
	tasks, err := h.svc.GetPendingTasks(nil, userID, isAdmin) // deptID=nil → 查本人（admin 时查全部）
	if err != nil {
		utils.InternalError(c, "查询我的待办失败")
		return
	}
	utils.Success(c, tasks)
}

// GetProcessHistory returns the task history for a process instance.
func (h *WorkflowHandler) GetProcessHistory(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	history, err := h.svc.GetProcessHistory(id)
	if err != nil {
		utils.InternalError(c, "查询流程历史失败")
		return
	}
	utils.Success(c, history)
}

// ApproveTask approves a pending task.
func (h *WorkflowHandler) ApproveTask(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的任务ID")
		return
	}

	var req struct {
		Comment string `json:"comment"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Comment = ""
	}

	userID := middleware.GetUserID(c)
	if err := h.svc.ApproveTask(id, userID, req.Comment); err != nil {
		utils.InternalError(c, fmt.Sprintf("审批失败: %v", err))
		return
	}
	utils.Success(c, nil)
}

// RejectTask rejects a pending task.
func (h *WorkflowHandler) RejectTask(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的任务ID")
		return
	}

	var req struct {
		Comment string `json:"comment"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Comment = ""
	}
	if req.Comment == "" {
		utils.BadRequest(c, "驳回时必须填写意见")
		return
	}

	userID := middleware.GetUserID(c)
	if err := h.svc.RejectTask(id, userID, req.Comment); err != nil {
		utils.InternalError(c, fmt.Sprintf("驳回失败: %v", err))
		return
	}
	utils.Success(c, nil)
}

// GetNodeDefinitions returns all workflow node definitions.
func (h *WorkflowHandler) GetNodeDefinitions(c *gin.Context) {
	defs := h.svc.GetNodeDefinitions()
	utils.Success(c, defs)
}

// parseUint parses a string to uint.
func parseUint(s string) (uint, error) {
	var v uint
	_, err := fmt.Sscanf(s, "%d", &v)
	return v, err
}