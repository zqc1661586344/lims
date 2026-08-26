package system

import (
	"lims-backend/internal/model"
	"lims-backend/internal/middleware"
	"lims-backend/internal/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserHandler struct {
	db *gorm.DB
}

func NewUserHandler(db *gorm.DB) *UserHandler {
	return &UserHandler{db: db}
}

func (h *UserHandler) getDB(c *gin.Context) *gorm.DB {
	if db := middleware.GetDB(c); db != nil {
		return db
	}
	return h.db
}

func (h *UserHandler) List(c *gin.Context) {
	var users []model.User
	query := h.getDB(c).Preload("Dept").Preload("Roles")

	if deptID := c.Query("dept_id"); deptID != "" {
		query = query.Where("dept_id = ?", deptID)
	}
	if keyword := c.Query("keyword"); keyword != "" {
		query = query.Where("username LIKE ? OR real_name LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	if err := query.Order("id DESC").Find(&users).Error; err != nil {
		utils.InternalError(c, "查询用户列表失败")
		return
	}
	utils.Success(c, users)
}

func (h *UserHandler) Get(c *gin.Context) {
	var user model.User
	if err := h.getDB(c).Preload("Dept").Preload("Roles").First(&user, c.Param("id")).Error; err != nil {
		utils.NotFound(c, "用户不存在")
		return
	}
	utils.Success(c, user)
}

type CreateUserRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	RealName string `json:"real_name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	DeptID   *uint  `json:"dept_id"`
	Status   int    `json:"status"`
	IsAdmin  bool   `json:"is_admin"`
}

func (h *UserHandler) Create(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}

	// Check username uniqueness
	var count int64
	h.getDB(c).Model(&model.User{}).Where("username = ?", req.Username).Count(&count)
	if count > 0 {
		utils.BadRequest(c, "用户名已存在")
		return
	}

	hashed, err := utils.HashPassword(req.Password)
	if err != nil {
		utils.InternalError(c, "密码加密失败")
		return
	}

	user := model.User{
		Username: req.Username,
		Password: hashed,
		RealName: req.RealName,
		Email:    req.Email,
		Phone:    req.Phone,
		DeptID:   req.DeptID,
		Status:   req.Status,
		IsAdmin:  req.IsAdmin,
	}
	if user.Status == 0 {
		user.Status = 1
	}

	if err := h.getDB(c).Create(&user).Error; err != nil {
		utils.InternalError(c, "创建用户失败")
		return
	}
	utils.Created(c, user)
}

type UpdateUserRequest struct {
	RealName string `json:"real_name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	DeptID   *uint  `json:"dept_id"`
	Status   *int   `json:"status"`
	IsAdmin  *bool  `json:"is_admin"`
}

func (h *UserHandler) Update(c *gin.Context) {
	var user model.User
	if err := h.getDB(c).First(&user, c.Param("id")).Error; err != nil {
		utils.NotFound(c, "用户不存在")
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}

	updates := map[string]interface{}{}
	if req.RealName != "" {
		updates["real_name"] = req.RealName
	}
	if req.Email != "" {
		updates["email"] = req.Email
	}
	if req.Phone != "" {
		updates["phone"] = req.Phone
	}
	if req.DeptID != nil {
		updates["dept_id"] = *req.DeptID
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.IsAdmin != nil {
		updates["is_admin"] = *req.IsAdmin
	}

	if err := h.getDB(c).Model(&user).Updates(updates).Error; err != nil {
		utils.InternalError(c, "更新用户失败")
		return
	}
	utils.Success(c, user)
}

func (h *UserHandler) Delete(c *gin.Context) {
	if err := h.getDB(c).Delete(&model.User{}, c.Param("id")).Error; err != nil {
		utils.InternalError(c, "删除用户失败")
		return
	}
	utils.Success(c, nil)
}

type UpdateUserRolesRequest struct {
	RoleIDs []uint `json:"role_ids"`
}

func (h *UserHandler) UpdateRoles(c *gin.Context) {
	var req UpdateUserRolesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}

	var user model.User
	if err := h.getDB(c).First(&user, c.Param("id")).Error; err != nil {
		utils.NotFound(c, "用户不存在")
		return
	}

	if err := h.getDB(c).Model(&user).Association("Roles").Replace(req.RoleIDs); err != nil {
		utils.InternalError(c, "更新角色失败")
		return
	}
	utils.Success(c, nil)
}