package handler

import (
	"encoding/json"
	"fmt"
	"lims-backend/internal/middleware"
	"lims-backend/internal/model"
	"lims-backend/internal/service"
	"lims-backend/internal/utils"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// LabSheetHandler handles CRUD for LabSheet (检验单实例) and LabSheetTemplate (模板).
type LabSheetHandler struct {
	logger *zap.Logger
	db     *gorm.DB
}

func NewLabSheetHandler(logger *zap.Logger, db *gorm.DB) *LabSheetHandler {
	return &LabSheetHandler{logger: logger, db: db}
}

func (h *LabSheetHandler) getDB(c *gin.Context) *gorm.DB {
	if db := middleware.GetDB(c); db != nil {
		return db
	}
	return h.db
}

func toUint(s string) (uint, error) {
	n, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(n), nil
}

// ─── LabSheetTemplate CRUD ──────────────────────────────────────

func (h *LabSheetHandler) ListTemplates(c *gin.Context) {
	page, pageSize, offset := utils.GetPagination(c)
	var items []model.LabSheetTemplate
	query := h.getDB(c).Order("version DESC, id DESC")
	if name := c.Query("name"); name != "" {
		query = query.Where("name ILIKE ?", "%"+name+"%")
	}
	if testItemID := c.Query("test_item_id"); testItemID != "" {
		query = query.Where("test_item_id = ?", testItemID)
	}
	if nodeCode := c.Query("node_code"); nodeCode != "" {
		query = query.Where("node_code = ?", nodeCode)
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	var total int64
	if err := query.Model(&model.LabSheetTemplate{}).Count(&total).Error; err != nil {
		utils.InternalError(c, "查询失败")
		return
	}
	if err := query.Limit(pageSize).Offset(offset).Find(&items).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("查询模板失败: %v", err))
		return
	}
	utils.SuccessPage(c, items, total, page, pageSize)
}

