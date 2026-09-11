package system

import (
	"lims-backend/internal/middleware"
	"lims-backend/internal/model"
	"lims-backend/internal/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type RoleHandler struct {
	db *gorm.DB
}

func NewRoleHandler(db *gorm.DB) *RoleHandler {
	return &RoleHandler{db: db}
}

func (h *RoleHandler) getDB(c *gin.Context) *gorm.DB {
	if db := middleware.GetDB(c); db != nil {
		return db
	}
	return h.db
}

func (h *RoleHandler) List(c *gin.Context) {
	page, pageSize, offset := utils.GetPagination(c)
	var roles []model.Role
	var total int64
	h.getDB(c).Preload("Permissions").Count(&total)
	if err := h.getDB(c).Preload("Permissions").Order("id ASC").Limit(pageSize).Offset(offset).Find(&roles).Error; err != nil {
		utils.InternalError(c, "查询角色列表失败")
		return
	}
	utils.SuccessPage(c, roles, total, page, pageSize)
}

func (h *RoleHandler) Get(c *gin.Context) {
	var role model.Role
	if err := h.getDB(c).Preload("Permissions").First(&role, c.Param("id")).Error; err != nil {
		utils.NotFound(c, "角色不存在")
		return
	}
	utils.Success(c, role)
}

type CreateRoleRequest struct {
	Name   string `json:"name" binding:"required"`
	Code   string `json:"code" binding:"required"`
	Status int    `json:"status"`
	Remark string `json:"remark"`
}

func (h *RoleHandler) Create(c *gin.Context) {
	var req CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}

	role := model.Role{
		Name:   req.Name,
		Code:   req.Code,
		Status: req.Status,
		Remark: req.Remark,
	}

	if err := h.getDB(c).Create(&role).Error; err != nil {
		utils.InternalError(c, "创建角色失败")
		return
	}
	utils.Created(c, role)
}

func (h *RoleHandler) Update(c *gin.Context) {
	var role model.Role
	if err := h.getDB(c).First(&role, c.Param("id")).Error; err != nil {
		utils.NotFound(c, "角色不存在")
		return
	}

	var req CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}

	if err := h.getDB(c).Model(&role).Updates(map[string]interface{}{
		"name":   req.Name,
		"code":   req.Code,
		"status": req.Status,
		"remark": req.Remark,
	}).Error; err != nil {
		utils.InternalError(c, "更新角色失败")
		return
	}
	utils.Success(c, role)
}

func (h *RoleHandler) Delete(c *gin.Context) {
	// Check if any users have this role
	var count int64
	h.getDB(c).Table("user_roles").Where("role_id = ?", c.Param("id")).Count(&count)
	if count > 0 {
		utils.BadRequest(c, "该角色下存在用户，无法删除")
		return
	}

	if err := h.getDB(c).Delete(&model.Role{}, c.Param("id")).Error; err != nil {
		utils.InternalError(c, "删除角色失败")
		return
	}
	utils.Success(c, nil)
}

type UpdateRolePermissionsRequest struct {
	PermissionIDs []uint `json:"permission_ids"`
}

func (h *RoleHandler) UpdatePermissions(c *gin.Context) {
	var req UpdateRolePermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}

	var role model.Role
	if err := h.getDB(c).First(&role, c.Param("id")).Error; err != nil {
		utils.NotFound(c, "角色不存在")
		return
	}

	if err := h.getDB(c).Model(&role).Association("Permissions").Replace(req.PermissionIDs); err != nil {
		utils.InternalError(c, "更新权限失败")
		return
	}
	utils.Success(c, nil)
}
