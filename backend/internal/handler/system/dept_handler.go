package system

import (
	"lims-backend/internal/middleware"
	"lims-backend/internal/model"
	"lims-backend/internal/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type DeptHandler struct {
	db *gorm.DB
}

func NewDeptHandler(db *gorm.DB) *DeptHandler {
	return &DeptHandler{db: db}
}

func (h *DeptHandler) getDB(c *gin.Context) *gorm.DB {
	if db := middleware.GetDB(c); db != nil {
		return db
	}
	return h.db
}

func (h *DeptHandler) List(c *gin.Context) {
	page, pageSize, offset := utils.GetPagination(c)
	var depts []model.Dept
	var total int64
	h.getDB(c).Preload("Children").Count(&total)
	if err := h.getDB(c).Preload("Children").Order("sort ASC").Limit(pageSize).Offset(offset).Find(&depts).Error; err != nil {
		utils.InternalError(c, "查询部门列表失败")
		return
	}
	utils.SuccessPage(c, depts, total, page, pageSize)
}

func (h *DeptHandler) Get(c *gin.Context) {
	var dept model.Dept
	if err := h.getDB(c).First(&dept, c.Param("id")).Error; err != nil {
		utils.NotFound(c, "部门不存在")
		return
	}
	utils.Success(c, dept)
}

type CreateDeptRequest struct {
	Name     string `json:"name" binding:"required"`
	Code     string `json:"code" binding:"required"`
	Sort     int    `json:"sort"`
	ParentID *uint  `json:"parent_id"`
	Status   int    `json:"status"`
}

func (h *DeptHandler) Create(c *gin.Context) {
	var req CreateDeptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}

	dept := model.Dept{
		Name:     req.Name,
		Code:     req.Code,
		Sort:     req.Sort,
		ParentID: req.ParentID,
		Status:   req.Status,
	}

	if err := h.getDB(c).Create(&dept).Error; err != nil {
		utils.InternalError(c, "创建部门失败")
		return
	}
	utils.Created(c, dept)
}

func (h *DeptHandler) Update(c *gin.Context) {
	var dept model.Dept
	if err := h.getDB(c).First(&dept, c.Param("id")).Error; err != nil {
		utils.NotFound(c, "部门不存在")
		return
	}

	var req CreateDeptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}

	updates := map[string]interface{}{
		"name":   req.Name,
		"code":   req.Code,
		"sort":   req.Sort,
		"status": req.Status,
	}
	if req.ParentID != nil {
		updates["parent_id"] = *req.ParentID
	}

	if err := h.getDB(c).Model(&dept).Updates(updates).Error; err != nil {
		utils.InternalError(c, "更新部门失败")
		return
	}
	utils.Success(c, dept)
}

func (h *DeptHandler) Delete(c *gin.Context) {
	// Check if any users belong to this department
	var count int64
	h.getDB(c).Model(&model.User{}).Where("dept_id = ?", c.Param("id")).Count(&count)
	if count > 0 {
		utils.BadRequest(c, "该部门下存在用户，无法删除")
		return
	}

	if err := h.getDB(c).Delete(&model.Dept{}, c.Param("id")).Error; err != nil {
		utils.InternalError(c, "删除部门失败")
		return
	}
	utils.Success(c, nil)
}
