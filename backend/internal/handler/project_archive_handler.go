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

func NewProjectArchiveHandler(logger *zap.Logger, db *gorm.DB) *ProjectArchiveHandler {
	return &ProjectArchiveHandler{
		GenericHandler: NewGenericHandler[model.ProjectArchive](logger, db),
		svc:            service.NewBusinessService(logger, db),
	}
}

func (h *ProjectArchiveHandler) List(c *gin.Context) {
	h.GenericHandler.List(c, nil, func(db *gorm.DB, c *gin.Context) *gorm.DB {
		if id := c.Query("task_order_id"); id != "" {
			db = db.Where("task_order_id = ?", id)
		}
		return db
	})
}

func (h *ProjectArchiveHandler) Get(c *gin.Context) {
	h.GenericHandler.Get(c)
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
	if err := h.GetDB(c).Create(&item).Error; err != nil {
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
	if err := h.GetDB(c).First(&item, id).Error; err != nil {
		utils.NotFound(c, "项目归档记录不存在")
		return
	}
	if err := h.svc.CheckNodeNotAdvanced(h.GetDB(c), "task_order", item.TaskOrderID, workflow.NodeProjectArchive); err != nil {
		utils.BadRequest(c, err.Error())
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
	if err := h.GetDB(c).Model(&item).Updates(updates).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("更新项目归档失败: %v", err))
		return
	}
	h.GetDB(c).First(&item, id)
	utils.Success(c, item)
}

func (h *ProjectArchiveHandler) Delete(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	var item model.ProjectArchive
	if err := h.GetDB(c).First(&item, id).Error; err != nil {
		utils.NotFound(c, "项目归档记录不存在")
		return
	}
	if err := h.svc.CheckNodeNotAdvanced(h.GetDB(c), "task_order", item.TaskOrderID, workflow.NodeProjectArchive); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	if err := h.svc.CheckInstanceRunning(h.GetDB(c), "task_order", item.TaskOrderID); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	if err := h.GetDB(c).Delete(&item).Error; err != nil {
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

	userID := middleware.GetUserID(c)
	if err := h.svc.ApproveWithBusiness(req.TaskID, userID, middleware.GetDeptIDVal(c), req.Comment, func(tx *gorm.DB) error {
		archiveFiles := req.ArchiveFiles
		if archiveFiles == "" {
			archiveFiles = h.aggregateArchiveFilesWithTx(tx, req.TaskID)
		}
		retentionPeriod := req.RetentionPeriod
		if retentionPeriod == 0 {
			retentionPeriod = 36
		}
		rec := model.ProjectArchive{
			TaskOrderID:     req.TaskID,
			ArchiveNo:       req.ArchiveNo,
			ArchiveLocation: req.ArchiveLocation,
			ArchiveDate:     req.ArchiveDate,
			ArchiveFiles:    archiveFiles,
			ArchiveComment:  req.ArchiveComment,
			RetentionPeriod: retentionPeriod,
		}
		return tx.Where("task_order_id = ?", req.TaskID).Assign(rec).FirstOrCreate(&rec).Error
	}); err != nil {
		HandleWorkflowError(c, err, "审批失败")
		return
	}
	utils.Success(c, gin.H{"message": "项目归档通过"})
}

func (h *ProjectArchiveHandler) aggregateArchiveFilesWithTx(db *gorm.DB, taskOrderID uint) string {
	type docItem struct {
		Stage   string `json:"stage"`
		DocName string `json:"doc_name"`
		Ref     string `json:"ref"`
	}
	docs := make([]docItem, 0)

	var order model.TaskOrder
	if err := db.First(&order, taskOrderID).Error; err == nil {
		docs = append(docs, docItem{Stage: "D1", DocName: "委托任务单", Ref: order.OrderNo})
	}

	var contract model.ContractReview
	if err := db.Where("task_order_id = ?", taskOrderID).First(&contract).Error; err == nil && contract.ContractFilePath != "" {
		docs = append(docs, docItem{Stage: "D2", DocName: "检测合同/协议", Ref: contract.ContractFilePath})
	}

	var qc model.QCTask
	if err := db.Where("task_order_id = ?", taskOrderID).First(&qc).Error; err == nil && qc.QCDetails != "" {
		docs = append(docs, docItem{Stage: "D3", DocName: "委托检测方案(含质控)", Ref: qc.QCDetails})
	}

	var field model.FieldSamplingRecord
	if err := db.Where("task_order_id = ?", taskOrderID).First(&field).Error; err == nil {
		if field.SamplingRecordFilePath != "" {
			docs = append(docs, docItem{Stage: "D4", DocName: "现场采样记录", Ref: field.SamplingRecordFilePath})
		}
		if field.EquipmentCalRecords != "" {
			docs = append(docs, docItem{Stage: "D4", DocName: "设备校准记录", Ref: field.EquipmentCalRecords})
		}
	}

	var sample model.SampleReceiving
	if err := db.Where("task_order_id = ?", taskOrderID).First(&sample).Error; err == nil && sample.ReceivingRecordPath != "" {
		docs = append(docs, docItem{Stage: "D5", DocName: "样品接收记录", Ref: sample.ReceivingRecordPath})
	}

	var entries []model.DataEntry
	if err := db.Where("task_order_id = ?", taskOrderID).Find(&entries).Error; err == nil {
		for i := range entries {
			docs = append(docs, docItem{Stage: "D6", DocName: "实验原始记录", Ref: fmt.Sprintf("data_entry#%d", i+1)})
		}
	}

	var prepare model.ReportPrepare
	if err := db.Where("task_order_id = ?", taskOrderID).First(&prepare).Error; err == nil && prepare.ReportFile != "" {
		docs = append(docs, docItem{Stage: "D9", DocName: "报告+" + prepare.ReportTitle, Ref: prepare.ReportFile})
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
		RejectTarget   string `json:"reject_target"`
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
	if err := h.svc.RejectWithBusiness(req.TaskID, userID, middleware.GetDeptIDVal(c), req.ArchiveComment, func(tx *gorm.DB) error {
		rec := model.ProjectArchive{
			TaskOrderID:    req.TaskID,
			ArchiveComment: req.ArchiveComment,
		}
		return tx.Where("task_order_id = ?", req.TaskID).Assign(rec).FirstOrCreate(&rec).Error
	}, opts...); err != nil {
		HandleWorkflowError(c, err, "驳回失败")
		return
	}
	utils.Success(c, gin.H{"message": "项目归档已驳回"})
}
