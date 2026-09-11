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

func NewTaskAssignHandler(logger *zap.Logger, db *gorm.DB) *TaskAssignHandler {
	return &TaskAssignHandler{
		svc: service.NewBusinessService(logger, db),
		db:  db,
	}
}

func (h *TaskAssignHandler) getDB(c *gin.Context) *gorm.DB {
	if db := middleware.GetDB(c); db != nil {
		return db
	}
	return h.db
}

func (h *TaskAssignHandler) List(c *gin.Context) {
	page, pageSize, offset := utils.GetPagination(c)
	var items []model.TaskAssign
	query := h.getDB(c).Order("id DESC")
	if taskOrderID := c.Query("task_order_id"); taskOrderID != "" {
		query = query.Where("task_order_id = ?", taskOrderID)
	}
	var total int64
	if err := query.Model(&model.TaskAssign{}).Count(&total).Error; err != nil {
		utils.InternalError(c, "查询失败")
		return
	}
	if err := query.Limit(pageSize).Offset(offset).Find(&items).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("查询任务分配失败: %v", err))
		return
	}
	utils.SuccessPage(c, items, total, page, pageSize)
}

func (h *TaskAssignHandler) Get(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	var item model.TaskAssign
	if err := h.getDB(c).First(&item, id).Error; err != nil {
		utils.NotFound(c, "任务分配记录不存在")
		return
	}
	utils.Success(c, item)
}

func (h *TaskAssignHandler) Create(c *gin.Context) {
	var req struct {
		TaskOrderID  uint   `json:"task_order_id" binding:"required"`
		AssignedTo   string `json:"assigned_to"`
		TestItemList string `json:"test_item_list"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数错误: %v", err))
		return
	}
	item := model.TaskAssign{
		TaskOrderID:  req.TaskOrderID,
		AssignedTo:   req.AssignedTo,
		TestItemList: req.TestItemList,
	}
	if err := h.getDB(c).Create(&item).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("创建任务分配失败: %v", err))
		return
	}
	utils.Created(c, item)
}

func (h *TaskAssignHandler) Update(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	var item model.TaskAssign
	if err := h.getDB(c).First(&item, id).Error; err != nil {
		utils.NotFound(c, "任务分配记录不存在")
		return
	}
	if err := h.svc.CheckNodeNotAdvanced(h.getDB(c), "task_order", item.TaskOrderID, workflow.NodeTaskAssign); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	var req struct {
		AssignedTo   string `json:"assigned_to"`
		TestItemList string `json:"test_item_list"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数错误: %v", err))
		return
	}
	updates := map[string]interface{}{}
	if req.AssignedTo != "" {
		updates["assigned_to"] = req.AssignedTo
	}
	if req.TestItemList != "" {
		updates["test_item_list"] = req.TestItemList
	}
	if err := h.getDB(c).Model(&item).Updates(updates).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("更新任务分配失败: %v", err))
		return
	}
	h.getDB(c).First(&item, id)
	utils.Success(c, item)
}

func (h *TaskAssignHandler) Delete(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	var item model.TaskAssign
	if err := h.getDB(c).First(&item, id).Error; err != nil {
		utils.NotFound(c, "任务分配记录不存在")
		return
	}
	if err := h.svc.CheckInstanceRunning(h.getDB(c), "task_order", item.TaskOrderID); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	if err := h.getDB(c).Delete(&item).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("删除任务分配失败: %v", err))
		return
	}
	utils.Success(c, nil)
}

func (h *TaskAssignHandler) Approve(c *gin.Context) {
	var req struct {
		TaskID       uint   `json:"task_id" binding:"required"`
		AssignedTo   string `json:"assigned_to"`
		TestItemList string `json:"test_item_list"`
		Comment      string `json:"comment"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数错误: %v", err))
		return
	}
	rec := model.TaskAssign{
		TaskOrderID:  req.TaskID,
		AssignedTo:   req.AssignedTo,
		TestItemList: req.TestItemList,
	}
	h.getDB(c).Where("task_order_id = ?", req.TaskID).Assign(rec).FirstOrCreate(&rec)

	userID := middleware.GetUserID(c)
	if err := h.svc.ApproveTaskByOrder(req.TaskID, userID, middleware.GetDeptIDVal(c), req.Comment); err != nil {
		utils.InternalError(c, fmt.Sprintf("审批失败: %v", err))
		return
	}
	utils.Success(c, gin.H{"message": "任务分配通过"})
}

func (h *TaskAssignHandler) Reject(c *gin.Context) {
	var req struct {
		TaskID  uint   `json:"task_id" binding:"required"`
		Comment string `json:"comment" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数错误: %v", err))
		return
	}
	rec := model.TaskAssign{
		TaskOrderID: req.TaskID,
	}
	h.getDB(c).Where("task_order_id = ?", req.TaskID).Assign(rec).FirstOrCreate(&rec)

	userID := middleware.GetUserID(c)
	if err := h.svc.RejectTaskByOrder(req.TaskID, userID, middleware.GetDeptIDVal(c), req.Comment); err != nil {
		utils.InternalError(c, fmt.Sprintf("驳回失败: %v", err))
		return
	}
	utils.Success(c, gin.H{"message": "任务分配已驳回"})
}

// ============================================================
// DataEntryHandler — 数据录入（节点8）
// 注意：DataEntry 是 1-to-many 关系（允许同一委托单多次录入），
// Approve 不创建 DataEntry 记录，直接调 svc.ApproveTask
// ============================================================

type DataEntryHandler struct {
	svc *service.BusinessService
	db  *gorm.DB
}
