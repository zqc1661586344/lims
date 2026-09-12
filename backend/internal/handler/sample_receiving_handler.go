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
		GenericHandler: NewGenericHandler[model.SampleReceiving](logger, db),
		svc:            service.NewBusinessService(logger, db),
	}
}

func (h *SampleReceivingHandler) List(c *gin.Context) {
	h.GenericHandler.List(c, nil, func(db *gorm.DB, c *gin.Context) *gorm.DB {
		if id := c.Query("task_order_id"); id != "" {
			db = db.Where("task_order_id = ?", id)
		}
		return db
	})
}

func (h *SampleReceivingHandler) Get(c *gin.Context) {
	h.GenericHandler.Get(c)
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
	if err := h.GetDB(c).Create(&item).Error; err != nil {
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
	if err := h.GetDB(c).First(&item, id).Error; err != nil {
		utils.NotFound(c, "样品接收记录不存在")
		return
	}
	if err := h.svc.CheckNodeNotAdvanced(h.GetDB(c), "task_order", item.TaskOrderID, workflow.NodeSampleReceiving); err != nil {
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
	if err := h.GetDB(c).Model(&item).Updates(updates).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("更新样品接收记录失败: %v", err))
		return
	}
	h.GetDB(c).First(&item, id)
	utils.Success(c, item)
}

func (h *SampleReceivingHandler) Delete(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	var item model.SampleReceiving
	if err := h.GetDB(c).First(&item, id).Error; err != nil {
		utils.NotFound(c, "样品接收记录不存在")
		return
	}
	if err := h.svc.CheckNodeNotAdvanced(h.GetDB(c), "task_order", item.TaskOrderID, workflow.NodeSampleReceiving); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	if err := h.svc.CheckInstanceRunning(h.GetDB(c), "task_order", item.TaskOrderID); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	if err := h.GetDB(c).Delete(&item).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("删除样品接收失败: %v", err))
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

	userID := middleware.GetUserID(c)
	if err := h.svc.ApproveWithBusiness(req.TaskID, userID, middleware.GetDeptIDVal(c), req.Comment, func(tx *gorm.DB) error {
		rec := model.SampleReceiving{
			TaskOrderID:         req.TaskID,
			SampleCondition:     req.SampleCondition,
			SampleCodes:         req.SampleCodes,
			ReceivingRecordPath: req.ReceivingRecordPath,
		}
		return tx.Where("task_order_id = ?", req.TaskID).Assign(rec).FirstOrCreate(&rec).Error
	}); err != nil {
		HandleWorkflowError(c, err, "审批失败")
		return
	}
	utils.Success(c, gin.H{"message": "样品接收通过"})
}

func (h *SampleReceivingHandler) Reject(c *gin.Context) {
	var req struct {
		TaskID       uint   `json:"task_id" binding:"required"`
		Comment      string `json:"comment" binding:"required"`
		RejectTarget string `json:"reject_target"`
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
	if err := h.svc.RejectWithBusiness(req.TaskID, userID, middleware.GetDeptIDVal(c), req.Comment, func(tx *gorm.DB) error {
		rec := model.SampleReceiving{TaskOrderID: req.TaskID}
		return tx.Where("task_order_id = ?", req.TaskID).Assign(rec).FirstOrCreate(&rec).Error
	}, opts...); err != nil {
		HandleWorkflowError(c, err, "驳回失败")
		return
	}
	utils.Success(c, gin.H{"message": "样品接收已驳回"})
}

// ============================================================
// TaskAssignHandler — 任务分配（节点7）
// ============================================================

type TaskAssignHandler struct {
	*GenericHandler[model.TaskAssign]
	svc *service.BusinessService
}
