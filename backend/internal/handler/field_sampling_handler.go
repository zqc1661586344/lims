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

func NewFieldSamplingRecordHandler(logger *zap.Logger, db *gorm.DB) *FieldSamplingRecordHandler {
	return &FieldSamplingRecordHandler{
		svc: service.NewBusinessService(logger, db),
		db:  db,
	}
}

func (h *FieldSamplingRecordHandler) getDB(c *gin.Context) *gorm.DB {
	if db := middleware.GetDB(c); db != nil {
		return db
	}
	return h.db
}

func (h *FieldSamplingRecordHandler) List(c *gin.Context) {
	var items []model.FieldSamplingRecord
	query := h.getDB(c).Order("id DESC")
	if taskOrderID := c.Query("task_order_id"); taskOrderID != "" {
		query = query.Where("task_order_id = ?", taskOrderID)
	}
	if err := query.Find(&items).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("查询现场采样记录失败: %v", err))
		return
	}
	utils.Success(c, items)
}

func (h *FieldSamplingRecordHandler) Get(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	var item model.FieldSamplingRecord
	if err := h.getDB(c).First(&item, id).Error; err != nil {
		utils.NotFound(c, "现场采样记录不存在")
		return
	}
	utils.Success(c, item)
}

func (h *FieldSamplingRecordHandler) Create(c *gin.Context) {
	var req struct {
		TaskOrderID            uint   `json:"task_order_id" binding:"required"`
		SamplePhotos           string `json:"sample_photos"`
		EquipmentCalRecords    string `json:"equipment_cal_records"`
		SamplingRecordFilePath string `json:"sampling_record_file_path"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数错误: %v", err))
		return
	}
	item := model.FieldSamplingRecord{
		TaskOrderID:            req.TaskOrderID,
		SamplePhotos:           req.SamplePhotos,
		EquipmentCalRecords:    req.EquipmentCalRecords,
		SamplingRecordFilePath: req.SamplingRecordFilePath,
	}
	if err := h.getDB(c).Create(&item).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("创建现场采样记录失败: %v", err))
		return
	}
	utils.Created(c, item)
}

func (h *FieldSamplingRecordHandler) Update(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	var item model.FieldSamplingRecord
	if err := h.getDB(c).First(&item, id).Error; err != nil {
		utils.NotFound(c, "现场采样记录不存在")
		return
	}
	if err := h.svc.CheckNodeNotAdvanced(h.getDB(c), "task_order", item.TaskOrderID, workflow.NodeFieldSampling); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	var req struct {
		SamplePhotos           string `json:"sample_photos"`
		EquipmentCalRecords    string `json:"equipment_cal_records"`
		SamplingRecordFilePath string `json:"sampling_record_file_path"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数错误: %v", err))
		return
	}
	updates := map[string]interface{}{}
	if req.SamplePhotos != "" {
		updates["sample_photos"] = req.SamplePhotos
	}
	if req.EquipmentCalRecords != "" {
		updates["equipment_cal_records"] = req.EquipmentCalRecords
	}
	if req.SamplingRecordFilePath != "" {
		updates["sampling_record_file_path"] = req.SamplingRecordFilePath
	}
	if err := h.getDB(c).Model(&item).Updates(updates).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("更新现场采样记录失败: %v", err))
		return
	}
	h.getDB(c).First(&item, id)
	utils.Success(c, item)
}

func (h *FieldSamplingRecordHandler) Delete(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	var item model.FieldSamplingRecord
	if err := h.getDB(c).First(&item, id).Error; err != nil {
		utils.NotFound(c, "现场采样记录不存在")
		return
	}
	if err := h.svc.CheckInstanceRunning(h.getDB(c), "task_order", item.TaskOrderID); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	if err := h.getDB(c).Delete(&item).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("删除现场采样记录失败: %v", err))
		return
	}
	utils.Success(c, nil)
}

func (h *FieldSamplingRecordHandler) Approve(c *gin.Context) {
	var req struct {
		TaskID                 uint   `json:"task_id" binding:"required"`
		SamplePhotos           string `json:"sample_photos"`
		EquipmentCalRecords    string `json:"equipment_cal_records"`
		SamplingRecordFilePath string `json:"sampling_record_file_path"`
		Comment                string `json:"comment"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数错误: %v", err))
		return
	}
	rec := model.FieldSamplingRecord{
		TaskOrderID:            req.TaskID,
		SamplePhotos:           req.SamplePhotos,
		EquipmentCalRecords:    req.EquipmentCalRecords,
		SamplingRecordFilePath: req.SamplingRecordFilePath,
	}
	h.getDB(c).Where("task_order_id = ?", req.TaskID).Assign(rec).FirstOrCreate(&rec)

	userID := middleware.GetUserID(c)
	if err := h.svc.ApproveTaskByOrder(req.TaskID, userID, middleware.GetDeptIDVal(c), req.Comment); err != nil {
		utils.InternalError(c, fmt.Sprintf("审批失败: %v", err))
		return
	}
	utils.Success(c, gin.H{"message": "现场采样通过"})
}

func (h *FieldSamplingRecordHandler) Reject(c *gin.Context) {
	var req struct {
		TaskID  uint   `json:"task_id" binding:"required"`
		Comment string `json:"comment" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数错误: %v", err))
		return
	}
	rec := model.FieldSamplingRecord{
		TaskOrderID: req.TaskID,
	}
	h.getDB(c).Where("task_order_id = ?", req.TaskID).Assign(rec).FirstOrCreate(&rec)

	userID := middleware.GetUserID(c)
	if err := h.svc.RejectTaskByOrder(req.TaskID, userID, middleware.GetDeptIDVal(c), req.Comment); err != nil {
		utils.InternalError(c, fmt.Sprintf("驳回失败: %v", err))
		return
	}
	utils.Success(c, gin.H{"message": "现场采样已驳回"})
}

// ============================================================
// SampleReceivingHandler — 样品接收（节点6）
// ============================================================

type SampleReceivingHandler struct {
	svc *service.BusinessService
	db  *gorm.DB
}
