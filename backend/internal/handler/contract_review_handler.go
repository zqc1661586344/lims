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

func NewContractReviewHandler(logger *zap.Logger, db *gorm.DB) *ContractReviewHandler {
	return &ContractReviewHandler{
		GenericHandler: NewGenericHandler[model.ContractReview](logger, db),
		svc:            service.NewBusinessService(logger, db),
	}
}

// List 返回合同评审列表
func (h *ContractReviewHandler) List(c *gin.Context) {
	h.GenericHandler.List(c, nil, func(db *gorm.DB, c *gin.Context) *gorm.DB {
		db = service.ApplyTaskOrderScope(db, c)
		if id := c.Query("task_order_id"); id != "" {
			db = db.Where("task_order_id = ?", id)
		}
		return db
	})
}

// Get 获取单个合同评审
func (h *ContractReviewHandler) Get(c *gin.Context) {
	h.GenericHandler.GetWithScope(c, service.ApplyTaskOrderScope)
}

type SaveContractReviewRequest struct {
	TaskOrderID      uint   `json:"task_order_id" binding:"required"`
	ReviewResult     string `json:"review_result"`
	ReviewComment    string `json:"review_comment"`
	ContractFilePath string `json:"contract_file_path"`
}

// Create 创建合同评审
func (h *ContractReviewHandler) Create(c *gin.Context) {
	var req SaveContractReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数错误: %v", err))
		return
	}
	item := model.ContractReview{
		TaskOrderID:      req.TaskOrderID,
		ReviewResult:     req.ReviewResult,
		ReviewComment:    req.ReviewComment,
		ContractFilePath: req.ContractFilePath,
	}
	if err := h.GetDB(c).Create(&item).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("创建合同评审失败: %v", err))
		return
	}
	utils.Created(c, item)
}

// Update 更新合同评审
func (h *ContractReviewHandler) Update(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	var item model.ContractReview
	if err := h.GetDB(c).First(&item, id).Error; err != nil {
		utils.NotFound(c, "合同评审记录不存在")
		return
	}
	if err := h.svc.CheckNodeNotAdvanced(h.GetDB(c), "task_order", item.TaskOrderID, workflow.NodeContractReview); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	var req SaveContractReviewRequest
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
	if req.ContractFilePath != "" {
		updates["contract_file_path"] = req.ContractFilePath
	}
	if err := h.GetDB(c).Model(&item).Updates(updates).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("更新合同评审失败: %v", err))
		return
	}
	h.GetDB(c).First(&item, id)
	utils.Success(c, item)
}

// Delete 删除合同评审
func (h *ContractReviewHandler) Delete(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	var item model.ContractReview
	if err := h.GetDB(c).First(&item, id).Error; err != nil {
		utils.NotFound(c, "合同评审记录不存在")
		return
	}
	if err := h.svc.CheckNodeNotAdvanced(h.GetDB(c), "task_order", item.TaskOrderID, workflow.NodeContractReview); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	if err := h.svc.CheckInstanceRunning(h.GetDB(c), "task_order", item.TaskOrderID); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	if err := h.GetDB(c).Delete(&item).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("删除合同评审失败: %v", err))
		return
	}
	utils.Success(c, nil)
}

// Approve 通过合同评审并推进流程
func (h *ContractReviewHandler) Approve(c *gin.Context) {
	var req struct {
		TaskID           uint   `json:"task_id" binding:"required"`
		ReviewResult     string `json:"review_result"`
		ReviewComment    string `json:"review_comment"`
		ContractFilePath string `json:"contract_file_path"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数错误: %v", err))
		return
	}

	userID := middleware.GetUserID(c)
	var effectiveResult = "通过"
	if req.ReviewResult != "" {
		effectiveResult = req.ReviewResult
	}
	if err := h.svc.ApproveWithBusiness(req.TaskID, userID, middleware.GetDeptIDVal(c), req.ReviewComment, func(tx *gorm.DB) error {
		review := model.ContractReview{
			TaskOrderID:      req.TaskID,
			ReviewResult:     effectiveResult,
			ReviewComment:    req.ReviewComment,
			ContractFilePath: req.ContractFilePath,
		}
		return tx.Where("task_order_id = ?", req.TaskID).Assign(review).FirstOrCreate(&review).Error
	}); err != nil {
		HandleWorkflowError(h.Logger, c, err, "审批失败")
		return
	}
	utils.Success(c, gin.H{"message": "合同评审通过"})
}

// Reject 驳回合同评审
func (h *ContractReviewHandler) Reject(c *gin.Context) {
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
		review := model.ContractReview{
			TaskOrderID:   req.TaskID,
			ReviewResult:  "驳回",
			ReviewComment: req.ReviewComment,
		}
		return tx.Where("task_order_id = ?", req.TaskID).Assign(review).FirstOrCreate(&review).Error
	}, opts...); err != nil {
		HandleWorkflowError(h.Logger, c, err, "驳回失败")
		return
	}
	utils.Success(c, gin.H{"message": "合同评审已驳回"})
}

// ============================================================
// QCTaskHandler — 质控任务（节点3）
// ============================================================

type QCTaskHandler struct {
	*GenericHandler[model.QCTask]
	svc *service.BusinessService
}
