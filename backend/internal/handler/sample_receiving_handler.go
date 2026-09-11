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

func NewSampleReceivingHandler(logger *zap.Logger, db *gorm.DB) *SampleReceivingHandler {
	return &SampleReceivingHandler{
		svc: service.NewBusinessService(logger, db),
		db:  db,
	}
}

func (h *SampleReceivingHandler) getDB(c *gin.Context) *gorm.DB {
	if db := middleware.GetDB(c); db != nil {
		return db
	}
	return h.db
}

func (h *SampleReceivingHandler) List(c *gin.Context) {
	page, pageSize, offset := utils.GetPagination(c)
	var items []model.SampleReceiving
	query := h.getDB(c).Order("id DESC")
	if taskOrderID := c.Query("task_order_id"); taskOrderID != "" {
		query = query.Where("task_order_id = ?", taskOrderID)
	}
	var total int64
	if err := query.Model(&model.SampleReceiving{}).Count(&total).Error; err != nil {
		utils.InternalError(c, "查询失败")
		return
	}
	if err := query.Limit(pageSize).Offset(offset).Find(&items).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("查询样品接收记录失败: %v", err))
		return
	}
	utils.SuccessPage(c, items, total, page, pageSize)
}

func (h *SampleReceivingHandler) Get(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	var item model.SampleReceiving
	if err := h.getDB(c).First(&item, id).Error; err != nil {
		utils.NotFound(c, "样品接收记录不存在")
		return
	}
	utils.Success(c, item)
}

func (h *SampleReceivingHandler) Create(c *gin.Context) {
	var req struct {
		TaskOrderID         uint   `json:"task_order_id" binding:"required"`
		SampleCondition     string `json:"sample_condition"`
		SampleCodes         string `json:"sample_codes"`
		ReceivingRecordPath string `json:"receiving_record_path"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数错误: %v", err))
		return
	}
	item := model.SampleReceiving{
		TaskOrderID:         req.TaskOrderID,
		SampleCondition:     req.SampleCondition,
		SampleCodes:         req.SampleCodes,
		ReceivingRecordPath: req.ReceivingRecordPath,
	}
	if err := h.getDB(c).Create(&item).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("创建样品接收记录失败: %v", err))
		return
	}
	utils.Created(c, item)
}

func (h *SampleReceivingHandler) Update(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	var item model.SampleReceiving
	if err := h.getDB(c).First(&item, id).Error; err != nil {
		utils.NotFound(c, "样品接收记录不存在")
		return
	}
	if err := h.svc.CheckNodeNotAdvanced(h.getDB(c), "task_order", item.TaskOrderID, workflow.NodeSampleReceiving); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	var req struct {
		SampleCondition     string `json:"sample_condition"`
		SampleCodes         string `json:"sample_codes"`
		ReceivingRecordPath string `json:"receiving_record_path"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数错误: %v", err))
		return
	}
	updates := map[string]interface{}{}
	if req.SampleCondition != "" {
		updates["sample_condition"] = req.SampleCondition
	}
	if req.SampleCodes != "" {
		updates["sample_codes"] = req.SampleCodes
	}
	if req.ReceivingRecordPath != "" {
		updates["receiving_record_path"] = req.ReceivingRecordPath
	}
	if err := h.getDB(c).Model(&item).Updates(updates).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("更新样品接收记录失败: %v", err))
		return
	}
	h.getDB(c).First(&item, id)
	utils.Success(c, item)
}

func (h *SampleReceivingHandler) Delete(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	var item model.SampleReceiving
	if err := h.getDB(c).First(&item, id).Error; err != nil {
		utils.NotFound(c, "样品接收记录不存在")
		return
	}
	if err := h.svc.CheckInstanceRunning(h.getDB(c), "task_order", item.TaskOrderID); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	if err := h.getDB(c).Delete(&item).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("删除样品接收记录失败: %v", err))
		return
	}
	utils.Success(c, nil)
}

func (h *SampleReceivingHandler) Approve(c *gin.Context) {
	var req struct {
		TaskID              uint   `json:"task_id" binding:"required"`
		SampleCondition     string `json:"sample_condition"`
		SampleCodes         string `json:"sample_codes"`
		ReceivingRecordPath string `json:"receiving_record_path"`
		Comment             string `json:"comment"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数错误: %v", err))
		return
	}
	rec := model.SampleReceiving{
		TaskOrderID:         req.TaskID,
		SampleCondition:     req.SampleCondition,
		SampleCodes:         req.SampleCodes,
		ReceivingRecordPath: req.ReceivingRecordPath,
	}
	h.getDB(c).Where("task_order_id = ?", req.TaskID).Assign(rec).FirstOrCreate(&rec)

	userID := middleware.GetUserID(c)
	if err := h.svc.ApproveTaskByOrder(req.TaskID, userID, middleware.GetDeptIDVal(c), req.Comment); err != nil {
		utils.InternalError(c, fmt.Sprintf("审批失败: %v", err))
		return
	}
	utils.Success(c, gin.H{"message": "样品接收通过"})
}

func (h *SampleReceivingHandler) Reject(c *gin.Context) {
	var req struct {
		TaskID  uint   `json:"task_id" binding:"required"`
		Comment string `json:"comment" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数错误: %v", err))
		return
	}
	rec := model.SampleReceiving{
		TaskOrderID: req.TaskID,
	}
	h.getDB(c).Where("task_order_id = ?", req.TaskID).Assign(rec).FirstOrCreate(&rec)

	userID := middleware.GetUserID(c)
	if err := h.svc.RejectTaskByOrder(req.TaskID, userID, middleware.GetDeptIDVal(c), req.Comment); err != nil {
		utils.InternalError(c, fmt.Sprintf("驳回失败: %v", err))
		return
	}
	utils.Success(c, gin.H{"message": "样品接收已驳回"})
}

// ============================================================
// TaskAssignHandler — 任务分配（节点7）
// ============================================================

type TaskAssignHandler struct {
	svc *service.BusinessService
	db  *gorm.DB
}
