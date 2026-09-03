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

func NewQCTaskHandler(logger *zap.Logger, db *gorm.DB) *QCTaskHandler {
	return &QCTaskHandler{
		svc: service.NewBusinessService(logger, db),
		db:  db,
	}
}

func (h *QCTaskHandler) getDB(c *gin.Context) *gorm.DB {
	if db := middleware.GetDB(c); db != nil {
		return db
	}
	return h.db
}

func (h *QCTaskHandler) List(c *gin.Context) {
	var items []model.QCTask
	query := h.getDB(c).Order("id DESC")
	if taskOrderID := c.Query("task_order_id"); taskOrderID != "" {
		query = query.Where("task_order_id = ?", taskOrderID)
	}
	if err := query.Find(&items).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("查询质控任务失败: %v", err))
		return
	}
	utils.Success(c, items)
}

func (h *QCTaskHandler) Get(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	var item model.QCTask
	if err := h.getDB(c).First(&item, id).Error; err != nil {
		utils.NotFound(c, "质控任务不存在")
		return
	}
	utils.Success(c, item)
}

func (h *QCTaskHandler) Create(c *gin.Context) {
	var req struct {
		TaskOrderID uint   `json:"task_order_id" binding:"required"`
		QCType      string `json:"qc_type"`
		QCDetails   string `json:"qc_details"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数错误: %v", err))
		return
	}
	item := model.QCTask{
		TaskOrderID: req.TaskOrderID,
		QCType:      req.QCType,
		QCDetails:   req.QCDetails,
	}
	if err := h.getDB(c).Create(&item).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("创建质控任务失败: %v", err))
		return
	}
	utils.Created(c, item)
}

func (h *QCTaskHandler) Update(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	var item model.QCTask
	if err := h.getDB(c).First(&item, id).Error; err != nil {
		utils.NotFound(c, "质控任务不存在")
		return
	}
	var req struct {
		QCType    string `json:"qc_type"`
		QCDetails string `json:"qc_details"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数错误: %v", err))
		return
	}
	updates := map[string]interface{}{}
	if req.QCType != "" {
		updates["qc_type"] = req.QCType
	}
	if req.QCDetails != "" {
		updates["qc_details"] = req.QCDetails
	}
	if err := h.getDB(c).Model(&item).Updates(updates).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("更新质控任务失败: %v", err))
		return
	}
	h.getDB(c).First(&item, id)
	utils.Success(c, item)
}

func (h *QCTaskHandler) Delete(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	if err := h.getDB(c).Delete(&model.QCTask{}, id).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("删除质控任务失败: %v", err))
		return
	}
	utils.Success(c, nil)
}

func (h *QCTaskHandler) Approve(c *gin.Context) {
	var req struct {
		TaskID    uint   `json:"task_id" binding:"required"`
		QCType    string `json:"qc_type"`
		QCDetails string `json:"qc_details"`
		Comment   string `json:"comment"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数错误: %v", err))
		return
	}
	qc := model.QCTask{
		TaskOrderID: req.TaskID,
		QCType:      req.QCType,
		QCDetails:   req.QCDetails,
	}
	h.getDB(c).Where("task_order_id = ?", req.TaskID).Assign(qc).FirstOrCreate(&qc)

	userID := middleware.GetUserID(c)
	if err := h.svc.ApproveTaskByOrder(req.TaskID, userID, req.Comment); err != nil {
		utils.InternalError(c, fmt.Sprintf("审批失败: %v", err))
		return
	}
	utils.Success(c, gin.H{"message": "质控任务通过"})
}

func (h *QCTaskHandler) Reject(c *gin.Context) {
	var req struct {
		TaskID  uint   `json:"task_id" binding:"required"`
		Comment string `json:"comment" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数错误: %v", err))
		return
	}
	qc := model.QCTask{
		TaskOrderID: req.TaskID,
	}
	h.getDB(c).Where("task_order_id = ?", req.TaskID).Assign(qc).FirstOrCreate(&qc)

	userID := middleware.GetUserID(c)
	if err := h.svc.RejectTaskByOrder(req.TaskID, userID, req.Comment); err != nil {
		utils.InternalError(c, fmt.Sprintf("驳回失败: %v", err))
		return
	}
	utils.Success(c, gin.H{"message": "质控任务已驳回"})
}

// ============================================================
// SamplingScheduleHandler — 采样调度（节点4）
// ============================================================

type SamplingScheduleHandler struct {
	svc *service.BusinessService
	db  *gorm.DB
}
