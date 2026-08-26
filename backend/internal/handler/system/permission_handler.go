package system

import (
	"lims-backend/internal/model"
	"lims-backend/internal/middleware"
	"lims-backend/internal/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PermissionHandler struct {
	db *gorm.DB
}

func NewPermissionHandler(db *gorm.DB) *PermissionHandler {
	return &PermissionHandler{db: db}
}

func (h *PermissionHandler) getDB(c *gin.Context) *gorm.DB {
	if db := middleware.GetDB(c); db != nil {
		return db
	}
	return h.db
}

func (h *PermissionHandler) List(c *gin.Context) {
	var permissions []model.Permission
	if err := h.getDB(c).Preload("Children").Order("sort ASC").Find(&permissions).Error; err != nil {
		utils.InternalError(c, "查询权限列表失败")
		return
	}
	utils.Success(c, permissions)
}

type CreatePermissionRequest struct {
	Name     string `json:"name" binding:"required"`
	Code     string `json:"code" binding:"required"`
	Type     string `json:"type" binding:"required"` // menu, button, api
	ParentID *uint  `json:"parent_id"`
	Path     string `json:"path"`
	Icon     string `json:"icon"`
	Sort     int    `json:"sort"`
}

func (h *PermissionHandler) Create(c *gin.Context) {
	var req CreatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}

	perm := model.Permission{
		Name:     req.Name,
		Code:     req.Code,
		Type:     req.Type,
		ParentID: req.ParentID,
		Path:     req.Path,
		Icon:     req.Icon,
		Sort:     req.Sort,
	}

	if err := h.getDB(c).Create(&perm).Error; err != nil {
		utils.InternalError(c, "创建权限失败")
		return
	}
	utils.Created(c, perm)
}

func (h *PermissionHandler) Update(c *gin.Context) {
	var perm model.Permission
	if err := h.getDB(c).First(&perm, c.Param("id")).Error; err != nil {
		utils.NotFound(c, "权限不存在")
		return
	}

	var req CreatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}

	if err := h.getDB(c).Model(&perm).Updates(map[string]interface{}{
		"name": req.Name,
		"code": req.Code,
		"type": req.Type,
		"path": req.Path,
		"icon": req.Icon,
		"sort": req.Sort,
	}).Error; err != nil {
		utils.InternalError(c, "更新权限失败")
		return
	}
	utils.Success(c, perm)
}

func (h *PermissionHandler) Delete(c *gin.Context) {
	if err := h.getDB(c).Delete(&model.Permission{}, c.Param("id")).Error; err != nil {
		utils.InternalError(c, "删除权限失败")
		return
	}
	utils.Success(c, nil)
}