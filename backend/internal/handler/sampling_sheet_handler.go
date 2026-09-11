package handler

import (
	"encoding/json"
	"fmt"
	"lims-backend/internal/middleware"
	"lims-backend/internal/model"
	"lims-backend/internal/utils"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type SamplingSheetHandler struct {
	logger *zap.Logger
	db     *gorm.DB
}

func NewSamplingSheetHandler(logger *zap.Logger, db *gorm.DB) *SamplingSheetHandler {
	return &SamplingSheetHandler{logger: logger, db: db}
}

func (h *SamplingSheetHandler) getDB(c *gin.Context) *gorm.DB {
	if db := middleware.GetDB(c); db != nil {
		return db
	}
	return h.db
}

// ─── SamplingSheetTemplate CRUD ──────────────────────────────────

func (h *SamplingSheetHandler) ListTemplates(c *gin.Context) {
	var items []model.SamplingSheetTemplate
	query := h.getDB(c).Order("version DESC, id DESC")
	if name := c.Query("name"); name != "" {
		query = query.Where("name ILIKE ?", "%"+name+"%")
	}
	if sampleType := c.Query("sample_type"); sampleType != "" {
		query = query.Where("sample_type = ?", sampleType)
	}
	if nodeCode := c.Query("node_code"); nodeCode != "" {
		query = query.Where("node_code = ?", nodeCode)
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if err := query.Find(&items).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("查询采样单模板失败: %v", err))
		return
	}
	utils.Success(c, items)
}

func (h *SamplingSheetHandler) GetTemplate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	var item model.SamplingSheetTemplate
	if err := h.getDB(c).First(&item, id).Error; err != nil {
		utils.NotFound(c, "采样单模板不存在")
		return
	}
	utils.Success(c, item)
}

func (h *SamplingSheetHandler) CreateTemplate(c *gin.Context) {
	var req struct {
		Code           string          `json:"code" binding:"required"`
		Name           string          `json:"name" binding:"required"`
		SampleType     string          `json:"sample_type"`
		NodeCode       string          `json:"node_code"`
		Structure      json.RawMessage `json:"structure" binding:"required"`
		EditableRanges json.RawMessage `json:"editable_ranges"`
		ReadOnlyRanges json.RawMessage `json:"readonly_ranges"`
		Version        int             `json:"version"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数错误: %v", err))
		return
	}
	tpl := model.SamplingSheetTemplate{
		Code:           req.Code,
		Name:           req.Name,
		SampleType:     req.SampleType,
		NodeCode:       req.NodeCode,
		Structure:      req.Structure,
		EditableRanges: req.EditableRanges,
		ReadOnlyRanges: req.ReadOnlyRanges,
		Version:        req.Version,
		Status:         1,
	}
	if len(tpl.EditableRanges) == 0 {
		tpl.EditableRanges = json.RawMessage("[]")
	}
	if len(tpl.ReadOnlyRanges) == 0 {
		tpl.ReadOnlyRanges = json.RawMessage("[]")
	}
	if err := h.getDB(c).Create(&tpl).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("创建采样单模板失败: %v", err))
		return
	}
	utils.Success(c, tpl)
}

func (h *SamplingSheetHandler) UpdateTemplate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	var tpl model.SamplingSheetTemplate
	if err := h.getDB(c).First(&tpl, id).Error; err != nil {
		utils.NotFound(c, "采样单模板不存在")
		return
	}
	var req struct {
		Code           string          `json:"code"`
		Name           string          `json:"name"`
		SampleType     string          `json:"sample_type"`
		NodeCode       string          `json:"node_code"`
		Structure      json.RawMessage `json:"structure"`
		EditableRanges json.RawMessage `json:"editable_ranges"`
		ReadOnlyRanges json.RawMessage `json:"readonly_ranges"`
		Version        int             `json:"version"`
		Status         int             `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数错误: %v", err))
		return
	}
	if req.Code != "" {
		tpl.Code = req.Code
	}
	if req.Name != "" {
		tpl.Name = req.Name
	}
	tpl.SampleType = req.SampleType
	tpl.NodeCode = req.NodeCode
	if len(req.Structure) > 0 {
		tpl.Structure = req.Structure
	}
	if len(req.EditableRanges) > 0 {
		tpl.EditableRanges = req.EditableRanges
	}
	if len(req.ReadOnlyRanges) > 0 {
		tpl.ReadOnlyRanges = req.ReadOnlyRanges
	}
	if req.Version > 0 {
		tpl.Version = req.Version
	}
	tpl.Status = req.Status
	if err := h.getDB(c).Save(&tpl).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("更新采样单模板失败: %v", err))
		return
	}
	utils.Success(c, tpl)
}

func (h *SamplingSheetHandler) DeleteTemplate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	res := h.getDB(c).Delete(&model.SamplingSheetTemplate{}, id)
	if res.Error != nil {
		utils.InternalError(c, fmt.Sprintf("删除采样单模板失败: %v", res.Error))
		return
	}
	if res.RowsAffected == 0 {
		utils.NotFound(c, "采样单模板不存在")
		return
	}
	utils.Success(c, gin.H{"deleted": id})
}

