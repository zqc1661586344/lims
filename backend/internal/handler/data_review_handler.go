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

func NewDataReviewHandler(logger *zap.Logger, db *gorm.DB) *DataReviewHandler {
	return &DataReviewHandler{
		GenericHandler: NewGenericHandler[model.DataReview](logger, db),
		svc:            service.NewBusinessService(logger, db),
	}
}

func (h *DataReviewHandler) List(c *gin.Context) {
	h.GenericHandler.List(c, nil, func(db *gorm.DB, c *gin.Context) *gorm.DB {
		if id := c.Query("task_order_id"); id != "" {
			db = db.Where("task_order_id = ?", id)
		}
		return db
	})
}

func (h *DataReviewHandler) Get(c *gin.Context) {
	h.GenericHandler.Get(c)
}

func (h *DataReviewHandler) Create(c *gin.Context) {
	var req struct {
		TaskOrderID   uint   `json:"task_order_id" binding:"required"`
		ReviewResult  string `json:"review_result"`
		ReviewComment string `json:"review_comment"`
		IssuesFound   string `json:"issues_found"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数错误: %v", err))
		return
	}
	item := model.DataReview{
		TaskOrderID:   req.TaskOrderID,
		ReviewResult:  req.ReviewResult,
		ReviewComment: req.ReviewComment,
		IssuesFound:   req.IssuesFound,
	}
	if err := h.GetDB(c).Create(&item).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("创建数据复核失败: %v", err))
		return
	}
	utils.Created(c, item)
}

func (h *DataReviewHandler) Update(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	var item model.DataReview
	if err := h.GetDB(c).First(&item, id).Error; err != nil {
		utils.NotFound(c, "数据复核记录不存在")
		return
	}
	if err := h.svc.CheckNodeNotAdvanced(h.GetDB(c), "task_order", item.TaskOrderID, workflow.NodeDataReview); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	var req struct {
		ReviewResult  string `json:"review_result"`
		ReviewComment string `json:"review_comment"`
		IssuesFound   string `json:"issues_found"`
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
	if req.IssuesFound != "" {
		updates["issues_found"] = req.IssuesFound
	}
	if err := h.GetDB(c).Model(&item).Updates(updates).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("更新数据复核失败: %v", err))
		return
	}
	h.GetDB(c).First(&item, id)
	utils.Success(c, item)
}

func (h *DataReviewHandler) Delete(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	var item model.DataReview
	if err := h.GetDB(c).First(&item, id).Error; err != nil {
		utils.NotFound(c, "数据复核记录不存在")
		return
	}
	if err := h.svc.CheckNodeNotAdvanced(h.GetDB(c), "task_order", item.TaskOrderID, workflow.NodeDataReview); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	if err := h.svc.CheckInstanceRunning(h.GetDB(c), "task_order", item.TaskOrderID); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	if err := h.GetDB(c).Delete(&item).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("删除数据复核失败: %v", err))
		return
	}
	utils.Success(c, nil)
}

func (h *DataReviewHandler) Approve(c *gin.Context) {
	var req struct {
		TaskID        uint   `json:"task_id" binding:"required"`
		ReviewResult  string `json:"review_result"`
		ReviewComment string `json:"review_comment"`
		IssuesFound   string `json:"issues_found"`
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
		review := model.DataReview{
			TaskOrderID:   req.TaskID,
			ReviewResult:  effectiveResult,
			ReviewComment: req.ReviewComment,
			IssuesFound:   req.IssuesFound,
		}
		return tx.Where("task_order_id = ?", req.TaskID).Assign(review).FirstOrCreate(&review).Error
	}); err != nil {
		HandleWorkflowError(h.Logger, c, err, "审批失败")
		return
	}
	utils.Success(c, gin.H{"message": "数据复核通过"})
}

func (h *DataReviewHandler) Reject(c *gin.Context) {
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
		review := model.DataReview{
			TaskOrderID:   req.TaskID,
			ReviewResult:  "驳回",
			ReviewComment: req.ReviewComment,
		}
		return tx.Where("task_order_id = ?", req.TaskID).Assign(review).FirstOrCreate(&review).Error
	}, opts...); err != nil {
		HandleWorkflowError(h.Logger, c, err, "驳回失败")
		return
	}
	utils.Success(c, gin.H{"message": "数据复核已驳回"})
}

// ============================================================
// DataAuditHandler — 数据审核（节点10）
// ============================================================

type DataAuditHandler struct {
	*GenericHandler[model.DataAudit]
	svc *service.BusinessService
}