func (h *LabSheetHandler) GetTemplate(c *gin.Context) {
	id, err := toUint(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	var item model.LabSheetTemplate
	if err := h.getDB(c).First(&item, id).Error; err != nil {
		utils.NotFound(c, "模板不存在")
		return
	}
	utils.Success(c, item)
}

func (h *LabSheetHandler) CreateTemplate(c *gin.Context) {
	var req struct {
		Code           string          `json:"code" binding:"required"`
		Name           string          `json:"name" binding:"required"`
		TestItemID     uint            `json:"test_item_id"`
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
	tpl := model.LabSheetTemplate{
		Code:           req.Code,
		Name:           req.Name,
		TestItemID:     req.TestItemID,
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
		utils.InternalError(c, fmt.Sprintf("创建模板失败: %v", err))
		return
	}
	utils.Success(c, tpl)
}

func (h *LabSheetHandler) UpdateTemplate(c *gin.Context) {
	id, err := toUint(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	var tpl model.LabSheetTemplate
	if err := h.getDB(c).First(&tpl, id).Error; err != nil {
		utils.NotFound(c, "模板不存在")
		return
	}
	var req struct {
		Code           string          `json:"code"`
		Name           string          `json:"name"`
		TestItemID     uint            `json:"test_item_id"`
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
	tpl.TestItemID = req.TestItemID
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
		utils.InternalError(c, fmt.Sprintf("更新模板失败: %v", err))
		return
	}
	utils.Success(c, tpl)
}

func (h *LabSheetHandler) DeleteTemplate(c *gin.Context) {
	id, err := toUint(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	res := h.getDB(c).Delete(&model.LabSheetTemplate{}, id)
	if res.Error != nil {
		utils.InternalError(c, fmt.Sprintf("删除模板失败: %v", res.Error))
		return
	}
	if res.RowsAffected == 0 {
		utils.NotFound(c, "模板不存在")
		return
	}
	utils.Success(c, gin.H{"deleted": id})
}

func (h *LabSheetHandler) GetTemplateByItem(c *gin.Context) {
	testItemID, err := toUint(c.Param("testItemID"))
	if err != nil {
		utils.BadRequest(c, "无效的检测项目ID")
		return
	}
	var tpl model.LabSheetTemplate
	if err := h.getDB(c).
		Where("test_item_id = ? AND status = 1", testItemID).
		Order("version DESC").
		First(&tpl).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.NotFound(c, "该检测项目暂无可用模板")
		} else {
			utils.InternalError(c, fmt.Sprintf("查询模板失败: %v", err))
		}
		return
	}
	utils.Success(c, tpl)
}

// ─── LabSheet (instance) CRUD ────────────────────────────────────

func (h *LabSheetHandler) List(c *gin.Context) {
	page, pageSize, offset := utils.GetPagination(c)
	var items []model.LabSheet
	query := service.ApplyTaskOrderScope(h.getDB(c).Preload("Template"), c).Order("id DESC")
	if taskOrderID := c.Query("task_order_id"); taskOrderID != "" {
		query = query.Where("task_order_id = ?", taskOrderID)
	}
	if testItemID := c.Query("test_item_id"); testItemID != "" {
		query = query.Where("test_item_id = ?", testItemID)
	}
	if nodeCode := c.Query("node_code"); nodeCode != "" {
		query = query.Where("node_code = ?", nodeCode)
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	var total int64
	if err := query.Model(&model.LabSheet{}).Count(&total).Error; err != nil {
		utils.InternalError(c, "查询失败")
		return
	}
	if err := query.Limit(pageSize).Offset(offset).Find(&items).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("查询检验单失败: %v", err))
		return
	}
	utils.SuccessPage(c, items, total, page, pageSize)
}

func (h *LabSheetHandler) Get(c *gin.Context) {
	id, err := toUint(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	var item model.LabSheet
	q := service.ApplyTaskOrderScope(h.getDB(c).Preload("Template"), c)
	if err := q.First(&item, id).Error; err != nil {
		utils.NotFound(c, "检验单不存在")
		return
	}
	utils.Success(c, item)
}

func (h *LabSheetHandler) Create(c *gin.Context) {
	var req struct {
		TaskOrderID uint            `json:"task_order_id" binding:"required"`
		TestItemID  uint            `json:"test_item_id" binding:"required"`
		TemplateID  *uint           `json:"template_id"`
		NodeCode    string          `json:"node_code"`
		SheetData   json.RawMessage `json:"sheet_data"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数错误: %v", err))
		return
	}

	if req.TemplateID == nil {
		var tpl model.LabSheetTemplate
		if err := h.getDB(c).
			Where("test_item_id = ? AND status = 1", req.TestItemID).
			Order("version DESC").
			First(&tpl).Error; err != nil {
			if err != gorm.ErrRecordNotFound {
				utils.InternalError(c, fmt.Sprintf("查询模板失败: %v", err))
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

	instance := model.LabSheet{
		TaskOrderID: req.TaskOrderID,
		TestItemID:  req.TestItemID,
		TemplateID:  req.TemplateID,
		NodeCode:    req.NodeCode,
		SheetData:   req.SheetData,
		Status:      0,
	}
	if len(instance.SheetData) == 0 {
		instance.SheetData = json.RawMessage("{}")
	}
	if err := h.getDB(c).Create(&instance).Error; err != nil {
		utils.InternalError(c, fmt.Sprintf("创建检验单失败: %v", err))
		return
	}
	utils.Success(c, instance)
}

func (h *LabSheetHandler) Update(c *gin.Context) {
	id, err := toUint(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	var instance model.LabSheet
	if err := h.getDB(c).First(&instance, id).Error; err != nil {
		utils.NotFound(c, "检验单不存在")
		return
	}
	var req struct {
		SheetData      json.RawMessage `json:"sheet_data"`
		FormulaResults json.RawMessage `json:"formula_results"`
		Status         *int            `json:"status"`
		NodeCode       string          `json:"node_code"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, fmt.Sprintf("参数错误: %v", err))
		return
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
		utils.InternalError(c, fmt.Sprintf("保存检验单失败: %v", err))
		return
	}
	utils.Success(c, instance)
}

func (h *LabSheetHandler) Delete(c *gin.Context) {
	id, err := toUint(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	res := h.getDB(c).Delete(&model.LabSheet{}, id)
	if res.Error != nil {
		utils.InternalError(c, fmt.Sprintf("删除检验单失败: %v", res.Error))
		return
	}
	if res.RowsAffected == 0 {
		utils.NotFound(c, "检验单不存在")
		return
	}
	utils.Success(c, gin.H{"deleted": id})
}
