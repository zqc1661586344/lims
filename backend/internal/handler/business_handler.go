package handler

import (
	"encoding/json"
	"fmt"
	"lims-backend/internal/middleware"
	"lims-backend/internal/model"
	"lims-backend/internal/service"
	"lims-backend/internal/utils"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ============================================================
// TaskOrderHandler — 任务委托（节点1：任务创建）
// ============================================================

type TaskOrderHandler struct {
	svc *service.BusinessService
	db  *gorm.DB
}

func NewTaskOrderHandler(db *gorm.DB) *TaskOrderHandler {
	return &TaskOrderHandler{
		svc: service.NewBusinessService(db),
		db:  db,
	}
}

func (h *TaskOrderHandler) getDB(c *gin.Context) *gorm.DB {
	if db := middleware.GetDB(c); db != nil {
		return db
	}
	return h.db
}

// List 返回任务委托列表
func (h *TaskOrderHandler) List(c *gin.Context) {
	var items []model.TaskOrder
	query := h.getDB(c).Order("id DESC")
	if keyword := c.Query("keyword"); keyword != "" {
		query = query.Where("order_no LIKE ? OR customer_name LIKE ? OR project_name LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	if err := query.Find(&items).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("查询任务委托失败: %v", err))
		return
	}
	utils.Success(c, items)
}

// Get 获取单个任务委托
func (h *TaskOrderHandler) Get(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	var item model.TaskOrder
	if err := h.getDB(c).First(&item, id).Error; err != nil {
		utils.NotFound(c, "任务委托不存在")
		return
	}
	utils.Success(c, item)
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
	if err := h.getDB(c).Create(&item).Error; err != nil {
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
	if err := h.getDB(c).First(&item, id).Error; err != nil {
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
	if err := h.getDB(c).Model(&item).Updates(updates).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("更新任务委托失败: %v", err))
		return
	}
	h.getDB(c).First(&item, id)
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
	if err := h.getDB(c).First(&item, id).Error; err != nil {
		utils.NotFound(c, "任务委托不存在")
		return
	}
	if item.Status != 0 {
		utils.BadRequest(c, "已提交的任务委托不可删除")
		return
	}
	if err := h.getDB(c).Delete(&item).Error; err != nil {
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
	if err := h.getDB(c).First(&item, id).Error; err != nil {
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
	if err := h.getDB(c).Model(&item).Updates(map[string]interface{}{
		"status":              1,
		"process_instance_id": instanceID,
	}).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("更新任务委托状态失败: %v", err))
		return
	}
	utils.Success(c, gin.H{
		"id":                 item.ID,
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

func NewNodeHandler(db *gorm.DB) *NodeHandler {
	return &NodeHandler{
		svc: service.NewBusinessService(db),
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
	svc *service.BusinessService
	db  *gorm.DB
}

func NewContractReviewHandler(db *gorm.DB) *ContractReviewHandler {
	return &ContractReviewHandler{
		svc: service.NewBusinessService(db),
		db:  db,
	}
}

func (h *ContractReviewHandler) getDB(c *gin.Context) *gorm.DB {
	if db := middleware.GetDB(c); db != nil {
		return db
	}
	return h.db
}

// List 返回合同评审列表
func (h *ContractReviewHandler) List(c *gin.Context) {
	var items []model.ContractReview
	query := h.getDB(c).Order("id DESC")
	if taskOrderID := c.Query("task_order_id"); taskOrderID != "" {
		query = query.Where("task_order_id = ?", taskOrderID)
	}
	if err := query.Find(&items).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("查询合同评审失败: %v", err))
		return
	}
	utils.Success(c, items)
}

// Get 获取单个合同评审
func (h *ContractReviewHandler) Get(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	var item model.ContractReview
	if err := h.getDB(c).First(&item, id).Error; err != nil {
		utils.NotFound(c, "合同评审记录不存在")
		return
	}
	utils.Success(c, item)
}

type SaveContractReviewRequest struct {
	TaskOrderID      uint   `json:"task_order_id" binding:"required"`
	ReviewResult     string `json:"review_result"`
	ReviewComment    string `json:"review_comment"`
	ContractFilePath string `json:"contract_file_path"`
}

// Create 创建合同评审
func (h *ContractReviewHandler) Create(c *gin.Context) {
	var req SaveContractReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数错误: %v", err))
		return
	}
	item := model.ContractReview{
		TaskOrderID:      req.TaskOrderID,
		ReviewResult:     req.ReviewResult,
		ReviewComment:    req.ReviewComment,
		ContractFilePath: req.ContractFilePath,
	}
	if err := h.getDB(c).Create(&item).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("创建合同评审失败: %v", err))
		return
	}
	utils.Created(c, item)
}

// Update 更新合同评审
func (h *ContractReviewHandler) Update(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	var item model.ContractReview
	if err := h.getDB(c).First(&item, id).Error; err != nil {
		utils.NotFound(c, "合同评审记录不存在")
		return
	}
	var req SaveContractReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数错误: %v", err))
		return
	}
	updates := map[string]interface{}{}
	if req.ReviewResult != "" {
		updates["review_result"] = req.ReviewResult
	}
	if req.ReviewComment != "" {
		updates["review_comment"] = req.ReviewComment
	}
	if req.ContractFilePath != "" {
		updates["contract_file_path"] = req.ContractFilePath
	}
	if err := h.getDB(c).Model(&item).Updates(updates).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("更新合同评审失败: %v", err))
		return
	}
	h.getDB(c).First(&item, id)
	utils.Success(c, item)
}

// Delete 删除合同评审
func (h *ContractReviewHandler) Delete(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	if err := h.getDB(c).Delete(&model.ContractReview{}, id).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("删除合同评审失败: %v", err))
		return
	}
	utils.Success(c, nil)
}

// Approve 通过合同评审并推进流程
func (h *ContractReviewHandler) Approve(c *gin.Context) {
	var req struct {
		TaskID           uint   `json:"task_id" binding:"required"`
		ReviewResult     string `json:"review_result"`
		ReviewComment    string `json:"review_comment"`
		ContractFilePath string `json:"contract_file_path"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数错误: %v", err))
		return
	}
	// 保存或更新合同评审记录
	review := model.ContractReview{
		TaskOrderID:      req.TaskID, // 使用 task_id 关联的业务ID
		ReviewResult:     "通过",
		ReviewComment:    req.ReviewComment,
		ContractFilePath: req.ContractFilePath,
	}
	if req.ReviewResult != "" {
		review.ReviewResult = req.ReviewResult
	}
	h.getDB(c).Where("task_order_id = ?", req.TaskID).Assign(review).FirstOrCreate(&review)

	userID := middleware.GetUserID(c)
	if err := h.svc.ApproveTask(req.TaskID, userID, req.ReviewComment); err != nil {
		utils.InternalError(c, fmt.Sprintf("审批失败: %v", err))
		return
	}
	utils.Success(c, gin.H{"message": "合同评审通过"})
}

// Reject 驳回合同评审
func (h *ContractReviewHandler) Reject(c *gin.Context) {
	var req struct {
		TaskID        uint   `json:"task_id" binding:"required"`
		ReviewComment string `json:"review_comment" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数错误: %v", err))
		return
	}
	// 保存驳回记录
	review := model.ContractReview{
		TaskOrderID:   req.TaskID,
		ReviewResult:  "驳回",
		ReviewComment: req.ReviewComment,
	}
	h.getDB(c).Where("task_order_id = ?", req.TaskID).Assign(review).FirstOrCreate(&review)

	userID := middleware.GetUserID(c)
	if err := h.svc.RejectTask(req.TaskID, userID, req.ReviewComment); err != nil {
		utils.InternalError(c, fmt.Sprintf("驳回失败: %v", err))
		return
	}
	utils.Success(c, gin.H{"message": "合同评审已驳回"})
}

// ============================================================
// QCTaskHandler — 质控任务（节点3）
// ============================================================

type QCTaskHandler struct {
	svc *service.BusinessService
	db  *gorm.DB
}

