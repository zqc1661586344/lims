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

// ============================================================
// TaskOrderHandler — 任务委托（节点1：任务创建）
// ============================================================

type TaskOrderHandler struct {
	*GenericHandler[model.TaskOrder]
	svc *service.BusinessService
}

func NewTaskOrderHandler(logger *zap.Logger, db *gorm.DB) *TaskOrderHandler {
	return &TaskOrderHandler{
		GenericHandler: NewGenericHandler[model.TaskOrder](logger, db),
		svc:            service.NewBusinessService(logger, db),
	}
}

func (h *TaskOrderHandler) List(c *gin.Context) {
	h.GenericHandler.List(c,
		[]string{"order_no", "customer_name", "project_name"},
		nil,
	)
}

func (h *TaskOrderHandler) Get(c *gin.Context) {
	h.GenericHandler.Get(c)
}

type CreateTaskOrderRequest struct {
	OrderNo      string `json:"order_no" binding:"required"`
	CustomerName string `json:"customer_name" binding:"required"`
	ProjectName  string `json:"project_name" binding:"required"`
	SampleType   string `json:"sample_type"`
	TestItems    string `json:"test_items"`
}

// Create 创建任务委托（草稿状态）
func (h *TaskOrderHandler) Create(c *gin.Context) {
	var req CreateTaskOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数错误: %v", err))
		return
	}
	userID := middleware.GetUserID(c)
	item := model.TaskOrder{
		OrderNo:      req.OrderNo,
		CustomerName: req.CustomerName,
		ProjectName:  req.ProjectName,
		SampleType:   req.SampleType,
		TestItems:    req.TestItems,
		Status:       0, // 草稿
		CreatedBy:    &userID,
	}
	if err := h.GetDB(c).Create(&item).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("创建任务委托失败: %v", err))
		return
	}
	utils.Created(c, item)
}

type UpdateTaskOrderRequest struct {
	OrderNo      string `json:"order_no"`
	CustomerName string `json:"customer_name"`
	ProjectName  string `json:"project_name"`
	SampleType   string `json:"sample_type"`
	TestItems    string `json:"test_items"`
}

// Update 更新任务委托（草稿状态下可编辑）
func (h *TaskOrderHandler) Update(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	var item model.TaskOrder
	if err := h.GetDB(c).First(&item, id).Error; err != nil {
		utils.NotFound(c, "任务委托不存在")
		return
	}
	if item.Status != 0 {
		utils.BadRequest(c, "已提交的任务委托不可修改")
		return
	}
	var req UpdateTaskOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数错误: %v", err))
		return
	}
	updates := map[string]interface{}{}
	if req.OrderNo != "" {
		updates["order_no"] = req.OrderNo
	}
	if req.CustomerName != "" {
		updates["customer_name"] = req.CustomerName
	}
	if req.ProjectName != "" {
		updates["project_name"] = req.ProjectName
	}
	if req.SampleType != "" {
		updates["sample_type"] = req.SampleType
	}
	if req.TestItems != "" {
		updates["test_items"] = req.TestItems
	}
	if err := h.GetDB(c).Model(&item).Updates(updates).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("更新任务委托失败: %v", err))
		return
	}
	h.GetDB(c).First(&item, id)
	utils.Success(c, item)
}

// Delete 删除任务委托（草稿状态）
func (h *TaskOrderHandler) Delete(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	var item model.TaskOrder
	if err := h.GetDB(c).First(&item, id).Error; err != nil {
		utils.NotFound(c, "任务委托不存在")
		return
	}
	if item.Status != 0 {
		utils.BadRequest(c, "已提交的任务委托不可删除")
		return
	}
	if err := h.GetDB(c).Delete(&item).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("删除任务委托失败: %v", err))
		return
	}
	utils.Success(c, nil)
}

// Submit 提交任务委托 → 启动流程
func (h *TaskOrderHandler) Submit(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	var item model.TaskOrder
	if err := h.GetDB(c).First(&item, id).Error; err != nil {
		utils.NotFound(c, "任务委托不存在")
		return
	}
	if item.Status != 0 {
		utils.BadRequest(c, "任务委托已提交，不可重复提交")
		return
	}
	userID := middleware.GetUserID(c)
	title := fmt.Sprintf("任务委托: %s - %s", item.OrderNo, item.CustomerName)
	instanceID, err := h.svc.StartWorkflow("task_order", item.ID, title, userID)
	if err != nil {
		utils.InternalError(c, fmt.Sprintf("启动流程失败: %v", err))
		return
	}
	// 更新状态和流程实例ID
	if err := h.GetDB(c).Model(&item).Updates(map[string]interface{}{
		"status":              1,
		"process_instance_id": instanceID,
	}).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("更新任务委托状态失败: %v", err))
		return
	}
	utils.Success(c, gin.H{
		"id":                  item.ID,
		"process_instance_id": instanceID,
	})
}

// ============================================================
// NodeHandler — 各流程节点的通用处理（节点2~6）
// ============================================================

// NodeHandler 接收 model type 和 table name，提供通用 CRUD + 审批
type NodeHandler struct {
	svc       *service.BusinessService
	db        *gorm.DB
	modelName string // 中文名，用于错误提示
}

func NewNodeHandler(logger *zap.Logger, db *gorm.DB) *NodeHandler {
	return &NodeHandler{
		svc: service.NewBusinessService(logger, db),
		db:  db,
	}
}

func (h *NodeHandler) getDB(c *gin.Context) *gorm.DB {
	if db := middleware.GetDB(c); db != nil {
		return db
	}
	return h.db
}

// ============================================================
// ContractReviewHandler — 合同评审（节点2）
// ============================================================

type ContractReviewHandler struct {
	*GenericHandler[model.ContractReview]
	svc *service.BusinessService
}
