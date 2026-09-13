package handler

import (
	"fmt"
	"time"

	"lims-backend/internal/middleware"
	"lims-backend/internal/model"
	"lims-backend/internal/service"
	"lims-backend/internal/utils"
	"lims-backend/internal/workflow"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func NewReportPrintHandler(logger *zap.Logger, db *gorm.DB) *ReportPrintHandler {
	return &ReportPrintHandler{
		GenericHandler: NewGenericHandler[model.ReportPrint](logger, db),
		svc:            service.NewBusinessService(logger, db),
	}
}

func (h *ReportPrintHandler) List(c *gin.Context) {
	h.GenericHandler.List(c, nil, func(db *gorm.DB, c *gin.Context) *gorm.DB {
		db = service.ApplyTaskOrderScope(db, c)
		if id := c.Query("task_order_id"); id != "" {
			db = db.Where("task_order_id = ?", id)
		}
		return db
	})
}

func (h *ReportPrintHandler) Get(c *gin.Context) {
	h.GenericHandler.GetWithScope(c, service.ApplyTaskOrderScope)
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
	if err := h.GetDB(c).Create(&item).Error; err != nil {
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
	if err := h.GetDB(c).First(&item, id).Error; err != nil {
		utils.NotFound(c, "报告打印发放记录不存在")
		return
	}
	if err := h.svc.CheckNodeNotAdvanced(h.GetDB(c), "task_order", item.TaskOrderID, workflow.NodeReportPrint); err != nil {
		utils.BadRequest(c, err.Error())
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
	if err := h.GetDB(c).Model(&item).Updates(updates).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("更新报告打印发放失败: %v", err))
		return
	}
	h.GetDB(c).First(&item, id)
	utils.Success(c, item)
}

func (h *ReportPrintHandler) Delete(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	var item model.ReportPrint
	if err := h.GetDB(c).First(&item, id).Error; err != nil {
		utils.NotFound(c, "报告打印发放记录不存在")
		return
	}
	if err := h.svc.CheckNodeNotAdvanced(h.GetDB(c), "task_order", item.TaskOrderID, workflow.NodeReportPrint); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	if err := h.svc.CheckInstanceRunning(h.GetDB(c), "task_order", item.TaskOrderID); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	if err := h.GetDB(c).Delete(&item).Error; err != nil {
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
	effectiveResult := "通过"
	if req.PrintResult != "" {
		effectiveResult = req.PrintResult
	}
	printCount := req.PrintCount
	if printCount == 0 {
		printCount = 1
	}

	userID := middleware.GetUserID(c)
	if err := h.svc.ApproveWithBusiness(req.TaskID, userID, middleware.GetDeptIDVal(c), req.Comment, func(tx *gorm.DB) error {
		rec := model.ReportPrint{
			TaskOrderID:    req.TaskID,
			PrintCount:     printCount,
			PrintResult:    effectiveResult,
			PrintComment:   req.PrintComment,
			RecipientName:  req.RecipientName,
			RecipientDate:  req.RecipientDate,
			DeliveryMethod: req.DeliveryMethod,
			TrackingNo:     req.TrackingNo,
		}
		return tx.Where("task_order_id = ?", req.TaskID).Assign(rec).FirstOrCreate(&rec).Error
	}); err != nil {
		HandleWorkflowError(h.Logger, c, err, "审批失败")
		return
	}
	utils.Success(c, gin.H{"message": "报告打印发放通过"})
}

func (h *ReportPrintHandler) Reject(c *gin.Context) {
	var req struct {
		TaskID       uint   `json:"task_id" binding:"required"`
		PrintComment string `json:"print_comment" binding:"required"`
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
	if err := h.svc.RejectWithBusiness(req.TaskID, userID, middleware.GetDeptIDVal(c), req.PrintComment, func(tx *gorm.DB) error {
		rec := model.ReportPrint{
			TaskOrderID:  req.TaskID,
			PrintResult:  "驳回",
			PrintComment: req.PrintComment,
		}
		return tx.Where("task_order_id = ?", req.TaskID).Assign(rec).FirstOrCreate(&rec).Error
	}, opts...); err != nil {
		HandleWorkflowError(h.Logger, c, err, "驳回失败")
		return
	}
	utils.Success(c, gin.H{"message": "报告打印发放已驳回"})
}

// ============================================================
// ProjectArchiveHandler — 项目归档（节点16，终节点）
// ============================================================

type ProjectArchiveHandler struct {
	*GenericHandler[model.ProjectArchive]
	svc *service.BusinessService
}