func NewQCTaskHandler(db *gorm.DB) *QCTaskHandler {
	return &QCTaskHandler{
		svc: service.NewBusinessService(db),
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
		TaskID     uint   `json:"task_id" binding:"required"`
		QCType     string `json:"qc_type"`
		QCDetails  string `json:"qc_details"`
		Comment    string `json:"comment"`
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
	if err := h.svc.ApproveTask(req.TaskID, userID, req.Comment); err != nil {
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
	if err := h.svc.RejectTask(req.TaskID, userID, req.Comment); err != nil {
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

func NewSamplingScheduleHandler(db *gorm.DB) *SamplingScheduleHandler {
	return &SamplingScheduleHandler{
		svc: service.NewBusinessService(db),
		db:  db,
	}
}

func (h *SamplingScheduleHandler) getDB(c *gin.Context) *gorm.DB {
	if db := middleware.GetDB(c); db != nil {
		return db
	}
	return h.db
}

func (h *SamplingScheduleHandler) List(c *gin.Context) {
	var items []model.SamplingSchedule
	query := h.getDB(c).Order("id DESC")
	if taskOrderID := c.Query("task_order_id"); taskOrderID != "" {
		query = query.Where("task_order_id = ?", taskOrderID)
	}
	if err := query.Find(&items).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("查询采样调度失败: %v", err))
		return
	}
	utils.Success(c, items)
}

func (h *SamplingScheduleHandler) Get(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	var item model.SamplingSchedule
	if err := h.getDB(c).First(&item, id).Error; err != nil {
		utils.NotFound(c, "采样调度记录不存在")
		return
	}
	utils.Success(c, item)
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
	if err := h.getDB(c).Create(&item).Error; err != nil {
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
	if err := h.getDB(c).First(&item, id).Error; err != nil {
		utils.NotFound(c, "采样调度记录不存在")
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
	if err := h.getDB(c).Model(&item).Updates(updates).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("更新采样调度失败: %v", err))
		return
	}
	h.getDB(c).First(&item, id)
	utils.Success(c, item)
}

func (h *SamplingScheduleHandler) Delete(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	if err := h.getDB(c).Delete(&model.SamplingSchedule{}, id).Error; err != nil {
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
	h.getDB(c).Where("task_order_id = ?", req.TaskID).Assign(sched).FirstOrCreate(&sched)

	userID := middleware.GetUserID(c)
	if err := h.svc.ApproveTask(req.TaskID, userID, req.Comment); err != nil {
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
	h.getDB(c).Where("task_order_id = ?", req.TaskID).Assign(sched).FirstOrCreate(&sched)

	userID := middleware.GetUserID(c)
	if err := h.svc.RejectTask(req.TaskID, userID, req.Comment); err != nil {
		utils.InternalError(c, fmt.Sprintf("驳回失败: %v", err))
		return
	}
	utils.Success(c, gin.H{"message": "采样调度已驳回"})
}

// ============================================================
// FieldSamplingRecordHandler — 现场采样（节点5）
// ============================================================

type FieldSamplingRecordHandler struct {
	svc *service.BusinessService
	db  *gorm.DB
}

func NewFieldSamplingRecordHandler(db *gorm.DB) *FieldSamplingRecordHandler {
	return &FieldSamplingRecordHandler{
		svc: service.NewBusinessService(db),
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
	if err := h.getDB(c).Delete(&model.FieldSamplingRecord{}, id).Error; err != nil {
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
	if err := h.svc.ApproveTask(req.TaskID, userID, req.Comment); err != nil {
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
	if err := h.svc.RejectTask(req.TaskID, userID, req.Comment); err != nil {
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

func NewSampleReceivingHandler(db *gorm.DB) *SampleReceivingHandler {
	return &SampleReceivingHandler{
		svc: service.NewBusinessService(db),
		db:  db,
	}
}

func (h *SampleReceivingHandler) getDB(c *gin.Context) *gorm.DB {
	if db := middleware.GetDB(c); db != nil {
		return db
	}
	return h.db
}

func (h *SampleReceivingHandler) List(c *gin.Context) {
	var items []model.SampleReceiving
	query := h.getDB(c).Order("id DESC")
	if taskOrderID := c.Query("task_order_id"); taskOrderID != "" {
		query = query.Where("task_order_id = ?", taskOrderID)
	}
	if err := query.Find(&items).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("查询样品接收记录失败: %v", err))
		return
	}
	utils.Success(c, items)
}

func (h *SampleReceivingHandler) Get(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	var item model.SampleReceiving
	if err := h.getDB(c).First(&item, id).Error; err != nil {
		utils.NotFound(c, "样品接收记录不存在")
		return
	}
	utils.Success(c, item)
}

func (h *SampleReceivingHandler) Create(c *gin.Context) {
	var req struct {
		TaskOrderID        uint   `json:"task_order_id" binding:"required"`
		SampleCondition    string `json:"sample_condition"`
		SampleCodes        string `json:"sample_codes"`
		ReceivingRecordPath string `json:"receiving_record_path"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数错误: %v", err))
		return
	}
	item := model.SampleReceiving{
		TaskOrderID:        req.TaskOrderID,
		SampleCondition:    req.SampleCondition,
		SampleCodes:        req.SampleCodes,
		ReceivingRecordPath: req.ReceivingRecordPath,
	}
	if err := h.getDB(c).Create(&item).Error; err != nil {
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
	if err := h.getDB(c).First(&item, id).Error; err != nil {
		utils.NotFound(c, "样品接收记录不存在")
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
	if err := h.getDB(c).Model(&item).Updates(updates).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("更新样品接收记录失败: %v", err))
		return
	}
	h.getDB(c).First(&item, id)
	utils.Success(c, item)
}

func (h *SampleReceivingHandler) Delete(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	if err := h.getDB(c).Delete(&model.SampleReceiving{}, id).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("删除样品接收记录失败: %v", err))
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
	rec := model.SampleReceiving{
		TaskOrderID:        req.TaskID,
		SampleCondition:    req.SampleCondition,
		SampleCodes:        req.SampleCodes,
		ReceivingRecordPath: req.ReceivingRecordPath,
	}
	h.getDB(c).Where("task_order_id = ?", req.TaskID).Assign(rec).FirstOrCreate(&rec)

	userID := middleware.GetUserID(c)
	if err := h.svc.ApproveTask(req.TaskID, userID, req.Comment); err != nil {
		utils.InternalError(c, fmt.Sprintf("审批失败: %v", err))
		return
	}
	utils.Success(c, gin.H{"message": "样品接收通过"})
}

func (h *SampleReceivingHandler) Reject(c *gin.Context) {
	var req struct {
		TaskID  uint   `json:"task_id" binding:"required"`
		Comment string `json:"comment" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数错误: %v", err))
		return
	}
	rec := model.SampleReceiving{
		TaskOrderID: req.TaskID,
	}
	h.getDB(c).Where("task_order_id = ?", req.TaskID).Assign(rec).FirstOrCreate(&rec)

	userID := middleware.GetUserID(c)
	if err := h.svc.RejectTask(req.TaskID, userID, req.Comment); err != nil {
		utils.InternalError(c, fmt.Sprintf("驳回失败: %v", err))
		return
	}
	utils.Success(c, gin.H{"message": "样品接收已驳回"})
}

// ============================================================
// TaskAssignHandler — 任务分配（节点7）
// ============================================================

type TaskAssignHandler struct {
	svc *service.BusinessService
	db  *gorm.DB
}

func NewTaskAssignHandler(db *gorm.DB) *TaskAssignHandler {
	return &TaskAssignHandler{
		svc: service.NewBusinessService(db),
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
	var items []model.TaskAssign
	query := h.getDB(c).Order("id DESC")
	if taskOrderID := c.Query("task_order_id"); taskOrderID != "" {
		query = query.Where("task_order_id = ?", taskOrderID)
	}
	if err := query.Find(&items).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("查询任务分配失败: %v", err))
		return
	}
	utils.Success(c, items)
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
	if err := h.getDB(c).Delete(&model.TaskAssign{}, id).Error; err != nil {
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
	if err := h.svc.ApproveTask(req.TaskID, userID, req.Comment); err != nil {
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
	if err := h.svc.RejectTask(req.TaskID, userID, req.Comment); err != nil {
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

func NewDataEntryHandler(db *gorm.DB) *DataEntryHandler {
	return &DataEntryHandler{
		svc: service.NewBusinessService(db),
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
	if err := h.svc.ApproveTask(req.TaskID, userID, req.Comment); err != nil {
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
	if err := h.svc.RejectTask(req.TaskID, userID, req.Comment); err != nil {
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

func NewDataReviewHandler(db *gorm.DB) *DataReviewHandler {
	return &DataReviewHandler{
		svc: service.NewBusinessService(db),
		db:  db,
	}
}

func (h *DataReviewHandler) getDB(c *gin.Context) *gorm.DB {
	if db := middleware.GetDB(c); db != nil {
		return db
	}
	return h.db
}

func (h *DataReviewHandler) List(c *gin.Context) {
	var items []model.DataReview
	query := h.getDB(c).Order("id DESC")
	if taskOrderID := c.Query("task_order_id"); taskOrderID != "" {
		query = query.Where("task_order_id = ?", taskOrderID)
	}
	if err := query.Find(&items).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("查询数据复核失败: %v", err))
		return
	}
	utils.Success(c, items)
}

func (h *DataReviewHandler) Get(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	var item model.DataReview
	if err := h.getDB(c).First(&item, id).Error; err != nil {
		utils.NotFound(c, "数据复核记录不存在")
		return
	}
	utils.Success(c, item)
}

func (h *DataReviewHandler) Create(c *gin.Context) {
	var req struct {
		TaskOrderID   uint   `json:"task_order_id" binding:"required"`
		ReviewResult  string `json:"review_result"`
		ReviewComment string `json:"review_comment"`
		IssuesFound   string `json:"issues_found"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数错误: %v", err))
		return
	}
	item := model.DataReview{
		TaskOrderID:   req.TaskOrderID,
		ReviewResult:  req.ReviewResult,
		ReviewComment: req.ReviewComment,
		IssuesFound:   req.IssuesFound,
	}
	if err := h.getDB(c).Create(&item).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("创建数据复核失败: %v", err))
		return
	}
	utils.Created(c, item)
}

func (h *DataReviewHandler) Update(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	var item model.DataReview
	if err := h.getDB(c).First(&item, id).Error; err != nil {
		utils.NotFound(c, "数据复核记录不存在")
		return
	}
	var req struct {
		ReviewResult  string `json:"review_result"`
		ReviewComment string `json:"review_comment"`
		IssuesFound   string `json:"issues_found"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数错误: %v", err))
		return
	}
	updates := map[string]interface{}{}
	if req.ReviewResult != "" {
		updates["review_result"] = req.ReviewResult
	}
	if req.ReviewComment != "" {
		updates["review_comment"] = req.ReviewComment
	}
	if req.IssuesFound != "" {
		updates["issues_found"] = req.IssuesFound
	}
	if err := h.getDB(c).Model(&item).Updates(updates).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("更新数据复核失败: %v", err))
		return
	}
	h.getDB(c).First(&item, id)
	utils.Success(c, item)
}

func (h *DataReviewHandler) Delete(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	if err := h.getDB(c).Delete(&model.DataReview{}, id).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("删除数据复核失败: %v", err))
		return
	}
	utils.Success(c, nil)
}

func (h *DataReviewHandler) Approve(c *gin.Context) {
	var req struct {
		TaskID        uint   `json:"task_id" binding:"required"`
		ReviewResult  string `json:"review_result"`
		ReviewComment string `json:"review_comment"`
		IssuesFound   string `json:"issues_found"`
		Comment       string `json:"comment"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数错误: %v", err))
		return
	}
	review := model.DataReview{
		TaskOrderID:   req.TaskID,
		ReviewResult:  "通过",
		ReviewComment: req.ReviewComment,
		IssuesFound:   req.IssuesFound,
	}
	if req.ReviewResult != "" {
		review.ReviewResult = req.ReviewResult
	}
	h.getDB(c).Where("task_order_id = ?", req.TaskID).Assign(review).FirstOrCreate(&review)

	userID := middleware.GetUserID(c)
	if err := h.svc.ApproveTask(req.TaskID, userID, req.Comment); err != nil {
		utils.InternalError(c, fmt.Sprintf("审批失败: %v", err))
		return
	}
	utils.Success(c, gin.H{"message": "数据复核通过"})
}

func (h *DataReviewHandler) Reject(c *gin.Context) {
	var req struct {
		TaskID        uint   `json:"task_id" binding:"required"`
		ReviewComment string `json:"review_comment" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数错误: %v", err))
		return
	}
	review := model.DataReview{
		TaskOrderID:   req.TaskID,
		ReviewResult:  "驳回",
		ReviewComment: req.ReviewComment,
	}
	h.getDB(c).Where("task_order_id = ?", req.TaskID).Assign(review).FirstOrCreate(&review)

	userID := middleware.GetUserID(c)
	if err := h.svc.RejectTask(req.TaskID, userID, req.ReviewComment); err != nil {
		utils.InternalError(c, fmt.Sprintf("驳回失败: %v", err))
		return
	}
	utils.Success(c, gin.H{"message": "数据复核已驳回"})
}

// ============================================================
// DataAuditHandler — 数据审核（节点10）
// ============================================================

type DataAuditHandler struct {
	svc *service.BusinessService
	db  *gorm.DB
}

func NewDataAuditHandler(db *gorm.DB) *DataAuditHandler {
	return &DataAuditHandler{
		svc: service.NewBusinessService(db),
		db:  db,
	}
}

func (h *DataAuditHandler) getDB(c *gin.Context) *gorm.DB {
	if db := middleware.GetDB(c); db != nil {
		return db
	}
	return h.db
}

func (h *DataAuditHandler) List(c *gin.Context) {
	var items []model.DataAudit
	query := h.getDB(c).Order("id DESC")
	if taskOrderID := c.Query("task_order_id"); taskOrderID != "" {
		query = query.Where("task_order_id = ?", taskOrderID)
	}
	if err := query.Find(&items).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("查询数据审核失败: %v", err))
		return
	}
	utils.Success(c, items)
}

func (h *DataAuditHandler) Get(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	var item model.DataAudit
	if err := h.getDB(c).First(&item, id).Error; err != nil {
		utils.NotFound(c, "数据审核记录不存在")
		return
	}
	utils.Success(c, item)
}

func (h *DataAuditHandler) Create(c *gin.Context) {
	var req struct {
		TaskOrderID  uint   `json:"task_order_id" binding:"required"`
		AuditResult  string `json:"audit_result"`
		AuditComment string `json:"audit_comment"`
		IssueList    string `json:"issue_list"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数错误: %v", err))
		return
	}
	item := model.DataAudit{
		TaskOrderID:  req.TaskOrderID,
		AuditResult:  req.AuditResult,
		AuditComment: req.AuditComment,
		IssueList:    req.IssueList,
	}
	if err := h.getDB(c).Create(&item).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("创建数据审核失败: %v", err))
		return
	}
	utils.Created(c, item)
}

func (h *DataAuditHandler) Update(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	var item model.DataAudit
	if err := h.getDB(c).First(&item, id).Error; err != nil {
		utils.NotFound(c, "数据审核记录不存在")
		return
	}
	var req struct {
		AuditResult  string `json:"audit_result"`
		AuditComment string `json:"audit_comment"`
		IssueList    string `json:"issue_list"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数错误: %v", err))
		return
	}
	updates := map[string]interface{}{}
	if req.AuditResult != "" {
		updates["audit_result"] = req.AuditResult
	}
	if req.AuditComment != "" {
		updates["audit_comment"] = req.AuditComment
	}
	if req.IssueList != "" {
		updates["issue_list"] = req.IssueList
	}
	if err := h.getDB(c).Model(&item).Updates(updates).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("更新数据审核失败: %v", err))
		return
	}
	h.getDB(c).First(&item, id)
	utils.Success(c, item)
}

func (h *DataAuditHandler) Delete(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	if err := h.getDB(c).Delete(&model.DataAudit{}, id).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("删除数据审核失败: %v", err))
		return
	}
	utils.Success(c, nil)
}

func (h *DataAuditHandler) Approve(c *gin.Context) {
	var req struct {
		TaskID       uint   `json:"task_id" binding:"required"`
		AuditResult  string `json:"audit_result"`
		AuditComment string `json:"audit_comment"`
		IssueList    string `json:"issue_list"`
		Comment      string `json:"comment"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数错误: %v", err))
		return
	}
	audit := model.DataAudit{
		TaskOrderID:  req.TaskID,
		AuditResult:  "通过",
		AuditComment: req.AuditComment,
		IssueList:    req.IssueList,
	}
	if req.AuditResult != "" {
		audit.AuditResult = req.AuditResult
	}
	h.getDB(c).Where("task_order_id = ?", req.TaskID).Assign(audit).FirstOrCreate(&audit)

	userID := middleware.GetUserID(c)
	if err := h.svc.ApproveTask(req.TaskID, userID, req.Comment); err != nil {
		utils.InternalError(c, fmt.Sprintf("审批失败: %v", err))
		return
	}
	utils.Success(c, gin.H{"message": "数据审核通过"})
}

func (h *DataAuditHandler) Reject(c *gin.Context) {
	var req struct {
		TaskID       uint   `json:"task_id" binding:"required"`
		AuditComment string `json:"audit_comment" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数错误: %v", err))
		return
	}
	audit := model.DataAudit{
		TaskOrderID:  req.TaskID,
		AuditResult:  "驳回",
		AuditComment: req.AuditComment,
	}
	h.getDB(c).Where("task_order_id = ?", req.TaskID).Assign(audit).FirstOrCreate(&audit)

	userID := middleware.GetUserID(c)
	if err := h.svc.RejectTask(req.TaskID, userID, req.AuditComment); err != nil {
		utils.InternalError(c, fmt.Sprintf("驳回失败: %v", err))
		return
	}
	utils.Success(c, gin.H{"message": "数据审核已驳回"})
}

// ============================================================
// ReportPrepareHandler — 报告编制（节点11）
// ============================================================

type ReportPrepareHandler struct {
	svc *service.BusinessService
	db  *gorm.DB
}

func NewReportPrepareHandler(db *gorm.DB) *ReportPrepareHandler {
	return &ReportPrepareHandler{
		svc: service.NewBusinessService(db),
		db:  db,
	}
}

func (h *ReportPrepareHandler) getDB(c *gin.Context) *gorm.DB {
	if db := middleware.GetDB(c); db != nil {
		return db
	}
	return h.db
}

func (h *ReportPrepareHandler) List(c *gin.Context) {
	var items []model.ReportPrepare
	query := h.getDB(c).Order("id DESC")
	if taskOrderID := c.Query("task_order_id"); taskOrderID != "" {
		query = query.Where("task_order_id = ?", taskOrderID)
	}
	if err := query.Find(&items).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("查询报告编制失败: %v", err))
		return
	}
	utils.Success(c, items)
}

