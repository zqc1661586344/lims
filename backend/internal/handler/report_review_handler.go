package handler

import (
	"fmt"
	"lims-backend/internal/middleware"
	"lims-backend/internal/model"
	"lims-backend/internal/service"
	"lims-backend/internal/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func NewReportReviewHandler(logger *zap.Logger, db *gorm.DB) *ReportReviewHandler {
	return &ReportReviewHandler{
		svc: service.NewBusinessService(logger, db),
		db:  db,
	}
}

func (h *ReportReviewHandler) getDB(c *gin.Context) *gorm.DB {
	if db := middleware.GetDB(c); db != nil {
		return db
	}
	return h.db
}

func (h *ReportReviewHandler) List(c *gin.Context) {
	var items []model.ReportReview
	query := h.getDB(c).Order("id DESC")
	if taskOrderID := c.Query("task_order_id"); taskOrderID != "" {
		query = query.Where("task_order_id = ?", taskOrderID)
	}
	if err := query.Find(&items).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("查询报告复核失败: %v", err))
		return
	}
	utils.Success(c, items)
}

func (h *ReportReviewHandler) Get(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	var item model.ReportReview
	if err := h.getDB(c).First(&item, id).Error; err != nil {
		utils.NotFound(c, "报告复核记录不存在")
		return
	}
	utils.Success(c, item)
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
		ReviewedItems: req.ReviewedItems,
	}
	if err := h.getDB(c).Create(&item).Error; err != nil {
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
	if err := h.getDB(c).First(&item, id).Error; err != nil {
		utils.NotFound(c, "报告复核记录不存在")
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
	if err := h.getDB(c).Model(&item).Updates(updates).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("更新报告复核失败: %v", err))
		return
	}
	h.getDB(c).First(&item, id)
	utils.Success(c, item)
}

func (h *ReportReviewHandler) Delete(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	if err := h.getDB(c).Delete(&model.ReportReview{}, id).Error; err != nil {
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
	rec := model.ReportReview{
		TaskOrderID:   req.TaskID,
		ReviewResult:  "通过",
		ReviewComment: req.ReviewComment,
		ReviewedItems: req.ReviewedItems,
	}
	if req.ReviewResult != "" {
		rec.ReviewResult = req.ReviewResult
	}
	h.getDB(c).Where("task_order_id = ?", req.TaskID).Assign(rec).FirstOrCreate(&rec)

	userID := middleware.GetUserID(c)
	if err := h.svc.ApproveTaskByOrder(req.TaskID, userID, req.Comment); err != nil {
		utils.InternalError(c, fmt.Sprintf("审批失败: %v", err))
		return
	}
	utils.Success(c, gin.H{"message": "报告复核通过"})
}

func (h *ReportReviewHandler) Reject(c *gin.Context) {
	var req struct {
		TaskID        uint   `json:"task_id" binding:"required"`
		ReviewComment string `json:"review_comment" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数错误: %v", err))
		return
	}
	rec := model.ReportReview{
		TaskOrderID:   req.TaskID,
		ReviewResult:  "驳回",
		ReviewComment: req.ReviewComment,
	}
	h.getDB(c).Where("task_order_id = ?", req.TaskID).Assign(rec).FirstOrCreate(&rec)

	userID := middleware.GetUserID(c)
	if err := h.svc.RejectTaskByOrder(req.TaskID, userID, req.ReviewComment); err != nil {
		utils.InternalError(c, fmt.Sprintf("驳回失败: %v", err))
		return
	}
	utils.Success(c, gin.H{"message": "报告复核已驳回"})
}

// ============================================================
// ReportAuditHandler — 报告审核（节点13）
// ============================================================

type ReportAuditHandler struct {
	svc *service.BusinessService
	db  *gorm.DB
}
