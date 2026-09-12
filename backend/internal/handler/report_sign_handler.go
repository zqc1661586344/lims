package handler

import (
	"encoding/json"
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

func NewReportSignHandler(logger *zap.Logger, db *gorm.DB) *ReportSignHandler {
	return &ReportSignHandler{
		GenericHandler: NewGenericHandler[model.ReportSign](logger, db),
		svc:            service.NewBusinessService(logger, db),
	}
}

func (h *ReportSignHandler) List(c *gin.Context) {
	h.GenericHandler.List(c, nil, func(db *gorm.DB, c *gin.Context) *gorm.DB {
		if id := c.Query("task_order_id"); id != "" {
			db = db.Where("task_order_id = ?", id)
		}
		return db
	})
}

func (h *ReportSignHandler) Get(c *gin.Context) {
	h.GenericHandler.Get(c)
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
	if err := h.GetDB(c).Create(&item).Error; err != nil {
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
	if err := h.GetDB(c).First(&item, id).Error; err != nil {
		utils.NotFound(c, "报告签发记录不存在")
		return
	}
	if err := h.svc.CheckNodeNotAdvanced(h.GetDB(c), "task_order", item.TaskOrderID, workflow.NodeReportSign); err != nil {
		utils.BadRequest(c, err.Error())
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
	if err := h.GetDB(c).Model(&item).Updates(updates).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("更新报告签发失败: %v", err))
		return
	}
	h.GetDB(c).First(&item, id)
	utils.Success(c, item)
}

func (h *ReportSignHandler) Delete(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	var item model.ReportSign
	if err := h.GetDB(c).First(&item, id).Error; err != nil {
		utils.NotFound(c, "报告签发记录不存在")
		return
	}
	if err := h.svc.CheckNodeNotAdvanced(h.GetDB(c), "task_order", item.TaskOrderID, workflow.NodeReportSign); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	if err := h.svc.CheckInstanceRunning(h.GetDB(c), "task_order", item.TaskOrderID); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	if err := h.GetDB(c).Delete(&item).Error; err != nil {
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
	effectiveResult := "通过"
	if req.SignResult != "" {
		effectiveResult = req.SignResult
	}

	userID := middleware.GetUserID(c)
	if err := h.svc.ApproveWithBusiness(req.TaskID, userID, middleware.GetDeptIDVal(c), req.Comment, func(tx *gorm.DB) error {
		reportNo, reportTitle, prepareOpinion, reviewOpinion, auditOpinion, rawRecords := h.aggregateSignSlipWithTx(tx, req.TaskID)
		rec := model.ReportSign{
			TaskOrderID:    req.TaskID,
			ReportNo:       reportNo,
			ReportTitle:    reportTitle,
			PrepareOpinion: prepareOpinion,
			ReviewOpinion:  reviewOpinion,
			AuditOpinion:   auditOpinion,
			RawRecords:     rawRecords,
			SignResult:     effectiveResult,
			SignComment:    req.SignComment,
			SignerName:     req.SignerName,
			SignDate:       req.SignDate,
			SignStamp:      req.SignStamp,
		}
		return tx.Where("task_order_id = ?", req.TaskID).Assign(rec).FirstOrCreate(&rec).Error
	}); err != nil {
		HandleWorkflowError(h.Logger, c, err, "审批失败")
		return
	}
	utils.Success(c, gin.H{"message": "报告签发通过"})
}

func (h *ReportSignHandler) aggregateSignSlipWithTx(db *gorm.DB, taskOrderID uint) (reportNo, reportTitle, prepareOpinion, reviewOpinion, auditOpinion, rawRecords string) {
	var prepare model.ReportPrepare
	if err := db.Where("task_order_id = ?", taskOrderID).First(&prepare).Error; err == nil {
		reportTitle = prepare.ReportTitle
		reportNo = prepare.ReportNo
		prepareOpinion = prepare.PrepareOpinion
	}

	var review model.ReportReview
	if err := db.Where("task_order_id = ?", taskOrderID).First(&review).Error; err == nil {
		reviewOpinion = review.ReviewComment
	}

	var audit model.ReportAudit
	if err := db.Where("task_order_id = ?", taskOrderID).First(&audit).Error; err == nil {
		auditOpinion = audit.AuditComment
	}

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
		TaskID       uint   `json:"task_id" binding:"required"`
		SignComment  string `json:"sign_comment" binding:"required"`
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
	if err := h.svc.RejectWithBusiness(req.TaskID, userID, middleware.GetDeptIDVal(c), req.SignComment, func(tx *gorm.DB) error {
		rec := model.ReportSign{
			TaskOrderID: req.TaskID,
			SignResult:  "驳回",
			SignComment: req.SignComment,
		}
		return tx.Where("task_order_id = ?", req.TaskID).Assign(rec).FirstOrCreate(&rec).Error
	}, opts...); err != nil {
		HandleWorkflowError(h.Logger, c, err, "驳回失败")
		return
	}
	utils.Success(c, gin.H{"message": "报告签发已驳回"})
}

// ============================================================
// ReportPrintHandler — 报告打印发放（节点15）
// ============================================================

type ReportPrintHandler struct {
	*GenericHandler[model.ReportPrint]
	svc *service.BusinessService
}