func (h *ReportPrepareHandler) Get(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	var item model.ReportPrepare
	if err := h.getDB(c).First(&item, id).Error; err != nil {
		utils.NotFound(c, "报告编制记录不存在")
		return
	}
	utils.Success(c, item)
}

func (h *ReportPrepareHandler) Create(c *gin.Context) {
	var req struct {
		TaskOrderID   uint   `json:"task_order_id" binding:"required"`
		ReportTitle   string `json:"report_title"`
		ReportContent string `json:"report_content"`
		ReportFile    string `json:"report_file"`
		Attachments   string `json:"attachments"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数错误: %v", err))
		return
	}
	item := model.ReportPrepare{
		TaskOrderID:   req.TaskOrderID,
		ReportTitle:   req.ReportTitle,
		ReportContent: req.ReportContent,
		ReportFile:    req.ReportFile,
		Attachments:   req.Attachments,
	}
	if err := h.getDB(c).Create(&item).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("创建报告编制失败: %v", err))
		return
	}
	utils.Created(c, item)
}

func (h *ReportPrepareHandler) Update(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	var item model.ReportPrepare
	if err := h.getDB(c).First(&item, id).Error; err != nil {
		utils.NotFound(c, "报告编制记录不存在")
		return
	}
	var req struct {
		ReportTitle   string `json:"report_title"`
		ReportContent string `json:"report_content"`
		ReportFile    string `json:"report_file"`
		Attachments   string `json:"attachments"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数错误: %v", err))
		return
	}
	updates := map[string]interface{}{}
	if req.ReportTitle != "" {
		updates["report_title"] = req.ReportTitle
	}
	if req.ReportContent != "" {
		updates["report_content"] = req.ReportContent
	}
	if req.ReportFile != "" {
		updates["report_file"] = req.ReportFile
	}
	if req.Attachments != "" {
		updates["attachments"] = req.Attachments
	}
	if err := h.getDB(c).Model(&item).Updates(updates).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("更新报告编制失败: %v", err))
		return
	}
	h.getDB(c).First(&item, id)
	utils.Success(c, item)
}

