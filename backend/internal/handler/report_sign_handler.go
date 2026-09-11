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
		svc: service.NewBusinessService(logger, db),
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
	if err := h.svc.CheckNodeNotAdvanced(h.getDB(c), "task_order", item.TaskOrderID, workflow.NodeReportSign); err != nil {
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
	var item model.ReportSign
	if err := h.getDB(c).First(&item, id).Error; err != nil {
		utils.NotFound(c, "报告签发记录不存在")
		return
	}
	if err := h.svc.CheckInstanceRunning(h.getDB(c), "task_order", item.TaskOrderID); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	if err := h.getDB(c).Delete(&item).Error; err != nil {
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
	if err := h.svc.ApproveTaskByOrder(req.TaskID, userID, middleware.GetDeptIDVal(c), req.Comment); err != nil {
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
	if err := h.svc.RejectTaskByOrder(req.TaskID, userID, middleware.GetDeptIDVal(c), req.SignComment); err != nil {
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
