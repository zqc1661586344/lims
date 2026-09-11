package handler

import (
	"fmt"
	"lims-backend/internal/middleware"
	"lims-backend/internal/model"
	"lims-backend/internal/service"
	"lims-backend/internal/utils"
	"lims-backend/internal/workflow"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func NewDataAuditHandler(logger *zap.Logger, db *gorm.DB) *DataAuditHandler {
	return &DataAuditHandler{
		GenericHandler: NewGenericHandler[model.DataAudit](logger, db),
		svc:            service.NewBusinessService(logger, db),
	}
}

func (h *DataAuditHandler) List(c *gin.Context) {
	h.GenericHandler.List(c, nil, func(db *gorm.DB, c *gin.Context) *gorm.DB {
		if id := c.Query("task_order_id"); id != "" {
			db = db.Where("task_order_id = ?", id)
		}
		return db
	})
}

func (h *DataAuditHandler) Get(c *gin.Context) {
	h.GenericHandler.Get(c)
}

func (h *DataAuditHandler) Create(c *gin.Context) {
	var req struct {
		TaskOrderID  uint   `json:"task_order_id" binding:"required"`
		AuditResult  string `json:"audit_result"`
		AuditComment string `json:"audit_comment"`
		IssueList    string `json:"issue_list"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数错误: %v", err))
		return
	}
	item := model.DataAudit{
		TaskOrderID:  req.TaskOrderID,
		AuditResult:  req.AuditResult,
		AuditComment: req.AuditComment,
		IssueList:    req.IssueList,
	}
	if err := h.GetDB(c).Create(&item).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("创建数据审核失败: %v", err))
		return
	}
	utils.Created(c, item)
}

func (h *DataAuditHandler) Update(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	var item model.DataAudit
	if err := h.GetDB(c).First(&item, id).Error; err != nil {
		utils.NotFound(c, "数据审核记录不存在")
		return
	}
	if err := h.svc.CheckNodeNotAdvanced(h.GetDB(c), "task_order", item.TaskOrderID, workflow.NodeDataAudit); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	var req struct {
		AuditResult  string `json:"audit_result"`
		AuditComment string `json:"audit_comment"`
		IssueList    string `json:"issue_list"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数错误: %v", err))
		return
	}
	updates := map[string]interface{}{}
	if req.AuditResult != "" {
		updates["audit_result"] = req.AuditResult
	}
	if req.AuditComment != "" {
		updates["audit_comment"] = req.AuditComment
	}
	if req.IssueList != "" {
		updates["issue_list"] = req.IssueList
	}
	if err := h.GetDB(c).Model(&item).Updates(updates).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("更新数据审核失败: %v", err))
		return
	}
	h.GetDB(c).First(&item, id)
	utils.Success(c, item)
}

func (h *DataAuditHandler) Delete(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	var item model.DataAudit
	if err := h.GetDB(c).First(&item, id).Error; err != nil {
		utils.NotFound(c, "数据审核记录不存在")
		return
	}
	if err := h.svc.CheckNodeNotAdvanced(h.GetDB(c), "task_order", item.TaskOrderID, workflow.NodeDataAudit); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	if err := h.svc.CheckInstanceRunning(h.GetDB(c), "task_order", item.TaskOrderID); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	if err := h.GetDB(c).Delete(&item).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("删除数据审核失败: %v", err))
		return
	}
	utils.Success(c, nil)
}

func (h *DataAuditHandler) Approve(c *gin.Context) {
	var req struct {
		TaskID       uint   `json:"task_id" binding:"required"`
		AuditResult  string `json:"audit_result"`
		AuditComment string `json:"audit_comment"`
		IssueList    string `json:"issue_list"`
		Comment      string `json:"comment"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数错误: %v", err))
		return
	}
	audit := model.DataAudit{
		TaskOrderID:  req.TaskID,
		AuditResult:  "通过",
		AuditComment: req.AuditComment,
		IssueList:    req.IssueList,
	}
	if req.AuditResult != "" {
		audit.AuditResult = req.AuditResult
	}
	h.GetDB(c).Where("task_order_id = ?", req.TaskID).Assign(audit).FirstOrCreate(&audit)

	userID := middleware.GetUserID(c)
	if err := h.svc.ApproveTaskByOrder(req.TaskID, userID, middleware.GetDeptIDVal(c), req.Comment); err != nil {
		utils.InternalError(c, fmt.Sprintf("审批失败: %v", err))
		return
	}
	utils.Success(c, gin.H{"message": "数据审核通过"})
}

func (h *DataAuditHandler) Reject(c *gin.Context) {
	var req struct {
		TaskID       uint   `json:"task_id" binding:"required"`
		AuditComment string `json:"audit_comment" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数错误: %v", err))
		return
	}
	audit := model.DataAudit{
		TaskOrderID:  req.TaskID,
		AuditResult:  "驳回",
		AuditComment: req.AuditComment,
	}
	h.GetDB(c).Where("task_order_id = ?", req.TaskID).Assign(audit).FirstOrCreate(&audit)

	userID := middleware.GetUserID(c)
	if err := h.svc.RejectTaskByOrder(req.TaskID, userID, middleware.GetDeptIDVal(c), req.AuditComment); err != nil {
		utils.InternalError(c, fmt.Sprintf("驳回失败: %v", err))
		return
	}
	utils.Success(c, gin.H{"message": "数据审核已驳回"})
}

// ============================================================
// ReportPrepareHandler — 报告编制（节点11）
// ============================================================

type ReportPrepareHandler struct {
	*GenericHandler[model.ReportPrepare]
	svc *service.BusinessService
}