func (h *ReportPrepareHandler) Delete(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	if err := h.getDB(c).Delete(&model.ReportPrepare{}, id).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("删除报告编制失败: %v", err))
		return
	}
	utils.Success(c, nil)
}

func (h *ReportPrepareHandler) Approve(c *gin.Context) {
	var req struct {
		TaskID         uint   `json:"task_id" binding:"required"`
		ReportNo       string `json:"report_no"`
		PrepareOpinion string `json:"prepare_opinion"`
		ReportTitle    string `json:"report_title"`
		ReportContent  string `json:"report_content"`
		ReportFile     string `json:"report_file"`
		Attachments    string `json:"attachments"`
		Comment        string `json:"comment"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数错误: %v", err))
		return
	}
	rec := model.ReportPrepare{
		TaskOrderID:    req.TaskID,
		ReportNo:       req.ReportNo,
		PrepareOpinion: req.PrepareOpinion,
		ReportTitle:    req.ReportTitle,
		ReportContent:  req.ReportContent,
		ReportFile:     req.ReportFile,
		Attachments:    req.Attachments,
	}
	h.getDB(c).Where("task_order_id = ?", req.TaskID).Assign(rec).FirstOrCreate(&rec)

	userID := middleware.GetUserID(c)
	if err := h.svc.ApproveTask(req.TaskID, userID, req.Comment); err != nil {
		utils.InternalError(c, fmt.Sprintf("审批失败: %v", err))
		return
	}
	utils.Success(c, gin.H{"message": "报告编制通过"})
}

func (h *ReportPrepareHandler) Reject(c *gin.Context) {
	var req struct {
		TaskID  uint   `json:"task_id" binding:"required"`
		Comment string `json:"comment" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数错误: %v", err))
		return
	}
	rec := model.ReportPrepare{
		TaskOrderID: req.TaskID,
	}
	h.getDB(c).Where("task_order_id = ?", req.TaskID).Assign(rec).FirstOrCreate(&rec)

	userID := middleware.GetUserID(c)
	if err := h.svc.RejectTask(req.TaskID, userID, req.Comment); err != nil {
		utils.InternalError(c, fmt.Sprintf("驳回失败: %v", err))
		return
	}
	utils.Success(c, gin.H{"message": "报告编制已驳回"})
}

// ============================================================
// ReportReviewHandler — 报告复核（节点12）
// ============================================================

type ReportReviewHandler struct {
	svc *service.BusinessService
	db  *gorm.DB
}

func NewReportReviewHandler(db *gorm.DB) *ReportReviewHandler {
	return &ReportReviewHandler{
		svc: service.NewBusinessService(db),
		db:  db,
	}
}

func (h *ReportReviewHandler) getDB(c *gin.Context) *gorm.DB {
	if db := middleware.GetDB(c); db != nil {
		return db
	}
	return h.db
}

func (h *ReportReviewHandler) List(c *gin.Context) {
	var items []model.ReportReview
	query := h.getDB(c).Order("id DESC")
	if taskOrderID := c.Query("task_order_id"); taskOrderID != "" {
		query = query.Where("task_order_id = ?", taskOrderID)
	}
	if err := query.Find(&items).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("查询报告复核失败: %v", err))
		return
	}
	utils.Success(c, items)
}

func (h *ReportReviewHandler) Get(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	var item model.ReportReview
	if err := h.getDB(c).First(&item, id).Error; err != nil {
		utils.NotFound(c, "报告复核记录不存在")
		return
	}
	utils.Success(c, item)
}

func (h *ReportReviewHandler) Create(c *gin.Context) {
	var req struct {
		TaskOrderID   uint   `json:"task_order_id" binding:"required"`
		ReviewResult  string `json:"review_result"`
		ReviewComment string `json:"review_comment"`
		ReviewedItems string `json:"reviewed_items"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数错误: %v", err))
		return
	}
	item := model.ReportReview{
		TaskOrderID:   req.TaskOrderID,
		ReviewResult:  req.ReviewResult,
		ReviewComment: req.ReviewComment,
		ReviewedItems: req.ReviewedItems,
	}
	if err := h.getDB(c).Create(&item).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("创建报告复核失败: %v", err))
		return
	}
	utils.Created(c, item)
}

func (h *ReportReviewHandler) Update(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	var item model.ReportReview
	if err := h.getDB(c).First(&item, id).Error; err != nil {
		utils.NotFound(c, "报告复核记录不存在")
		return
	}
	var req struct {
		ReviewResult  string `json:"review_result"`
		ReviewComment string `json:"review_comment"`
		ReviewedItems string `json:"reviewed_items"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数错误: %v", err))
		return
	}
	updates := map[string]interface{}{}
	if req.ReviewResult != "" {
		updates["review_result"] = req.ReviewResult
	}
	if req.ReviewComment != "" {
		updates["review_comment"] = req.ReviewComment
	}
	if req.ReviewedItems != "" {
		updates["reviewed_items"] = req.ReviewedItems
	}
	if err := h.getDB(c).Model(&item).Updates(updates).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("更新报告复核失败: %v", err))
		return
	}
	h.getDB(c).First(&item, id)
	utils.Success(c, item)
}