func (h *SamplingSheetHandler) GetTemplateBySampleType(c *gin.Context) {
	sampleType := c.Param("sampleType")
	if sampleType == "" {
		utils.BadRequest(c, "缺少样品类型参数")
		return
	}
	var tpl model.SamplingSheetTemplate
	if err := h.getDB(c).
		Where("sample_type = ? AND status = 1", sampleType).
		Order("version DESC").
		First(&tpl).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.NotFound(c, "该样品类型暂无可用采样单模板")
		} else {
			utils.InternalError(c, fmt.Sprintf("查询采样单模板失败: %v", err))
		}
		return
	}
	utils.Success(c, tpl)
}

// ─── SamplingSheet (instance) CRUD ───────────────────────────────

func (h *SamplingSheetHandler) List(c *gin.Context) {
	var items []model.SamplingSheet
	query := h.getDB(c).Preload("Template").Order("id DESC")
	if taskOrderID := c.Query("task_order_id"); taskOrderID != "" {
		query = query.Where("task_order_id = ?", taskOrderID)
	}
	if nodeCode := c.Query("node_code"); nodeCode != "" {
		query = query.Where("node_code = ?", nodeCode)
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if err := query.Find(&items).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("查询采样单失败: %v", err))
		return
	}
	utils.Success(c, items)
}

func (h *SamplingSheetHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	var item model.SamplingSheet
	if err := h.getDB(c).Preload("Template").First(&item, id).Error; err != nil {
		utils.NotFound(c, "采样单不存在")
		return
	}
	utils.Success(c, item)
}

func (h *SamplingSheetHandler) Create(c *gin.Context) {
	var req struct {
		TaskOrderID   uint            `json:"task_order_id" binding:"required"`
		SamplingPoint string          `json:"sampling_point"`
		SampleType    string          `json:"sample_type"`
		TemplateID    *uint           `json:"template_id"`
		NodeCode      string          `json:"node_code"`
		SheetData     json.RawMessage `json:"sheet_data"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数错误: %v", err))
		return
	}

	if req.TemplateID == nil {
		var tpl model.SamplingSheetTemplate
		query := h.getDB(c).Where("status = 1")
		if req.SampleType != "" {
			query = query.Where("sample_type = ?", req.SampleType)
		}
		if req.NodeCode != "" {
			query = query.Where("node_code = ?", req.NodeCode)
		}
		if err := query.Order("version DESC").First(&tpl).Error; err != nil {
			if err != gorm.ErrRecordNotFound {
				utils.InternalError(c, fmt.Sprintf("查询采样单模板失败: %v", err))
				return
			}
		} else {
			id := tpl.ID
			req.TemplateID = &id
			if len(req.SheetData) == 0 {
				req.SheetData = tpl.Structure
			}
		}
	}

	instance := model.SamplingSheet{
		TaskOrderID:   req.TaskOrderID,
		SamplingPoint: req.SamplingPoint,
		TemplateID:    req.TemplateID,
		NodeCode:      req.NodeCode,
		SheetData:     req.SheetData,
		Status:        0,
	}
	if len(instance.SheetData) == 0 {
		instance.SheetData = json.RawMessage("{}")
	}
	if err := h.getDB(c).Create(&instance).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("创建采样单失败: %v", err))
		return
	}
	utils.Success(c, instance)
}

func (h *SamplingSheetHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	var instance model.SamplingSheet
	if err := h.getDB(c).First(&instance, id).Error; err != nil {
		utils.NotFound(c, "采样单不存在")
		return
	}
	var req struct {
		SamplingPoint  string          `json:"sampling_point"`
		SheetData      json.RawMessage `json:"sheet_data"`
		FormulaResults json.RawMessage `json:"formula_results"`
		Status         *int            `json:"status"`
		NodeCode       string          `json:"node_code"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数错误: %v", err))
		return
	}
	if req.SamplingPoint != "" {
		instance.SamplingPoint = req.SamplingPoint
	}
	if len(req.SheetData) > 0 {
		instance.SheetData = req.SheetData
	}
	if len(req.FormulaResults) > 0 {
		instance.FormulaResults = req.FormulaResults
	}
	if req.Status != nil {
		instance.Status = *req.Status
	}
	if req.NodeCode != "" {
		instance.NodeCode = req.NodeCode
	}
	userID := middleware.GetUserID(c)
	instance.UpdatedBy = &userID
	if err := h.getDB(c).Save(&instance).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("保存采样单失败: %v", err))
		return
	}
	utils.Success(c, instance)
}

func (h *SamplingSheetHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	res := h.getDB(c).Delete(&model.SamplingSheet{}, id)
	if res.Error != nil {
		utils.InternalError(c, fmt.Sprintf("删除采样单失败: %v", res.Error))
		return
	}
	if res.RowsAffected == 0 {
		utils.NotFound(c, "采样单不存在")
		return
	}
	utils.Success(c, gin.H{"deleted": id})
}
