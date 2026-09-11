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

func NewReportPrepareHandler(logger *zap.Logger, db *gorm.DB) *ReportPrepareHandler {
	return &ReportPrepareHandler{
		svc: service.NewBusinessService(logger, db),
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
	page, pageSize, offset := utils.GetPagination(c)
	var items []model.ReportPrepare
	query := h.getDB(c).Order("id DESC")
	if taskOrderID := c.Query("task_order_id"); taskOrderID != "" {
		query = query.Where("task_order_id = ?", taskOrderID)
	}
	var total int64
	if err := query.Model(&model.ReportPrepare{}).Count(&total).Error; err != nil {
		utils.InternalError(c, "查询失败")
		return
	}
	if err := query.Limit(pageSize).Offset(offset).Find(&items).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("查询报告编制失败: %v", err))
		return
	}
	utils.SuccessPage(c, items, total, page, pageSize)
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
	if err := h.svc.CheckNodeNotAdvanced(h.getDB(c), "task_order", item.TaskOrderID, workflow.NodeReportPrepare); err != nil {
		utils.BadRequest(c, err.Error())
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
	var item model.ReportPrepare
	if err := h.getDB(c).First(&item, id).Error; err != nil {
		utils.NotFound(c, "报告编制记录不存在")
		return
	}
	if err := h.svc.CheckInstanceRunning(h.getDB(c), "task_order", item.TaskOrderID); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	if err := h.getDB(c).Delete(&item).Error; err != nil {
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
	if err := h.svc.ApproveTaskByOrder(req.TaskID, userID, middleware.GetDeptIDVal(c), req.Comment); err != nil {
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
	if err := h.svc.RejectTaskByOrder(req.TaskID, userID, middleware.GetDeptIDVal(c), req.Comment); err != nil {
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