func (h *ReportReviewHandler) Delete(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	if err := h.getDB(c).Delete(&model.ReportReview{}, id).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("删除报告复核失败: %v", err))
		return
	}
	utils.Success(c, nil)
}

func (h *ReportReviewHandler) Approve(c *gin.Context) {
	var req struct {
		TaskID        uint   `json:"task_id" binding:"required"`
		ReviewResult  string `json:"review_result"`
		ReviewComment string `json:"review_comment"`
		ReviewedItems string `json:"reviewed_items"`
		Comment       string `json:"comment"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数错误: %v", err))
		return
	}
	rec := model.ReportReview{
		TaskOrderID:   req.TaskID,
		ReviewResult:  "通过",
		ReviewComment: req.ReviewComment,
		ReviewedItems: req.ReviewedItems,
	}
	if req.ReviewResult != "" {
		rec.ReviewResult = req.ReviewResult
	}
	h.getDB(c).Where("task_order_id = ?", req.TaskID).Assign(rec).FirstOrCreate(&rec)

	userID := middleware.GetUserID(c)
	if err := h.svc.ApproveTask(req.TaskID, userID, req.Comment); err != nil {
		utils.InternalError(c, fmt.Sprintf("审批失败: %v", err))
		return
	}
	utils.Success(c, gin.H{"message": "报告复核通过"})
}

func (h *ReportReviewHandler) Reject(c *gin.Context) {
	var req struct {
		TaskID        uint   `json:"task_id" binding:"required"`
		ReviewComment string `json:"review_comment" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数错误: %v", err))
		return
	}
	rec := model.ReportReview{
		TaskOrderID:   req.TaskID,
		ReviewResult:  "驳回",
		ReviewComment: req.ReviewComment,
	}
	h.getDB(c).Where("task_order_id = ?", req.TaskID).Assign(rec).FirstOrCreate(&rec)

	userID := middleware.GetUserID(c)
	if err := h.svc.RejectTask(req.TaskID, userID, req.ReviewComment); err != nil {
		utils.InternalError(c, fmt.Sprintf("驳回失败: %v", err))
		return
	}
	utils.Success(c, gin.H{"message": "报告复核已驳回"})
}

// ============================================================
// ReportAuditHandler — 报告审核（节点13）
// ============================================================

type ReportAuditHandler struct {
	svc *service.BusinessService
	db  *gorm.DB
}

func NewReportAuditHandler(db *gorm.DB) *ReportAuditHandler {
	return &ReportAuditHandler{
		svc: service.NewBusinessService(db),
		db:  db,
	}
}

func (h *ReportAuditHandler) getDB(c *gin.Context) *gorm.DB {
	if db := middleware.GetDB(c); db != nil {
		return db
	}
	return h.db
}

func (h *ReportAuditHandler) List(c *gin.Context) {
	var items []model.ReportAudit
	query := h.getDB(c).Order("id DESC")
	if taskOrderID := c.Query("task_order_id"); taskOrderID != "" {
		query = query.Where("task_order_id = ?", taskOrderID)
	}
	if err := query.Find(&items).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("查询报告审核失败: %v", err))
		return
	}
	utils.Success(c, items)
}

func (h *ReportAuditHandler) Get(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	var item model.ReportAudit
	if err := h.getDB(c).First(&item, id).Error; err != nil {
		utils.NotFound(c, "报告审核记录不存在")
		return
	}
	utils.Success(c, item)
}

func (h *ReportAuditHandler) Create(c *gin.Context) {
	var req struct {
		TaskOrderID  uint   `json:"task_order_id" binding:"required"`
		AuditResult  string `json:"audit_result"`
		AuditComment string `json:"audit_comment"`
		AuditIssues  string `json:"audit_issues"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数错误: %v", err))
		return
	}
	item := model.ReportAudit{
		TaskOrderID:  req.TaskOrderID,
		AuditResult:  req.AuditResult,
		AuditComment: req.AuditComment,
		AuditIssues:  req.AuditIssues,
	}
	if err := h.getDB(c).Create(&item).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("创建报告审核失败: %v", err))
		return
	}
	utils.Created(c, item)
}

func (h *ReportAuditHandler) Update(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	var item model.ReportAudit
	if err := h.getDB(c).First(&item, id).Error; err != nil {
		utils.NotFound(c, "报告审核记录不存在")
		return
	}
	var req struct {
		AuditResult  string `json:"audit_result"`
		AuditComment string `json:"audit_comment"`
		AuditIssues  string `json:"audit_issues"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数错误: %v", err))
		return
	}
	updates := map[string]interface{}{}
	if req.AuditResult != "" {
		updates["audit_result"] = req.AuditResult
	}
	if req.AuditComment != "" {
		updates["audit_comment"] = req.AuditComment
	}
	if req.AuditIssues != "" {
		updates["audit_issues"] = req.AuditIssues
	}
	if err := h.getDB(c).Model(&item).Updates(updates).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("更新报告审核失败: %v", err))
		return
	}
	h.getDB(c).First(&item, id)
	utils.Success(c, item)
}

func (h *ReportAuditHandler) Delete(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	if err := h.getDB(c).Delete(&model.ReportAudit{}, id).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("删除报告审核失败: %v", err))
		return
	}
	utils.Success(c, nil)
}

func (h *ReportAuditHandler) Approve(c *gin.Context) {
	var req struct {
		TaskID       uint   `json:"task_id" binding:"required"`
		AuditResult  string `json:"audit_result"`
		AuditComment string `json:"audit_comment"`
		AuditIssues  string `json:"audit_issues"`
		Comment      string `json:"comment"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数错误: %v", err))
		return
	}
	rec := model.ReportAudit{
		TaskOrderID:  req.TaskID,
		AuditResult:  "通过",
		AuditComment: req.AuditComment,
		AuditIssues:  req.AuditIssues,
	}
	if req.AuditResult != "" {
		rec.AuditResult = req.AuditResult
	}
	h.getDB(c).Where("task_order_id = ?", req.TaskID).Assign(rec).FirstOrCreate(&rec)

	userID := middleware.GetUserID(c)
	if err := h.svc.ApproveTask(req.TaskID, userID, req.Comment); err != nil {
		utils.InternalError(c, fmt.Sprintf("审批失败: %v", err))
		return
	}
	utils.Success(c, gin.H{"message": "报告审核通过"})
}

func (h *ReportAuditHandler) Reject(c *gin.Context) {
	var req struct {
		TaskID       uint   `json:"task_id" binding:"required"`
		AuditComment string `json:"audit_comment" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数错误: %v", err))
		return
	}
	rec := model.ReportAudit{
		TaskOrderID:  req.TaskID,
		AuditResult:  "驳回",
		AuditComment: req.AuditComment,
	}
	h.getDB(c).Where("task_order_id = ?", req.TaskID).Assign(rec).FirstOrCreate(&rec)

	userID := middleware.GetUserID(c)
	if err := h.svc.RejectTask(req.TaskID, userID, req.AuditComment); err != nil {
		utils.InternalError(c, fmt.Sprintf("驳回失败: %v", err))
		return
	}
	utils.Success(c, gin.H{"message": "报告审核已驳回"})
}

// ============================================================
// ReportSignHandler — 报告签发（节点14）
// ============================================================

type ReportSignHandler struct {
	svc *service.BusinessService
	db  *gorm.DB
}

func NewReportSignHandler(db *gorm.DB) *ReportSignHandler {
	return &ReportSignHandler{
		svc: service.NewBusinessService(db),
		db:  db,
	}
}

func (h *ReportSignHandler) getDB(c *gin.Context) *gorm.DB {
	if db := middleware.GetDB(c); db != nil {
		return db
	}
	return h.db
}

