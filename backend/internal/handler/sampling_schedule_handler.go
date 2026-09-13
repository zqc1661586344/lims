package handler

import (
	"encoding/json"
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
		SamplingPoints: model.JSONB(req.SamplingPoints),
		EquipmentList:  model.JSONB(req.EquipmentList),
	}
	if err := h.GetDB(c).Create(&item).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("创建采样调度失败: %v", err))
		return
	}
	if err := syncSamplingPoints(h.GetDB(c), item.ID, req.SamplingPoints); err != nil {
		h.Logger.Warn("failed to sync sampling points to new table", zap.Error(err), zap.Uint("schedule_id", item.ID))
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
	if req.SamplingPoints != "" {
		if err := syncSamplingPoints(h.GetDB(c), item.ID, req.SamplingPoints); err != nil {
			h.Logger.Warn("failed to sync sampling points on update", zap.Error(err), zap.Uint("schedule_id", item.ID))
		}
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

	userID := middleware.GetUserID(c)
	if err := h.svc.ApproveWithBusiness(req.TaskID, userID, middleware.GetDeptIDVal(c), req.Comment, func(tx *gorm.DB) error {
		sched := model.SamplingSchedule{
			TaskOrderID:    req.TaskID,
			SamplingTeam:   req.SamplingTeam,
			SamplingPoints: model.JSONB(req.SamplingPoints),
			EquipmentList:  model.JSONB(req.EquipmentList),
		}
		if err := tx.Where("task_order_id = ?", req.TaskID).Assign(sched).FirstOrCreate(&sched).Error; err != nil {
			return err
		}
		if req.SamplingPoints != "" {
			return syncSamplingPoints(tx, sched.ID, req.SamplingPoints)
		}
		return nil
	}); err != nil {
		HandleWorkflowError(h.Logger, c, err, "审批失败")
		return
	}
	utils.Success(c, gin.H{"message": "采样调度通过"})
}

func (h *SamplingScheduleHandler) Reject(c *gin.Context) {
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
		sched := model.SamplingSchedule{TaskOrderID: req.TaskID}
		return tx.Where("task_order_id = ?", req.TaskID).Assign(sched).FirstOrCreate(&sched).Error
	}, opts...); err != nil {
		HandleWorkflowError(h.Logger, c, err, "驳回失败")
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

type legacySamplingPoint struct {
	Name           string  `json:"name"`
	PointName      string  `json:"point_name"`
	Code           string  `json:"code"`
	PointCode      string  `json:"point_code"`
	Location       string  `json:"location"`
	Longitude      float64 `json:"longitude"`
	Latitude       float64 `json:"latitude"`
	SamplingMethod string  `json:"sampling_method"`
	SampleCount    int     `json:"sample_count"`
	Type           string  `json:"type"`
	Count          int     `json:"count"`
}

func syncSamplingPoints(db *gorm.DB, scheduleID uint, raw string) error {
	if raw == "" || raw == "{}" || raw == "[]" {
		return nil
	}
	var rawArr []map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &rawArr); err != nil {
		return err
	}
	if len(rawArr) == 0 {
		return nil
	}
	if err := db.Where("sampling_schedule_id = ?", scheduleID).Delete(&model.SamplingPoint{}).Error; err != nil {
		return err
	}
	for _, m := range rawArr {
		var pt legacySamplingPoint
		b, _ := json.Marshal(m)
		if err := json.Unmarshal(b, &pt); err != nil {
			continue
		}
		name := pt.Name
		if name == "" {
			name = pt.PointName
		}
		code := pt.Code
		if code == "" {
			code = pt.PointCode
		}
		count := pt.SampleCount
		if count == 0 {
			count = pt.Count
		}
		if name == "" {
			name = fmt.Sprintf("点位-%d", scheduleID)
		}
		if err := db.Create(&model.SamplingPoint{
			SamplingScheduleID: scheduleID,
			PointName:          name,
			PointCode:          code,
			Location:           pt.Location,
			Longitude:          pt.Longitude,
			Latitude:           pt.Latitude,
			SamplingMethod:     pt.SamplingMethod,
			SampleCount:        count,
		}).Error; err != nil {
			return err
		}
	}
	return nil
}
