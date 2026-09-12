package handler

import (
	"errors"
	"fmt"
	"lims-backend/internal/middleware"
	"lims-backend/internal/service"
	"lims-backend/internal/utils"
	"lims-backend/internal/workflow"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// WorkflowHandler handles workflow-related HTTP requests.
type WorkflowHandler struct {
	logger *zap.Logger
	svc    *service.WorkflowService
}

// NewWorkflowHandler creates a new workflow handler.
func NewWorkflowHandler(logger *zap.Logger, db *gorm.DB) *WorkflowHandler {
	return &WorkflowHandler{
		logger: logger,
		svc:    service.NewWorkflowService(logger, db),
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
	if err := h.svc.ApproveTask(id, userID, middleware.GetDeptIDVal(c), req.Comment); err != nil {
		HandleWorkflowError(c, err, "审批失败")
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
		Comment      string `json:"comment"`
		RejectTarget string `json:"reject_target"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Comment = ""
	}
	if req.Comment == "" {
		utils.BadRequest(c, "驳回时必须填写意见")
		return
	}

	userID := middleware.GetUserID(c)
	var opts []string
	if req.RejectTarget != "" {
		opts = append(opts, req.RejectTarget)
	}
	if err := h.svc.RejectTask(id, userID, middleware.GetDeptIDVal(c), req.Comment, opts...); err != nil {
		HandleWorkflowError(c, err, "驳回失败")
		return
	}
	utils.Success(c, nil)
}

// HandleWorkflowError maps workflow engine sentinel errors to the correct
// HTTP status code so the frontend can distinguish authz (403), conflict
// (409), bad request (400), and server errors (500).
func HandleWorkflowError(c *gin.Context, err error, action string) {
	switch {
	case errors.Is(err, workflow.ErrDeptNotMatch):
		utils.Forbidden(c, fmt.Sprintf("%s: 当前用户部门无权操作该任务", action))
	case errors.Is(err, workflow.ErrAssigneeNotMatch):
		utils.Forbidden(c, fmt.Sprintf("%s: 该任务已分配给其他人员", action))
	case errors.Is(err, workflow.ErrSoDViolation):
		utils.Forbidden(c, fmt.Sprintf("%s: 职责分离违规——审核/复核人不能是前序节点的操作人", action))
	case errors.Is(err, workflow.ErrOptimisticLock):
		utils.Error(c, http.StatusConflict, fmt.Sprintf("%s: 数据已被他人修改，请刷新后重试", action))
	case errors.Is(err, workflow.ErrTaskAlreadyCompleted):
		utils.BadRequest(c, fmt.Sprintf("%s: 任务已完成或已驳回", action))
	case errors.Is(err, workflow.ErrInstanceNotRunning):
		utils.BadRequest(c, fmt.Sprintf("%s: 流程实例已终止", action))
	case errors.Is(err, workflow.ErrForbidden):
		utils.Forbidden(c, fmt.Sprintf("%s: 无权访问此流程", action))
	default:
		utils.InternalError(c, fmt.Sprintf("%s: %v", action, err))
	}
}

// GetNodeDefinitions returns all workflow node definitions.
func (h *WorkflowHandler) GetNodeDefinitions(c *gin.Context) {
	defs := h.svc.GetNodeDefinitions()
	utils.Success(c, defs)
}

// GetProgress returns workflow progress for a business object.
// Path: /workflow/progress/:businessType/:businessId
func (h *WorkflowHandler) GetProgress(c *gin.Context) {
	businessType := c.Param("businessType")
	businessID, err := parseUint(c.Param("businessId"))
	if err != nil {
		utils.BadRequest(c, "无效的业务ID")
		return
	}
	userID := middleware.GetUserID(c)
	userDeptID := middleware.GetDeptIDVal(c)
	isAdmin := middleware.IsAdmin(c)
	progress, err := h.svc.GetProgressByBusiness(businessType, businessID, userID, userDeptID, isAdmin)
	if err != nil {
		if errors.Is(err, workflow.ErrForbidden) {
			utils.Forbidden(c, "无权查看该流程进度")
			return
		}
		utils.InternalError(c, "查询流程进度失败")
		return
	}
	if progress == nil {
		utils.Success(c, nil)
		return
	}
	utils.Success(c, progress)
}

// parseUint parses a string to uint.
func parseUint(s string) (uint, error) {
	var v uint
	_, err := fmt.Sscanf(s, "%d", &v)
	return v, err
}