func (h *ReportSignHandler) List(c *gin.Context) {
	var items []model.ReportSign
	query := h.getDB(c).Order("id DESC")
	if taskOrderID := c.Query("task_order_id"); taskOrderID != "" {
		query = query.Where("task_order_id = ?", taskOrderID)
	}
	if err := query.Find(&items).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("查询报告签发失败: %v", err))
		return
	}
	utils.Success(c, items)
}

func (h *ReportSignHandler) Get(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	var item model.ReportSign
	if err := h.getDB(c).First(&item, id).Error; err != nil {
		utils.NotFound(c, "报告签发记录不存在")
		return
	}
	utils.Success(c, item)
}

func (h *ReportSignHandler) Create(c *gin.Context) {
	var req struct {
		TaskOrderID uint       `json:"task_order_id" binding:"required"`
		SignResult  string     `json:"sign_result"`
		SignComment string     `json:"sign_comment"`
		SignerName  string     `json:"signer_name"`
		SignDate    *time.Time `json:"sign_date"`
		SignStamp   string     `json:"sign_stamp"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数错误: %v", err))
		return
	}
	item := model.ReportSign{
		TaskOrderID: req.TaskOrderID,
		SignResult:  req.SignResult,
		SignComment: req.SignComment,
		SignerName:  req.SignerName,
		SignDate:    req.SignDate,
		SignStamp:   req.SignStamp,
	}
	if err := h.getDB(c).Create(&item).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("创建报告签发失败: %v", err))
		return
	}
	utils.Created(c, item)
}

func (h *ReportSignHandler) Update(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	var item model.ReportSign
	if err := h.getDB(c).First(&item, id).Error; err != nil {
		utils.NotFound(c, "报告签发记录不存在")
		return
	}
	var req struct {
		SignResult  string     `json:"sign_result"`
		SignComment string     `json:"sign_comment"`
		SignerName  string     `json:"signer_name"`
		SignDate    *time.Time `json:"sign_date"`
		SignStamp   string     `json:"sign_stamp"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数错误: %v", err))
		return
	}
	updates := map[string]interface{}{}
	if req.SignResult != "" {
		updates["sign_result"] = req.SignResult
	}
	if req.SignComment != "" {
		updates["sign_comment"] = req.SignComment
	}
	if req.SignerName != "" {
		updates["signer_name"] = req.SignerName
	}
	if req.SignDate != nil {
		updates["sign_date"] = req.SignDate
	}
	if req.SignStamp != "" {
		updates["sign_stamp"] = req.SignStamp
	}
	if err := h.getDB(c).Model(&item).Updates(updates).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("更新报告签发失败: %v", err))
		return
	}
	h.getDB(c).First(&item, id)
	utils.Success(c, item)
}

func (h *ReportSignHandler) Delete(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	if err := h.getDB(c).Delete(&model.ReportSign{}, id).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("删除报告签发失败: %v", err))
		return
	}
	utils.Success(c, nil)
}

func (h *ReportSignHandler) Approve(c *gin.Context) {
	var req struct {
		TaskID      uint       `json:"task_id" binding:"required"`
		SignResult  string     `json:"sign_result"`
		SignComment string     `json:"sign_comment"`
		SignerName  string     `json:"signer_name"`
		SignDate    *time.Time `json:"sign_date"`
		SignStamp   string     `json:"sign_stamp"`
		Comment     string     `json:"comment"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数错误: %v", err))
		return
	}

	// 自动汇聚"报告审核签发单"（流程图 D15）：承接报告编制标题、编制/复核/审核意见及实验原始记录
	reportNo, reportTitle, prepareOpinion, reviewOpinion, auditOpinion, rawRecords := h.aggregateSignSlip(c, req.TaskID)

	rec := model.ReportSign{
		TaskOrderID:    req.TaskID,
		ReportNo:       reportNo,
		ReportTitle:    reportTitle,
		PrepareOpinion: prepareOpinion,
		ReviewOpinion:  reviewOpinion,
		AuditOpinion:   auditOpinion,
		RawRecords:     rawRecords,
		SignResult:     "通过",
		SignComment:    req.SignComment,
		SignerName:     req.SignerName,
		SignDate:       req.SignDate,
		SignStamp:      req.SignStamp,
	}
	if req.SignResult != "" {
		rec.SignResult = req.SignResult
	}
	h.getDB(c).Where("task_order_id = ?", req.TaskID).Assign(rec).FirstOrCreate(&rec)

	userID := middleware.GetUserID(c)
	if err := h.svc.ApproveTask(req.TaskID, userID, req.Comment); err != nil {
		utils.InternalError(c, fmt.Sprintf("审批失败: %v", err))
		return
	}
	utils.Success(c, gin.H{"message": "报告签发通过"})
}

// aggregateSignSlip 汇聚生成"报告审核签发单"（流程图 D15）所需数据：
// 报告编号/标题（取自报告编制）、编制/复核/审核各环节意见、以及实验原始记录（data_entries）。
func (h *ReportSignHandler) aggregateSignSlip(c *gin.Context, taskOrderID uint) (reportNo, reportTitle, prepareOpinion, reviewOpinion, auditOpinion, rawRecords string) {
	db := h.getDB(c)

	// 报告编制（D9）——报告编号、标题与编制意见
	var prepare model.ReportPrepare
	if err := db.Where("task_order_id = ?", taskOrderID).First(&prepare).Error; err == nil {
		reportTitle = prepare.ReportTitle
		reportNo = prepare.ReportNo // 报告编号（报告编制阶段赋号）
		prepareOpinion = prepare.PrepareOpinion
	}

	// 报告复核（D10）——复核意见
	var review model.ReportReview
	if err := db.Where("task_order_id = ?", taskOrderID).First(&review).Error; err == nil {
		reviewOpinion = review.ReviewComment
	}

	// 报告审核（D11）——审核意见
	var audit model.ReportAudit
	if err := db.Where("task_order_id = ?", taskOrderID).First(&audit).Error; err == nil {
		auditOpinion = audit.AuditComment
	}

	// 实验原始记录（D9/D12 中的"实验原始记录"部分，来自数据录入 data_entries）
	var entries []model.DataEntry
	if err := db.Where("task_order_id = ?", taskOrderID).Find(&entries).Error; err == nil && len(entries) > 0 {
		type rawRec struct {
			TestItemID   uint   `json:"test_item_id"`
			OriginalData string `json:"original_data"`
		}
		list := make([]rawRec, 0, len(entries))
		for _, e := range entries {
			list = append(list, rawRec{TestItemID: e.TestItemID, OriginalData: e.OriginalData})
		}
		if b, err := json.Marshal(list); err == nil {
			rawRecords = string(b)
		}
	}

	return reportNo, reportTitle, prepareOpinion, reviewOpinion, auditOpinion, rawRecords
}

func (h *ReportSignHandler) Reject(c *gin.Context) {
	var req struct {
		TaskID      uint   `json:"task_id" binding:"required"`
		SignComment string `json:"sign_comment" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数错误: %v", err))
		return
	}
	rec := model.ReportSign{
		TaskOrderID: req.TaskID,
		SignResult:  "驳回",
		SignComment: req.SignComment,
	}
	h.getDB(c).Where("task_order_id = ?", req.TaskID).Assign(rec).FirstOrCreate(&rec)

	userID := middleware.GetUserID(c)
	if err := h.svc.RejectTask(req.TaskID, userID, req.SignComment); err != nil {
		utils.InternalError(c, fmt.Sprintf("驳回失败: %v", err))
		return
	}
	utils.Success(c, gin.H{"message": "报告签发已驳回"})
}

// ============================================================
// ReportPrintHandler — 报告打印发放（节点15）
// ============================================================

type ReportPrintHandler struct {
	svc *service.BusinessService
	db  *gorm.DB
}

func NewReportPrintHandler(db *gorm.DB) *ReportPrintHandler {
	return &ReportPrintHandler{
		svc: service.NewBusinessService(db),
		db:  db,
	}
}

func (h *ReportPrintHandler) getDB(c *gin.Context) *gorm.DB {
	if db := middleware.GetDB(c); db != nil {
		return db
	}
	return h.db
}

func (h *ReportPrintHandler) List(c *gin.Context) {
	var items []model.ReportPrint
	query := h.getDB(c).Order("id DESC")
	if taskOrderID := c.Query("task_order_id"); taskOrderID != "" {
		query = query.Where("task_order_id = ?", taskOrderID)
	}
	if err := query.Find(&items).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("查询报告打印发放失败: %v", err))
		return
	}
	utils.Success(c, items)
}

