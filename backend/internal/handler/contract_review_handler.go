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
		svc: service.NewBusinessService(logger, db),
		db:  db,
	}
}

func (h *ContractReviewHandler) getDB(c *gin.Context) *gorm.DB {
	if db := middleware.GetDB(c); db != nil {
		return db
	}
	return h.db
}

// List 返回合同评审列表
func (h *ContractReviewHandler) List(c *gin.Context) {
	page, pageSize, offset := utils.GetPagination(c)
	var items []model.ContractReview
	query := h.getDB(c).Order("id DESC")
	if taskOrderID := c.Query("task_order_id"); taskOrderID != "" {
		query = query.Where("task_order_id = ?", taskOrderID)
	}
	var total int64
	if err := query.Model(&model.ContractReview{}).Count(&total).Error; err != nil {
		utils.InternalError(c, "查询失败")
		return
	}
	if err := query.Limit(pageSize).Offset(offset).Find(&items).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("查询合同评审失败: %v", err))
		return
	}
	utils.SuccessPage(c, items, total, page, pageSize)
}

// Get 获取单个合同评审
func (h *ContractReviewHandler) Get(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	var item model.ContractReview
	if err := h.getDB(c).First(&item, id).Error; err != nil {
		utils.NotFound(c, "合同评审记录不存在")
		return
	}
	utils.Success(c, item)
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
	if err := h.getDB(c).Create(&item).Error; err != nil {
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
	if err := h.getDB(c).First(&item, id).Error; err != nil {
		utils.NotFound(c, "合同评审记录不存在")
		return
	}
	if err := h.svc.CheckNodeNotAdvanced(h.getDB(c), "task_order", item.TaskOrderID, workflow.NodeContractReview); err != nil {
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
	if err := h.getDB(c).Model(&item).Updates(updates).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("更新合同评审失败: %v", err))
		return
	}
	h.getDB(c).First(&item, id)
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
	if err := h.getDB(c).First(&item, id).Error; err != nil {
		utils.NotFound(c, "合同评审记录不存在")
		return
	}
	if err := h.svc.CheckInstanceRunning(h.getDB(c), "task_order", item.TaskOrderID); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	if err := h.getDB(c).Delete(&item).Error; err != nil {
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
	// 保存或更新合同评审记录
	review := model.ContractReview{
		TaskOrderID:      req.TaskID, // 使用 task_id 关联的业务ID
		ReviewResult:     "通过",
		ReviewComment:    req.ReviewComment,
		ContractFilePath: req.ContractFilePath,
	}
	if req.ReviewResult != "" {
		review.ReviewResult = req.ReviewResult
	}
	h.getDB(c).Where("task_order_id = ?", req.TaskID).Assign(review).FirstOrCreate(&review)

	userID := middleware.GetUserID(c)
	if err := h.svc.ApproveTaskByOrder(req.TaskID, userID, middleware.GetDeptIDVal(c), req.ReviewComment); err != nil {
		utils.InternalError(c, fmt.Sprintf("审批失败: %v", err))
		return
	}
	utils.Success(c, gin.H{"message": "合同评审通过"})
}

// Reject 驳回合同评审
func (h *ContractReviewHandler) Reject(c *gin.Context) {
	var req struct {
		TaskID        uint   `json:"task_id" binding:"required"`
		ReviewComment string `json:"review_comment" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数错误: %v", err))
		return
	}
	// 保存驳回记录
	review := model.ContractReview{
		TaskOrderID:   req.TaskID,
		ReviewResult:  "驳回",
		ReviewComment: req.ReviewComment,
	}
	h.getDB(c).Where("task_order_id = ?", req.TaskID).Assign(review).FirstOrCreate(&review)

	userID := middleware.GetUserID(c)
	if err := h.svc.RejectTaskByOrder(req.TaskID, userID, middleware.GetDeptIDVal(c), req.ReviewComment); err != nil {
		utils.InternalError(c, fmt.Sprintf("驳回失败: %v", err))
		return
	}
	utils.Success(c, gin.H{"message": "合同评审已驳回"})
}

// ============================================================
// QCTaskHandler — 质控任务（节点3）
// ============================================================

type QCTaskHandler struct {
	svc *service.BusinessService
	db  *gorm.DB
}
