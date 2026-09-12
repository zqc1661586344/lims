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
		GenericHandler: NewGenericHandler[model.TaskAssign](logger, db),
		svc:            service.NewBusinessService(logger, db),
	}
}

func (h *TaskAssignHandler) List(c *gin.Context) {
	h.GenericHandler.List(c, nil, func(db *gorm.DB, c *gin.Context) *gorm.DB {
		if id := c.Query("task_order_id"); id != "" {
			db = db.Where("task_order_id = ?", id)
		}
		return db
	})
}

func (h *TaskAssignHandler) Get(c *gin.Context) {
	h.GenericHandler.Get(c)
}

func (h *TaskAssignHandler) Create(c *gin.Context) {
	var req struct {
		TaskOrderID    uint   `json:"task_order_id" binding:"required"`
		AssignedTo     string `json:"assigned_to"`
		AssigneeUserID *uint  `json:"assignee_user_id"`
		TestItemList   string `json:"test_item_list"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数错误: %v", err))
		return
	}
	item := model.TaskAssign{
		TaskOrderID:    req.TaskOrderID,
		AssignedTo:     req.AssignedTo,
		AssigneeUserID: req.AssigneeUserID,
		TestItemList:   req.TestItemList,
	}
	if err := h.GetDB(c).Create(&item).Error; err != nil {
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
	if err := h.GetDB(c).First(&item, id).Error; err != nil {
		utils.NotFound(c, "任务分配记录不存在")
		return
	}
	if err := h.svc.CheckNodeNotAdvanced(h.GetDB(c), "task_order", item.TaskOrderID, workflow.NodeTaskAssign); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	var req struct {
		AssignedTo     string `json:"assigned_to"`
		AssigneeUserID *uint  `json:"assignee_user_id"`
		TestItemList   string `json:"test_item_list"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数错误: %v", err))
		return
	}
	updates := map[string]interface{}{}
	if req.AssignedTo != "" {
		updates["assigned_to"] = req.AssignedTo
	}
	if req.AssigneeUserID != nil {
		updates["assignee_user_id"] = *req.AssigneeUserID
	}
	if req.TestItemList != "" {
		updates["test_item_list"] = req.TestItemList
	}
	if err := h.GetDB(c).Model(&item).Updates(updates).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("更新任务分配失败: %v", err))
		return
	}
	h.GetDB(c).First(&item, id)
	utils.Success(c, item)
}

func (h *TaskAssignHandler) Delete(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	var item model.TaskAssign
	if err := h.GetDB(c).First(&item, id).Error; err != nil {
		utils.NotFound(c, "任务分配记录不存在")
		return
	}
	if err := h.svc.CheckNodeNotAdvanced(h.GetDB(c), "task_order", item.TaskOrderID, workflow.NodeTaskAssign); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	if err := h.svc.CheckInstanceRunning(h.GetDB(c), "task_order", item.TaskOrderID); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	if err := h.GetDB(c).Delete(&item).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("删除任务分配失败: %v", err))
		return
	}
	utils.Success(c, nil)
}

func (h *TaskAssignHandler) Approve(c *gin.Context) {
	var req struct {
		TaskID         uint   `json:"task_id" binding:"required"`
		AssignedTo     string `json:"assigned_to"`
		AssigneeUserID *uint  `json:"assignee_user_id"`
		TestItemList   string `json:"test_item_list"`
		Comment        string `json:"comment"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数错误: %v", err))
		return
	}

	userID := middleware.GetUserID(c)
	if err := h.svc.ApproveWithBusiness(req.TaskID, userID, middleware.GetDeptIDVal(c), req.Comment, func(tx *gorm.DB) error {
		rec := model.TaskAssign{
			TaskOrderID:    req.TaskID,
			AssignedTo:     req.AssignedTo,
			AssigneeUserID: req.AssigneeUserID,
			TestItemList:   req.TestItemList,
		}
		if err := tx.Where("task_order_id = ?", req.TaskID).Assign(rec).FirstOrCreate(&rec).Error; err != nil {
			return err
		}
		if req.AssigneeUserID != nil && *req.AssigneeUserID > 0 {
			if err := h.svc.AssignNextNodeTaskByOrderTx(tx, req.TaskID, *req.AssigneeUserID); err != nil {
				return fmt.Errorf("任务分配完成，但指派失败: %w", err)
			}
		}
		return nil
	}); err != nil {
		HandleWorkflowError(c, err, "审批失败")
		return
	}
	utils.Success(c, gin.H{"message": "任务分配通过"})
}

func (h *TaskAssignHandler) Reject(c *gin.Context) {
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
		rec := model.TaskAssign{
			TaskOrderID: req.TaskID,
		}
		return tx.Where("task_order_id = ?", req.TaskID).Assign(rec).FirstOrCreate(&rec).Error
	}, opts...); err != nil {
		HandleWorkflowError(c, err, "驳回失败")
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
	*GenericHandler[model.DataEntry]
	svc *service.BusinessService
}
