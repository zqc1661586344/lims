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

func NewReportReviewHandler(logger *zap.Logger, db *gorm.DB) *ReportReviewHandler {
	return &ReportReviewHandler{
		GenericHandler: NewGenericHandler[model.ReportReview](logger, db),
		svc:            service.NewBusinessService(logger, db),
	}
}

func (h *ReportReviewHandler) List(c *gin.Context) {
	h.GenericHandler.List(c, nil, func(db *gorm.DB, c *gin.Context) *gorm.DB {
		if id := c.Query("task_order_id"); id != "" {
			db = db.Where("task_order_id = ?", id)
		}
		return db
	})
}

func (h *ReportReviewHandler) Get(c *gin.Context) {
	h.GenericHandler.Get(c)
}

func (h *ReportReviewHandler) Create(c *gin.Context) {
	var req struct {
		TaskOrderID   uint   `json:"task_order_id" binding:"required"`
		ReviewResult  string `json:"review_result"`
		ReviewComment string `json:"review_comment"`
		ReviewedItems string `json:"reviewed_items"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数错误: %v", err))
		return
	}
	item := model.ReportReview{
		TaskOrderID:   req.TaskOrderID,
		ReviewResult:  req.ReviewResult,
		ReviewComment: req.ReviewComment,
		ReviewedItems: model.JSONB(req.ReviewedItems),
	}
	if err := h.GetDB(c).Create(&item).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("创建报告复核失败: %v", err))
		return
	}
	utils.Created(c, item)
}

func (h *ReportReviewHandler) Update(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	var item model.ReportReview
	if err := h.GetDB(c).First(&item, id).Error; err != nil {
		utils.NotFound(c, "报告复核记录不存在")
		return
	}
	if err := h.svc.CheckNodeNotAdvanced(h.GetDB(c), "task_order", item.TaskOrderID, workflow.NodeReportReview); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	var req struct {
		ReviewResult  string `json:"review_result"`
		ReviewComment string `json:"review_comment"`
		ReviewedItems string `json:"reviewed_items"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数错误: %v", err))
		return
	}
	updates := map[string]interface{}{}
	if req.ReviewResult != "" {
		updates["review_result"] = req.ReviewResult
	}
	if req.ReviewComment != "" {
		updates["review_comment"] = req.ReviewComment
	}
	if req.ReviewedItems != "" {
		updates["reviewed_items"] = req.ReviewedItems
	}
	if err := h.GetDB(c).Model(&item).Updates(updates).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("更新报告复核失败: %v", err))
		return
	}
	h.GetDB(c).First(&item, id)
	utils.Success(c, item)
}

func (h *ReportReviewHandler) Delete(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	var item model.ReportReview
	if err := h.GetDB(c).First(&item, id).Error; err != nil {
		utils.NotFound(c, "报告复核记录不存在")
		return
	}
	if err := h.svc.CheckNodeNotAdvanced(h.GetDB(c), "task_order", item.TaskOrderID, workflow.NodeReportReview); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	if err := h.svc.CheckInstanceRunning(h.GetDB(c), "task_order", item.TaskOrderID); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	if err := h.GetDB(c).Delete(&item).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("删除报告复核失败: %v", err))
		return
	}
	utils.Success(c, nil)
}

func (h *ReportReviewHandler) Approve(c *gin.Context) {
	var req struct {
		TaskID        uint   `json:"task_id" binding:"required"`
		ReviewResult  string `json:"review_result"`
		ReviewComment string `json:"review_comment"`
		ReviewedItems string `json:"reviewed_items"`
		Comment       string `json:"comment"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数错误: %v", err))
		return
	}
	effectiveResult := "通过"
	if req.ReviewResult != "" {
		effectiveResult = req.ReviewResult
	}

	userID := middleware.GetUserID(c)
	if err := h.svc.ApproveWithBusiness(req.TaskID, userID, middleware.GetDeptIDVal(c), req.Comment, func(tx *gorm.DB) error {
		rec := model.ReportReview{
			TaskOrderID:   req.TaskID,
			ReviewResult:  effectiveResult,
			ReviewComment: req.ReviewComment,
			ReviewedItems: model.JSONB(req.ReviewedItems),
		}
		return tx.Where("task_order_id = ?", req.TaskID).Assign(rec).FirstOrCreate(&rec).Error
	}); err != nil {
		HandleWorkflowError(h.Logger, c, err, "审批失败")
		return
	}
	utils.Success(c, gin.H{"message": "报告复核通过"})
}

func (h *ReportReviewHandler) Reject(c *gin.Context) {
	var req struct {
		TaskID        uint   `json:"task_id" binding:"required"`
		ReviewComment string `json:"review_comment" binding:"required"`
		RejectTarget  string `json:"reject_target"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数错误: %v", err))
		return
	}

	userID := middleware.GetUserID(c)
	var opts []string
	if req.RejectTarget != "" {
		opts = append(opts, req.RejectTarget)
	}
	if err := h.svc.RejectWithBusiness(req.TaskID, userID, middleware.GetDeptIDVal(c), req.ReviewComment, func(tx *gorm.DB) error {
		rec := model.ReportReview{
			TaskOrderID:   req.TaskID,
			ReviewResult:  "驳回",
			ReviewComment: req.ReviewComment,
		}
		return tx.Where("task_order_id = ?", req.TaskID).Assign(rec).FirstOrCreate(&rec).Error
	}, opts...); err != nil {
		HandleWorkflowError(h.Logger, c, err, "驳回失败")
		return
	}
	utils.Success(c, gin.H{"message": "报告复核已驳回"})
}

// ============================================================
// ReportAuditHandler — 报告审核（节点13）
// ============================================================

type ReportAuditHandler struct {
	*GenericHandler[model.ReportAudit]
	svc *service.BusinessService
}