func (h *ReportPrintHandler) Get(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	var item model.ReportPrint
	if err := h.getDB(c).First(&item, id).Error; err != nil {
		utils.NotFound(c, "报告打印发放记录不存在")
		return
	}
	utils.Success(c, item)
}

func (h *ReportPrintHandler) Create(c *gin.Context) {
	var req struct {
		TaskOrderID    uint       `json:"task_order_id" binding:"required"`
		PrintCount     int        `json:"print_count"`
		PrintResult    string     `json:"print_result"`
		PrintComment   string     `json:"print_comment"`
		RecipientName  string     `json:"recipient_name"`
		RecipientDate  *time.Time `json:"recipient_date"`
		DeliveryMethod string     `json:"delivery_method"`
		TrackingNo     string     `json:"tracking_no"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数错误: %v", err))
		return
	}
	item := model.ReportPrint{
		TaskOrderID:    req.TaskOrderID,
		PrintCount:     req.PrintCount,
		PrintResult:    req.PrintResult,
		PrintComment:   req.PrintComment,
		RecipientName:  req.RecipientName,
		RecipientDate:  req.RecipientDate,
		DeliveryMethod: req.DeliveryMethod,
		TrackingNo:     req.TrackingNo,
	}
	if err := h.getDB(c).Create(&item).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("创建报告打印发放失败: %v", err))
		return
	}
	utils.Created(c, item)
}

func (h *ReportPrintHandler) Update(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	var item model.ReportPrint
	if err := h.getDB(c).First(&item, id).Error; err != nil {
		utils.NotFound(c, "报告打印发放记录不存在")
		return
	}
	var req struct {
		PrintCount     int        `json:"print_count"`
		PrintResult    string     `json:"print_result"`
		PrintComment   string     `json:"print_comment"`
		RecipientName  string     `json:"recipient_name"`
		RecipientDate  *time.Time `json:"recipient_date"`
		DeliveryMethod string     `json:"delivery_method"`
		TrackingNo     string     `json:"tracking_no"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数错误: %v", err))
		return
	}
	updates := map[string]interface{}{}
	if req.PrintCount > 0 {
		updates["print_count"] = req.PrintCount
	}
	if req.PrintResult != "" {
		updates["print_result"] = req.PrintResult
	}
	if req.PrintComment != "" {
		updates["print_comment"] = req.PrintComment
	}
	if req.RecipientName != "" {
		updates["recipient_name"] = req.RecipientName
	}
	if req.RecipientDate != nil {
		updates["recipient_date"] = req.RecipientDate
	}
	if req.DeliveryMethod != "" {
		updates["delivery_method"] = req.DeliveryMethod
	}
	if req.TrackingNo != "" {
		updates["tracking_no"] = req.TrackingNo
	}
	if err := h.getDB(c).Model(&item).Updates(updates).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("更新报告打印发放失败: %v", err))
		return
	}
	h.getDB(c).First(&item, id)
	utils.Success(c, item)
}

func (h *ReportPrintHandler) Delete(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	if err := h.getDB(c).Delete(&model.ReportPrint{}, id).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("删除报告打印发放失败: %v", err))
		return
	}
	utils.Success(c, nil)
}

func (h *ReportPrintHandler) Approve(c *gin.Context) {
	var req struct {
		TaskID         uint       `json:"task_id" binding:"required"`
		PrintCount     int        `json:"print_count"`
		PrintResult    string     `json:"print_result"`
		PrintComment   string     `json:"print_comment"`
		RecipientName  string     `json:"recipient_name"`
		RecipientDate  *time.Time `json:"recipient_date"`
		DeliveryMethod string     `json:"delivery_method"`
		TrackingNo     string     `json:"tracking_no"`
		Comment        string     `json:"comment"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数错误: %v", err))
		return
	}
	rec := model.ReportPrint{
		TaskOrderID:    req.TaskID,
		PrintCount:     req.PrintCount,
		PrintResult:    "通过",
		PrintComment:   req.PrintComment,
		RecipientName:  req.RecipientName,
		RecipientDate:  req.RecipientDate,
		DeliveryMethod: req.DeliveryMethod,
		TrackingNo:     req.TrackingNo,
	}
	if req.PrintResult != "" {
		rec.PrintResult = req.PrintResult
	}
	if rec.PrintCount == 0 {
		rec.PrintCount = 1
	}
	h.getDB(c).Where("task_order_id = ?", req.TaskID).Assign(rec).FirstOrCreate(&rec)

	userID := middleware.GetUserID(c)
	if err := h.svc.ApproveTask(req.TaskID, userID, req.Comment); err != nil {
		utils.InternalError(c, fmt.Sprintf("审批失败: %v", err))
		return
	}
	utils.Success(c, gin.H{"message": "报告打印发放通过"})
}

func (h *ReportPrintHandler) Reject(c *gin.Context) {
	var req struct {
		TaskID       uint   `json:"task_id" binding:"required"`
		PrintComment string `json:"print_comment" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数错误: %v", err))
		return
	}
	rec := model.ReportPrint{
		TaskOrderID:  req.TaskID,
		PrintResult:  "驳回",
		PrintComment: req.PrintComment,
	}
	h.getDB(c).Where("task_order_id = ?", req.TaskID).Assign(rec).FirstOrCreate(&rec)

	userID := middleware.GetUserID(c)
	if err := h.svc.RejectTask(req.TaskID, userID, req.PrintComment); err != nil {
		utils.InternalError(c, fmt.Sprintf("驳回失败: %v", err))
		return
	}
	utils.Success(c, gin.H{"message": "报告打印发放已驳回"})
}

// ============================================================
// ProjectArchiveHandler — 项目归档（节点16，终节点）
// ============================================================

type ProjectArchiveHandler struct {
	svc *service.BusinessService
	db  *gorm.DB
}

func NewProjectArchiveHandler(db *gorm.DB) *ProjectArchiveHandler {
	return &ProjectArchiveHandler{
		svc: service.NewBusinessService(db),
		db:  db,
	}
}

func (h *ProjectArchiveHandler) getDB(c *gin.Context) *gorm.DB {
	if db := middleware.GetDB(c); db != nil {
		return db
	}
	return h.db
}

func (h *ProjectArchiveHandler) List(c *gin.Context) {
	var items []model.ProjectArchive
	query := h.getDB(c).Order("id DESC")
	if taskOrderID := c.Query("task_order_id"); taskOrderID != "" {
		query = query.Where("task_order_id = ?", taskOrderID)
	}
	if err := query.Find(&items).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("查询项目归档失败: %v", err))
		return
	}
	utils.Success(c, items)
}

func (h *ProjectArchiveHandler) Get(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	var item model.ProjectArchive
	if err := h.getDB(c).First(&item, id).Error; err != nil {
		utils.NotFound(c, "项目归档记录不存在")
		return
	}
	utils.Success(c, item)
}

func (h *ProjectArchiveHandler) Create(c *gin.Context) {
	var req struct {
		TaskOrderID     uint       `json:"task_order_id" binding:"required"`
		ArchiveNo       string     `json:"archive_no"`
		ArchiveLocation string     `json:"archive_location"`
		ArchiveDate     *time.Time `json:"archive_date"`
		ArchiveFiles    string     `json:"archive_files"`
		ArchiveComment  string     `json:"archive_comment"`
		RetentionPeriod int        `json:"retention_period"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数错误: %v", err))
		return
	}
	item := model.ProjectArchive{
		TaskOrderID:     req.TaskOrderID,
		ArchiveNo:       req.ArchiveNo,
		ArchiveLocation: req.ArchiveLocation,
		ArchiveDate:     req.ArchiveDate,
		ArchiveFiles:    req.ArchiveFiles,
		ArchiveComment:  req.ArchiveComment,
		RetentionPeriod: req.RetentionPeriod,
	}
	if err := h.getDB(c).Create(&item).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("创建项目归档失败: %v", err))
		return
	}
	utils.Created(c, item)
}

func (h *ProjectArchiveHandler) Update(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	var item model.ProjectArchive
	if err := h.getDB(c).First(&item, id).Error; err != nil {
		utils.NotFound(c, "项目归档记录不存在")
		return
	}
	var req struct {
		ArchiveNo       string     `json:"archive_no"`
		ArchiveLocation string     `json:"archive_location"`
		ArchiveDate     *time.Time `json:"archive_date"`
		ArchiveFiles    string     `json:"archive_files"`
		ArchiveComment  string     `json:"archive_comment"`
		RetentionPeriod int        `json:"retention_period"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数错误: %v", err))
		return
	}
	updates := map[string]interface{}{}
	if req.ArchiveNo != "" {
		updates["archive_no"] = req.ArchiveNo
	}
	if req.ArchiveLocation != "" {
		updates["archive_location"] = req.ArchiveLocation
	}
	if req.ArchiveDate != nil {
		updates["archive_date"] = req.ArchiveDate
	}
	if req.ArchiveFiles != "" {
		updates["archive_files"] = req.ArchiveFiles
	}
	if req.ArchiveComment != "" {
		updates["archive_comment"] = req.ArchiveComment
	}
	if req.RetentionPeriod > 0 {
		updates["retention_period"] = req.RetentionPeriod
	}
	if err := h.getDB(c).Model(&item).Updates(updates).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("更新项目归档失败: %v", err))
		return
	}
	h.getDB(c).First(&item, id)
	utils.Success(c, item)
}

