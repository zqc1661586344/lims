package handler

import (
	"fmt"
	"lims-backend/internal/middleware"
	"lims-backend/internal/model"
	"lims-backend/internal/utils"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ------------------------------------------------------------
// TestItemHandler
// ------------------------------------------------------------

type TestItemHandler struct {
	db *gorm.DB
}

func NewTestItemHandler(db *gorm.DB) *TestItemHandler {
	return &TestItemHandler{db: db}
}

func (h *TestItemHandler) getDB(c *gin.Context) *gorm.DB {
	if db := middleware.GetDB(c); db != nil {
		return db
	}
	return h.db
}

func (h *TestItemHandler) List(c *gin.Context) {
	var items []model.TestItem
	query := h.getDB(c).Preload("Standard")

	if keyword := c.Query("keyword"); keyword != "" {
		query = query.Where("name LIKE ? OR code LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if cat := c.Query("category"); cat != "" {
		query = query.Where("category = ?", cat)
	}
	if err := query.Order("id DESC").Find(&items).Error; err != nil {
		utils.InternalError(c, "查询检测项目失败")
		return
	}
	utils.Success(c, items)
}

func (h *TestItemHandler) Get(c *gin.Context) {
	var item model.TestItem
	if err := h.getDB(c).Preload("Standard").First(&item, c.Param("id")).Error; err != nil {
		utils.NotFound(c, "检测项目不存在")
		return
	}
	utils.Success(c, item)
}

type CreateTestItemRequest struct {
	Name       string  `json:"name" binding:"required"`
	Code       string  `json:"code" binding:"required"`
	Category   string  `json:"category"`
	Unit       string  `json:"unit"`
	Method     string  `json:"method"`
	StandardID *uint   `json:"standard_id"`
	Price      float64 `json:"price"`
	Status     int     `json:"status"`
}

func (h *TestItemHandler) Create(c *gin.Context) {
	var req CreateTestItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}

	item := model.TestItem{
		Name:       req.Name,
		Code:       req.Code,
		Category:   req.Category,
		Unit:       req.Unit,
		Method:     req.Method,
		StandardID: req.StandardID,
		Price:      req.Price,
		Status:     req.Status,
	}
	if item.Status == 0 {
		item.Status = 1
	}

	if err := h.getDB(c).Create(&item).Error; err != nil {
		utils.InternalError(c, "创建检测项目失败")
		return
	}
	utils.Created(c, item)
}

type UpdateTestItemRequest struct {
	Name       string  `json:"name"`
	Code       string  `json:"code"`
	Category   string  `json:"category"`
	Unit       string  `json:"unit"`
	Method     string  `json:"method"`
	StandardID *uint   `json:"standard_id"`
	Price      float64 `json:"price"`
	Status     *int    `json:"status"`
}

func (h *TestItemHandler) Update(c *gin.Context) {
	var item model.TestItem
	if err := h.getDB(c).First(&item, c.Param("id")).Error; err != nil {
		utils.NotFound(c, "检测项目不存在")
		return
	}

	var req UpdateTestItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Code != "" {
		updates["code"] = req.Code
	}
	if req.Category != "" {
		updates["category"] = req.Category
	}
	if req.Unit != "" {
		updates["unit"] = req.Unit
	}
	if req.Method != "" {
		updates["method"] = req.Method
	}
	if req.StandardID != nil {
		updates["standard_id"] = *req.StandardID
	}
	if req.Price != 0 {
		updates["price"] = req.Price
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}

	if err := h.getDB(c).Model(&item).Updates(updates).Error; err != nil {
		utils.InternalError(c, "更新检测项目失败")
		return
	}
	utils.Success(c, item)
}

func (h *TestItemHandler) Delete(c *gin.Context) {
	if err := h.getDB(c).Delete(&model.TestItem{}, c.Param("id")).Error; err != nil {
		utils.InternalError(c, "删除检测项目失败")
		return
	}
	utils.Success(c, nil)
}

// ------------------------------------------------------------
// TestStandardHandler
// ------------------------------------------------------------

type TestStandardHandler struct {
	db *gorm.DB
}

func NewTestStandardHandler(db *gorm.DB) *TestStandardHandler {
	return &TestStandardHandler{db: db}
}

func (h *TestStandardHandler) getDB(c *gin.Context) *gorm.DB {
	if db := middleware.GetDB(c); db != nil {
		return db
	}
	return h.db
}

func (h *TestStandardHandler) List(c *gin.Context) {
	var standards []model.TestStandard
	query := h.getDB(c)

	if keyword := c.Query("keyword"); keyword != "" {
		query = query.Where("name LIKE ? OR code LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if err := query.Order("id DESC").Find(&standards).Error; err != nil {
		utils.InternalError(c, "查询检测标准失败")
		return
	}
	utils.Success(c, standards)
}

func (h *TestStandardHandler) Get(c *gin.Context) {
	var std model.TestStandard
	if err := h.getDB(c).First(&std, c.Param("id")).Error; err != nil {
		utils.NotFound(c, "检测标准不存在")
		return
	}
	utils.Success(c, std)
}

type CreateTestStandardRequest struct {
	Name        string `json:"name" binding:"required"`
	Code        string `json:"code" binding:"required"`
	Issuer      string `json:"issuer"`
	Version     string `json:"version"`
	PublishDate string `json:"publish_date"`
	FilePath    string `json:"file_path"`
	Status      int    `json:"status"`
}

func (h *TestStandardHandler) Create(c *gin.Context) {
	var req CreateTestStandardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}

	std := model.TestStandard{
		Name:     req.Name,
		Code:     req.Code,
		Issuer:   req.Issuer,
		Version:  req.Version,
		FilePath: req.FilePath,
		Status:   req.Status,
	}
	if std.Status == 0 {
		std.Status = 1
	}
	if req.PublishDate != "" {
		parsed, err := parseDate(req.PublishDate)
		if err != nil {
			utils.BadRequest(c, "日期格式错误")
			return
		}
		std.PublishDate = parsed
	}

	if err := h.getDB(c).Create(&std).Error; err != nil {
		utils.InternalError(c, "创建检测标准失败")
		return
	}
	utils.Created(c, std)
}

type UpdateTestStandardRequest struct {
	Name        string `json:"name"`
	Code        string `json:"code"`
	Issuer      string `json:"issuer"`
	Version     string `json:"version"`
	PublishDate string `json:"publish_date"`
	FilePath    string `json:"file_path"`
	Status      *int   `json:"status"`
}

func (h *TestStandardHandler) Update(c *gin.Context) {
	var std model.TestStandard
	if err := h.getDB(c).First(&std, c.Param("id")).Error; err != nil {
		utils.NotFound(c, "检测标准不存在")
		return
	}

	var req UpdateTestStandardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Code != "" {
		updates["code"] = req.Code
	}
	if req.Issuer != "" {
		updates["issuer"] = req.Issuer
	}
	if req.Version != "" {
		updates["version"] = req.Version
	}
	if req.FilePath != "" {
		updates["file_path"] = req.FilePath
	}
	if req.PublishDate != "" {
		parsed, err := parseDate(req.PublishDate)
		if err != nil {
			utils.BadRequest(c, "日期格式错误")
			return
		}
		updates["publish_date"] = parsed
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}

	if err := h.getDB(c).Model(&std).Updates(updates).Error; err != nil {
		utils.InternalError(c, "更新检测标准失败")
		return
	}
	utils.Success(c, std)
}

func (h *TestStandardHandler) Delete(c *gin.Context) {
	if err := h.getDB(c).Delete(&model.TestStandard{}, c.Param("id")).Error; err != nil {
		utils.InternalError(c, "删除检测标准失败")
		return
	}
	utils.Success(c, nil)
}

// ------------------------------------------------------------
// EquipmentHandler
// ------------------------------------------------------------

type EquipmentHandler struct {
	db *gorm.DB
}

func NewEquipmentHandler(db *gorm.DB) *EquipmentHandler {
	return &EquipmentHandler{db: db}
}

func (h *EquipmentHandler) getDB(c *gin.Context) *gorm.DB {
	if db := middleware.GetDB(c); db != nil {
		return db
	}
	return h.db
}

func (h *EquipmentHandler) List(c *gin.Context) {
	var equipments []model.Equipment
	query := h.getDB(c)

	if keyword := c.Query("keyword"); keyword != "" {
		query = query.Where("name LIKE ? OR code LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if err := query.Order("id DESC").Find(&equipments).Error; err != nil {
		utils.InternalError(c, "查询设备失败")
		return
	}
	utils.Success(c, equipments)
}

func (h *EquipmentHandler) Get(c *gin.Context) {
	var equip model.Equipment
	if err := h.getDB(c).First(&equip, c.Param("id")).Error; err != nil {
		utils.NotFound(c, "设备不存在")
		return
	}
	utils.Success(c, equip)
}

type CreateEquipmentRequest struct {
	Name            string `json:"name" binding:"required"`
	Code            string `json:"code" binding:"required"`
	Model           string `json:"model"`
	Factory         string `json:"factory"`
	CalibrationDate string `json:"calibration_date"`
	NextCalDate     string `json:"next_cal_date"`
	Status          int    `json:"status"`
}

func (h *EquipmentHandler) Create(c *gin.Context) {
	var req CreateEquipmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}

	equip := model.Equipment{
		Name:    req.Name,
		Code:    req.Code,
		Model:   req.Model,
		Factory: req.Factory,
		Status:  req.Status,
	}
	if equip.Status == 0 {
		equip.Status = 1
	}
	if req.CalibrationDate != "" {
		parsed, err := parseDate(req.CalibrationDate)
		if err != nil {
			utils.BadRequest(c, "校准日期格式错误")
			return
		}
		equip.CalibrationDate = parsed
	}
	if req.NextCalDate != "" {
		parsed, err := parseDate(req.NextCalDate)
		if err != nil {
			utils.BadRequest(c, "下次校准日期格式错误")
			return
		}
		equip.NextCalDate = parsed
	}

	if err := h.getDB(c).Create(&equip).Error; err != nil {
		utils.InternalError(c, "创建设备失败")
		return
	}
	utils.Created(c, equip)
}

type UpdateEquipmentRequest struct {
	Name            string `json:"name"`
	Code            string `json:"code"`
	Model           string `json:"model"`
	Factory         string `json:"factory"`
	CalibrationDate string `json:"calibration_date"`
	NextCalDate     string `json:"next_cal_date"`
	Status          *int   `json:"status"`
}

func (h *EquipmentHandler) Update(c *gin.Context) {
	var equip model.Equipment
	if err := h.getDB(c).First(&equip, c.Param("id")).Error; err != nil {
		utils.NotFound(c, "设备不存在")
		return
	}

	var req UpdateEquipmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Code != "" {
		updates["code"] = req.Code
	}
	if req.Model != "" {
		updates["model"] = req.Model
	}
	if req.Factory != "" {
		updates["factory"] = req.Factory
	}
	if req.CalibrationDate != "" {
		parsed, err := parseDate(req.CalibrationDate)
		if err != nil {
			utils.BadRequest(c, "校准日期格式错误")
			return
		}
		updates["calibration_date"] = parsed
	}
	if req.NextCalDate != "" {
		parsed, err := parseDate(req.NextCalDate)
		if err != nil {
			utils.BadRequest(c, "下次校准日期格式错误")
			return
		}
		updates["next_cal_date"] = parsed
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}

	if err := h.getDB(c).Model(&equip).Updates(updates).Error; err != nil {
		utils.InternalError(c, "更新设备失败")
		return
	}
	utils.Success(c, equip)
}

func (h *EquipmentHandler) Delete(c *gin.Context) {
	if err := h.getDB(c).Delete(&model.Equipment{}, c.Param("id")).Error; err != nil {
		utils.InternalError(c, "删除设备失败")
		return
	}
	utils.Success(c, nil)
}

// ------------------------------------------------------------
// ReagentHandler
// ------------------------------------------------------------

type ReagentHandler struct {
	db *gorm.DB
}

func NewReagentHandler(db *gorm.DB) *ReagentHandler {
	return &ReagentHandler{db: db}
}

func (h *ReagentHandler) getDB(c *gin.Context) *gorm.DB {
	if db := middleware.GetDB(c); db != nil {
		return db
	}
	return h.db
}

func (h *ReagentHandler) List(c *gin.Context) {
	var reagents []model.Reagent
	query := h.getDB(c)

	if keyword := c.Query("keyword"); keyword != "" {
		query = query.Where("name LIKE ? OR code LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if err := query.Order("id DESC").Find(&reagents).Error; err != nil {
		utils.InternalError(c, "查询试剂失败")
		return
	}
	utils.Success(c, reagents)
}

func (h *ReagentHandler) Get(c *gin.Context) {
	var reagent model.Reagent
	if err := h.getDB(c).First(&reagent, c.Param("id")).Error; err != nil {
		utils.NotFound(c, "试剂不存在")
		return
	}
	utils.Success(c, reagent)
}

type CreateReagentRequest struct {
	Name         string  `json:"name" binding:"required"`
	Code         string  `json:"code" binding:"required"`
	Spec         string  `json:"spec"`
	Manufacturer string  `json:"manufacturer"`
	BatchNo      string  `json:"batch_no"`
	StockQty     float64 `json:"stock_qty"`
	Unit         string  `json:"unit"`
	ExpireDate   string  `json:"expire_date"`
	Status       int     `json:"status"`
}

func (h *ReagentHandler) Create(c *gin.Context) {
	var req CreateReagentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}

	reagent := model.Reagent{
		Name:         req.Name,
		Code:         req.Code,
		Spec:         req.Spec,
		Manufacturer: req.Manufacturer,
		BatchNo:      req.BatchNo,
		StockQty:     req.StockQty,
		Unit:         req.Unit,
		Status:       req.Status,
	}
	if reagent.Status == 0 {
		reagent.Status = 1
	}
	if req.ExpireDate != "" {
		parsed, err := parseDate(req.ExpireDate)
		if err != nil {
			utils.BadRequest(c, "有效期格式错误")
			return
		}
		reagent.ExpireDate = parsed
	}

	if err := h.getDB(c).Create(&reagent).Error; err != nil {
		utils.InternalError(c, "创建试剂失败")
		return
	}
	utils.Created(c, reagent)
}

type UpdateReagentRequest struct {
	Name         string  `json:"name"`
	Code         string  `json:"code"`
	Spec         string  `json:"spec"`
	Manufacturer string  `json:"manufacturer"`
	BatchNo      string  `json:"batch_no"`
	StockQty     float64 `json:"stock_qty"`
	Unit         string  `json:"unit"`
	ExpireDate   string  `json:"expire_date"`
	Status       *int    `json:"status"`
}

func (h *ReagentHandler) Update(c *gin.Context) {
	var reagent model.Reagent
	if err := h.getDB(c).First(&reagent, c.Param("id")).Error; err != nil {
		utils.NotFound(c, "试剂不存在")
		return
	}

	var req UpdateReagentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Code != "" {
		updates["code"] = req.Code
	}
	if req.Spec != "" {
		updates["spec"] = req.Spec
	}
	if req.Manufacturer != "" {
		updates["manufacturer"] = req.Manufacturer
	}
	if req.BatchNo != "" {
		updates["batch_no"] = req.BatchNo
	}
	if req.StockQty != 0 {
		updates["stock_qty"] = req.StockQty
	}
	if req.Unit != "" {
		updates["unit"] = req.Unit
	}
	if req.ExpireDate != "" {
		parsed, err := parseDate(req.ExpireDate)
		if err != nil {
			utils.BadRequest(c, "有效期格式错误")
			return
		}
		updates["expire_date"] = parsed
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}

	if err := h.getDB(c).Model(&reagent).Updates(updates).Error; err != nil {
		utils.InternalError(c, "更新试剂失败")
		return
	}
	utils.Success(c, reagent)
}

func (h *ReagentHandler) Delete(c *gin.Context) {
	if err := h.getDB(c).Delete(&model.Reagent{}, c.Param("id")).Error; err != nil {
		utils.InternalError(c, "删除试剂失败")
		return
	}
	utils.Success(c, nil)
}

// ------------------------------------------------------------
// Helpers
// ------------------------------------------------------------

// parseDate attempts to parse a date string in multiple formats.
func parseDate(s string) (time.Time, error) {
	formats := []string{
		"2006-01-02",
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02 15:04:05",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("cannot parse date: %s", s)
}
