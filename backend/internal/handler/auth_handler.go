package handler

import (
	"time"

	"lims-backend/internal/config"
	"lims-backend/internal/middleware"
	"lims-backend/internal/model"
	"lims-backend/internal/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AuthHandler struct {
	cfg *config.Config
	db  *gorm.DB
}

func NewAuthHandler(cfg *config.Config, db *gorm.DB) *AuthHandler {
	return &AuthHandler{cfg: cfg, db: db}
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请输入用户名和密码")
		return
	}

	var user model.User
	if err := h.db.Where("username = ?", req.Username).Preload("Dept").Preload("Roles").First(&user).Error; err != nil {
		utils.Unauthorized(c, "用户名或密码错误")
		return
	}

	if user.Status == 0 {
		utils.Forbidden(c, "账号已被禁用")
		return
	}

	if !utils.CheckPassword(user.Password, req.Password) {
		utils.Unauthorized(c, "用户名或密码错误")
		return
	}

	token, err := utils.GenerateToken(
		&h.cfg.JWT,
		user.ID,
		user.Username,
		user.DeptID,
		user.IsAdmin,
	)
	if err != nil {
		utils.InternalError(c, "生成令牌失败")
		return
	}

	// Update last login time
	h.db.Model(&user).Update("last_login", time.Now())

	utils.Success(c, gin.H{
		"token": token,
		"user": gin.H{
			"id":        user.ID,
			"username":  user.Username,
			"real_name": user.RealName,
			"dept_id":   user.DeptID,
			"is_admin":  user.IsAdmin,
		},
	})
}

func (h *AuthHandler) Profile(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		utils.Unauthorized(c, "未认证")
		return
	}

	var user model.User
	if err := h.db.Preload("Dept").Preload("Roles.Permissions").First(&user, userID).Error; err != nil {
		utils.NotFound(c, "用户不存在")
		return
	}

	utils.Success(c, user)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	utils.Success(c, nil)
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	userID := middleware.GetUserID(c)
	username := c.GetString("username")
	deptID := middleware.GetDeptID(c)
	isAdmin := middleware.IsAdmin(c)

	token, err := utils.GenerateToken(&h.cfg.JWT, userID, username, deptID, isAdmin)
	if err != nil {
		utils.InternalError(c, "刷新令牌失败")
		return
	}

	utils.Success(c, gin.H{"token": token})
}