func (h *ProjectArchiveHandler) Delete(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	if err := h.getDB(c).Delete(&model.ProjectArchive{}, id).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("删除项目归档失败: %v", err))
		return
	}
	utils.Success(c, nil)
}

func (h *ProjectArchiveHandler) Approve(c *gin.Context) {
	var req struct {
		TaskID          uint       `json:"task_id" binding:"required"`
		ArchiveNo       string     `json:"archive_no"`
		ArchiveLocation string     `json:"archive_location"`
		ArchiveDate     *time.Time `json:"archive_date"`
		ArchiveFiles    string     `json:"archive_files"`
		ArchiveComment  string     `json:"archive_comment"`
		RetentionPeriod int        `json:"retention_period"`
		Comment         string     `json:"comment"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数错误: %v", err))
		return
	}
	// 若未手动指定归档文件清单，则自动汇聚 D1–D13 全部环节文档（流程图 D14 = 以上所有文档）
	archiveFiles := req.ArchiveFiles
	if archiveFiles == "" {
		archiveFiles = h.aggregateArchiveFiles(c, req.TaskID)
	}
	rec := model.ProjectArchive{
		TaskOrderID:     req.TaskID,
		ArchiveNo:       req.ArchiveNo,
		ArchiveLocation: req.ArchiveLocation,
		ArchiveDate:     req.ArchiveDate,
		ArchiveFiles:    archiveFiles,
		ArchiveComment:  req.ArchiveComment,
		RetentionPeriod: req.RetentionPeriod,
	}
	if rec.RetentionPeriod == 0 {
		rec.RetentionPeriod = 36
	}
	h.getDB(c).Where("task_order_id = ?", req.TaskID).Assign(rec).FirstOrCreate(&rec)

	userID := middleware.GetUserID(c)
	if err := h.svc.ApproveTask(req.TaskID, userID, req.Comment); err != nil {
		utils.InternalError(c, fmt.Sprintf("审批失败: %v", err))
		return
	}
	utils.Success(c, gin.H{"message": "项目归档通过"})
}

// aggregateArchiveFiles 自动汇聚该委托单 D1–D13 全部环节文档，生成归档文件清单（流程图 D14 = 以上所有文档）。
// 返回 JSON 数组字符串：[{stage, doc_name, ref}]
func (h *ProjectArchiveHandler) aggregateArchiveFiles(c *gin.Context, taskOrderID uint) string {
	db := h.getDB(c)
	type docItem struct {
		Stage   string `json:"stage"`
		DocName string `json:"doc_name"`
		Ref     string `json:"ref"`
	}
	docs := make([]docItem, 0)

	// D1 委托任务单
	var order model.TaskOrder
	if err := db.First(&order, taskOrderID).Error; err == nil {
		docs = append(docs, docItem{Stage: "D1", DocName: "委托任务单", Ref: order.OrderNo})
	}

	// D2 检测合同/协议
	var contract model.ContractReview
	if err := db.Where("task_order_id = ?", taskOrderID).First(&contract).Error; err == nil && contract.ContractFilePath != "" {
		docs = append(docs, docItem{Stage: "D2", DocName: "检测合同/协议", Ref: contract.ContractFilePath})
	}

	// D3 委托检测方案（质控任务）
	var qc model.QCTask
	if err := db.Where("task_order_id = ?", taskOrderID).First(&qc).Error; err == nil && qc.QCDetails != "" {
		docs = append(docs, docItem{Stage: "D3", DocName: "委托检测方案(含质控)", Ref: qc.QCDetails})
	}

	// D4 现场采样记录及设备校准记录
	var field model.FieldSamplingRecord
	if err := db.Where("task_order_id = ?", taskOrderID).First(&field).Error; err == nil {
		if field.SamplingRecordFilePath != "" {
			docs = append(docs, docItem{Stage: "D4", DocName: "现场采样记录", Ref: field.SamplingRecordFilePath})
		}
		if field.EquipmentCalRecords != "" {
			docs = append(docs, docItem{Stage: "D4", DocName: "设备校准记录", Ref: field.EquipmentCalRecords})
		}
	}

	// D5 样品接收记录
	var sample model.SampleReceiving
	if err := db.Where("task_order_id = ?", taskOrderID).First(&sample).Error; err == nil && sample.ReceivingRecordPath != "" {
		docs = append(docs, docItem{Stage: "D5", DocName: "样品接收记录", Ref: sample.ReceivingRecordPath})
	}

	// D6/D7/D8 实验原始记录（数据录入）
	var entries []model.DataEntry
	if err := db.Where("task_order_id = ?", taskOrderID).Find(&entries).Error; err == nil {
		for i := range entries {
			docs = append(docs, docItem{Stage: "D6", DocName: "实验原始记录", Ref: fmt.Sprintf("data_entry#%d", i+1)})
		}
	}

	// D9–D12 报告及审核签发单（报告编制/复核/审核/签发）
	var prepare model.ReportPrepare
	if err := db.Where("task_order_id = ?", taskOrderID).First(&prepare).Error; err == nil && prepare.ReportFile != "" {
		docs = append(docs, docItem{Stage: "D9", DocName: "报告+"+prepare.ReportTitle, Ref: prepare.ReportFile})
		docs = append(docs, docItem{Stage: "D9", DocName: "实验原始记录", Ref: "随报告"})
	}
	var rReview model.ReportReview
	if err := db.Where("task_order_id = ?", taskOrderID).First(&rReview).Error; err == nil {
		docs = append(docs, docItem{Stage: "D10", DocName: "报告复核记录", Ref: "已复核"})
	}
	var rAudit model.ReportAudit
	if err := db.Where("task_order_id = ?", taskOrderID).First(&rAudit).Error; err == nil {
		docs = append(docs, docItem{Stage: "D11", DocName: "报告审核记录", Ref: "已审核"})
	}
	var sign model.ReportSign
	if err := db.Where("task_order_id = ?", taskOrderID).First(&sign).Error; err == nil {
		docs = append(docs, docItem{Stage: "D12/D15", DocName: "报告审核签发单", Ref: sign.SignStamp})
	}

	// D13 报告发放记录
	var print model.ReportPrint
	if err := db.Where("task_order_id = ?", taskOrderID).First(&print).Error; err == nil {
		docs = append(docs, docItem{Stage: "D13", DocName: "报告发放记录", Ref: fmt.Sprintf("份数:%d 领取:%s", print.PrintCount, print.RecipientName)})
	}

	if len(docs) == 0 {
		return ""
	}
	if b, err := json.Marshal(docs); err == nil {
		return string(b)
	}
	return ""
}

func (h *ProjectArchiveHandler) Reject(c *gin.Context) {
	var req struct {
		TaskID         uint   `json:"task_id" binding:"required"`
		ArchiveComment string `json:"archive_comment" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数错误: %v", err))
		return
	}
	rec := model.ProjectArchive{
		TaskOrderID:    req.TaskID,
		ArchiveComment: req.ArchiveComment,
	}
	h.getDB(c).Where("task_order_id = ?", req.TaskID).Assign(rec).FirstOrCreate(&rec)

	userID := middleware.GetUserID(c)
	if err := h.svc.RejectTask(req.TaskID, userID, req.ArchiveComment); err != nil {
		utils.InternalError(c, fmt.Sprintf("驳回失败: %v", err))
		return
	}
	utils.Success(c, gin.H{"message": "项目归档已驳回"})
}