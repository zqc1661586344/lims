package handler

import (
	"fmt"
	"lims-backend/internal/middleware"
	"lims-backend/internal/model"
	"lims-backend/internal/service"
	"lims-backend/internal/utils"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func NewReportPrintHandler(logger *zap.Logger, db *gorm.DB) *ReportPrintHandler {
	return &ReportPrintHandler{
		svc: service.NewBusinessService(logger, db),
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
	if err := h.svc.ApproveTaskByOrder(req.TaskID, userID, req.Comment); err != nil {
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
	if err := h.svc.RejectTaskByOrder(req.TaskID, userID, req.PrintComment); err != nil {
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
