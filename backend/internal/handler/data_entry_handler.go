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

func NewDataEntryHandler(logger *zap.Logger, db *gorm.DB) *DataEntryHandler {
	return &DataEntryHandler{
		svc: service.NewBusinessService(logger, db),
		db:  db,
	}
}

func (h *DataEntryHandler) getDB(c *gin.Context) *gorm.DB {
	if db := middleware.GetDB(c); db != nil {
		return db
	}
	return h.db
}

func (h *DataEntryHandler) List(c *gin.Context) {
	var items []model.DataEntry
	query := h.getDB(c).Order("id DESC")
	if taskOrderID := c.Query("task_order_id"); taskOrderID != "" {
		query = query.Where("task_order_id = ?", taskOrderID)
	}
	if testItemID := c.Query("test_item_id"); testItemID != "" {
		query = query.Where("test_item_id = ?", testItemID)
	}
	if err := query.Find(&items).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("查询数据录入失败: %v", err))
		return
	}
	utils.Success(c, items)
}

func (h *DataEntryHandler) Get(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	var item model.DataEntry
	if err := h.getDB(c).First(&item, id).Error; err != nil {
		utils.NotFound(c, "数据录入记录不存在")
		return
	}
	utils.Success(c, item)
}

func (h *DataEntryHandler) Create(c *gin.Context) {
	var req struct {
		TaskOrderID  uint   `json:"task_order_id" binding:"required"`
		TestItemID   uint   `json:"test_item_id"`
		OriginalData string `json:"original_data"`
		RawRecordID  *uint  `json:"raw_record_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数错误: %v", err))
		return
	}
	item := model.DataEntry{
		TaskOrderID:  req.TaskOrderID,
		TestItemID:   req.TestItemID,
		OriginalData: req.OriginalData,
		RawRecordID:  req.RawRecordID,
	}
	if err := h.getDB(c).Create(&item).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("创建数据录入失败: %v", err))
		return
	}
	utils.Created(c, item)
}

func (h *DataEntryHandler) Update(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	var item model.DataEntry
	if err := h.getDB(c).First(&item, id).Error; err != nil {
		utils.NotFound(c, "数据录入记录不存在")
		return
	}
	var req struct {
		TestItemID   uint   `json:"test_item_id"`
		OriginalData string `json:"original_data"`
		RawRecordID  *uint  `json:"raw_record_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数错误: %v", err))
		return
	}
	updates := map[string]interface{}{}
	if req.OriginalData != "" {
		updates["original_data"] = req.OriginalData
	}
	// Always update test_item_id and raw_record_id if set
	updates["test_item_id"] = req.TestItemID
	updates["raw_record_id"] = req.RawRecordID
	if err := h.getDB(c).Model(&item).Updates(updates).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("更新数据录入失败: %v", err))
		return
	}
	h.getDB(c).First(&item, id)
	utils.Success(c, item)
}

func (h *DataEntryHandler) Delete(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	if err := h.getDB(c).Delete(&model.DataEntry{}, id).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("删除数据录入失败: %v", err))
		return
	}
	utils.Success(c, nil)
}

// Approve 数据录入审批 — 不创建 DataEntry 记录（1-to-many 关系）
func (h *DataEntryHandler) Approve(c *gin.Context) {
	var req struct {
		TaskID  uint   `json:"task_id" binding:"required"`
		Comment string `json:"comment"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数错误: %v", err))
		return
	}
	// 检查是否有至少一条数据录入记录
	var count int64
	h.getDB(c).Model(&model.DataEntry{}).Where("task_order_id = ?", req.TaskID).Count(&count)
	if count == 0 {
		utils.BadRequest(c, "请先录入数据")
		return
	}

	userID := middleware.GetUserID(c)
	if err := h.svc.ApproveTaskByOrder(req.TaskID, userID, req.Comment); err != nil {
		utils.InternalError(c, fmt.Sprintf("审批失败: %v", err))
		return
	}
	utils.Success(c, gin.H{"message": "数据录入通过"})
}

func (h *DataEntryHandler) Reject(c *gin.Context) {
	var req struct {
		TaskID  uint   `json:"task_id" binding:"required"`
		Comment string `json:"comment" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数错误: %v", err))
		return
	}
	userID := middleware.GetUserID(c)
	if err := h.svc.RejectTaskByOrder(req.TaskID, userID, req.Comment); err != nil {
		utils.InternalError(c, fmt.Sprintf("驳回失败: %v", err))
		return
	}
	utils.Success(c, gin.H{"message": "数据录入已驳回"})
}

// ============================================================
// DataReviewHandler — 数据复核（节点9）
// ============================================================

type DataReviewHandler struct {
	svc *service.BusinessService
	db  *gorm.DB
}
