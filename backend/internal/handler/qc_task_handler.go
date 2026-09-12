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

func NewQCTaskHandler(logger *zap.Logger, db *gorm.DB) *QCTaskHandler {
	return &QCTaskHandler{
		GenericHandler: NewGenericHandler[model.QCTask](logger, db),
		svc:            service.NewBusinessService(logger, db),
	}
}

func (h *QCTaskHandler) List(c *gin.Context) {
	h.GenericHandler.List(c, nil, func(db *gorm.DB, c *gin.Context) *gorm.DB {
		if id := c.Query("task_order_id"); id != "" {
			db = db.Where("task_order_id = ?", id)
		}
		return db
	})
}

func (h *QCTaskHandler) Get(c *gin.Context) {
	h.GenericHandler.Get(c)
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
	if err := h.GetDB(c).Create(&item).Error; err != nil {
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
	if err := h.GetDB(c).First(&item, id).Error; err != nil {
		utils.NotFound(c, "质控任务不存在")
		return
	}
	if err := h.svc.CheckNodeNotAdvanced(h.GetDB(c), "task_order", item.TaskOrderID, workflow.NodeQCTask); err != nil {
		utils.BadRequest(c, err.Error())
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
	if err := h.GetDB(c).Model(&item).Updates(updates).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("更新质控任务失败: %v", err))
		return
	}
	h.GetDB(c).First(&item, id)
	utils.Success(c, item)
}

func (h *QCTaskHandler) Delete(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	var item model.QCTask
	if err := h.GetDB(c).First(&item, id).Error; err != nil {
		utils.NotFound(c, "质控任务不存在")
		return
	}
	if err := h.svc.CheckNodeNotAdvanced(h.GetDB(c), "task_order", item.TaskOrderID, workflow.NodeQCTask); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	if err := h.svc.CheckInstanceRunning(h.GetDB(c), "task_order", item.TaskOrderID); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	if err := h.GetDB(c).Delete(&item).Error; err != nil {
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

	userID := middleware.GetUserID(c)
	if err := h.svc.ApproveWithBusiness(req.TaskID, userID, middleware.GetDeptIDVal(c), req.Comment, func(tx *gorm.DB) error {
		qc := model.QCTask{
			TaskOrderID: req.TaskID,
			QCType:      req.QCType,
			QCDetails:   req.QCDetails,
		}
		return tx.Where("task_order_id = ?", req.TaskID).Assign(qc).FirstOrCreate(&qc).Error
	}); err != nil {
		HandleWorkflowError(h.Logger, c, err, "审批失败")
		return
	}
	utils.Success(c, gin.H{"message": "质控任务通过"})
}

func (h *QCTaskHandler) Reject(c *gin.Context) {
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
		qc := model.QCTask{TaskOrderID: req.TaskID}
		return tx.Where("task_order_id = ?", req.TaskID).Assign(qc).FirstOrCreate(&qc).Error
	}, opts...); err != nil {
		HandleWorkflowError(h.Logger, c, err, "驳回失败")
		return
	}
	utils.Success(c, gin.H{"message": "质控任务已驳回"})
}

// ============================================================
// SamplingScheduleHandler — 采样调度（节点4）
// ============================================================

type SamplingScheduleHandler struct {
	*GenericHandler[model.SamplingSchedule]
	svc *service.BusinessService
}
