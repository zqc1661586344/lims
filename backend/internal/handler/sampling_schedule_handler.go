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

func NewSamplingScheduleHandler(logger *zap.Logger, db *gorm.DB) *SamplingScheduleHandler {
	return &SamplingScheduleHandler{
		GenericHandler: NewGenericHandler[model.SamplingSchedule](logger, db),
		svc:            service.NewBusinessService(logger, db),
	}
}

func (h *SamplingScheduleHandler) List(c *gin.Context) {
	h.GenericHandler.List(c, nil, func(db *gorm.DB, c *gin.Context) *gorm.DB {
		if id := c.Query("task_order_id"); id != "" {
			db = db.Where("task_order_id = ?", id)
		}
		return db
	})
}

func (h *SamplingScheduleHandler) Get(c *gin.Context) {
	h.GenericHandler.Get(c)
}

func (h *SamplingScheduleHandler) Create(c *gin.Context) {
	var req struct {
		TaskOrderID    uint   `json:"task_order_id" binding:"required"`
		SamplingTeam   string `json:"sampling_team"`
		SamplingPoints string `json:"sampling_points"`
		EquipmentList  string `json:"equipment_list"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数错误: %v", err))
		return
	}
	item := model.SamplingSchedule{
		TaskOrderID:    req.TaskOrderID,
		SamplingTeam:   req.SamplingTeam,
		SamplingPoints: req.SamplingPoints,
		EquipmentList:  req.EquipmentList,
	}
	if err := h.GetDB(c).Create(&item).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("创建采样调度失败: %v", err))
		return
	}
	utils.Created(c, item)
}

func (h *SamplingScheduleHandler) Update(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	var item model.SamplingSchedule
	if err := h.GetDB(c).First(&item, id).Error; err != nil {
		utils.NotFound(c, "采样调度记录不存在")
		return
	}
	if err := h.svc.CheckNodeNotAdvanced(h.GetDB(c), "task_order", item.TaskOrderID, workflow.NodeSamplingSchedule); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	var req struct {
		SamplingTeam   string `json:"sampling_team"`
		SamplingPoints string `json:"sampling_points"`
		EquipmentList  string `json:"equipment_list"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数错误: %v", err))
		return
	}
	updates := map[string]interface{}{}
	if req.SamplingTeam != "" {
		updates["sampling_team"] = req.SamplingTeam
	}
	if req.SamplingPoints != "" {
		updates["sampling_points"] = req.SamplingPoints
	}
	if req.EquipmentList != "" {
		updates["equipment_list"] = req.EquipmentList
	}
	if err := h.GetDB(c).Model(&item).Updates(updates).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("更新采样调度失败: %v", err))
		return
	}
	h.GetDB(c).First(&item, id)
	utils.Success(c, item)
}

func (h *SamplingScheduleHandler) Delete(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	var item model.SamplingSchedule
	if err := h.GetDB(c).First(&item, id).Error; err != nil {
		utils.NotFound(c, "采样调度记录不存在")
		return
	}
	if err := h.svc.CheckNodeNotAdvanced(h.GetDB(c), "task_order", item.TaskOrderID, workflow.NodeSamplingSchedule); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	if err := h.svc.CheckInstanceRunning(h.GetDB(c), "task_order", item.TaskOrderID); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	if err := h.GetDB(c).Delete(&item).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("删除采样调度失败: %v", err))
		return
	}
	utils.Success(c, nil)
}

func (h *SamplingScheduleHandler) Approve(c *gin.Context) {
	var req struct {
		TaskID         uint   `json:"task_id" binding:"required"`
		SamplingTeam   string `json:"sampling_team"`
		SamplingPoints string `json:"sampling_points"`
		EquipmentList  string `json:"equipment_list"`
		Comment        string `json:"comment"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数错误: %v", err))
		return
	}
	sched := model.SamplingSchedule{
		TaskOrderID:    req.TaskID,
		SamplingTeam:   req.SamplingTeam,
		SamplingPoints: req.SamplingPoints,
		EquipmentList:  req.EquipmentList,
	}
	h.GetDB(c).Where("task_order_id = ?", req.TaskID).Assign(sched).FirstOrCreate(&sched)

	userID := middleware.GetUserID(c)
	if err := h.svc.ApproveTaskByOrder(req.TaskID, userID, middleware.GetDeptIDVal(c), req.Comment); err != nil {
		utils.InternalError(c, fmt.Sprintf("审批失败: %v", err))
		return
	}
	utils.Success(c, gin.H{"message": "采样调度通过"})
}

func (h *SamplingScheduleHandler) Reject(c *gin.Context) {
	var req struct {
		TaskID  uint   `json:"task_id" binding:"required"`
		Comment string `json:"comment" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数错误: %v", err))
		return
	}
	sched := model.SamplingSchedule{
		TaskOrderID: req.TaskID,
	}
	h.GetDB(c).Where("task_order_id = ?", req.TaskID).Assign(sched).FirstOrCreate(&sched)

	userID := middleware.GetUserID(c)
	if err := h.svc.RejectTaskByOrder(req.TaskID, userID, middleware.GetDeptIDVal(c), req.Comment); err != nil {
		utils.InternalError(c, fmt.Sprintf("驳回失败: %v", err))
		return
	}
	utils.Success(c, gin.H{"message": "采样调度已驳回"})
}

// ============================================================
// FieldSamplingRecordHandler — 现场采样（节点5）
// ============================================================

type FieldSamplingRecordHandler struct {
	*GenericHandler[model.FieldSamplingRecord]
	svc *service.BusinessService
}
