package handler

import (
	"fmt"
	"time"

	"lims-backend/internal/config"
	"lims-backend/internal/middleware"
	"lims-backend/internal/model"
	"lims-backend/internal/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AuthHandler struct {
	cfg          *config.Config
	db           *gorm.DB
	loginLimiter *middleware.RateLimiter
}

func NewAuthHandler(cfg *config.Config, db *gorm.DB, loginLimiter *middleware.RateLimiter) *AuthHandler {
	return &AuthHandler{cfg: cfg, db: db, loginLimiter: loginLimiter}
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

	userKey := "user:" + req.Username
	if h.loginLimiter != nil {
		if locked, remaining := h.loginLimiter.IsLocked(userKey); locked {
			utils.Error(c, 429, fmt.Sprintf("该账号已被锁定，请在 %d 分钟后重试", int(remaining.Minutes())+1))
			return
		}
		if !h.loginLimiter.AllowKey(userKey) {
			utils.Error(c, 429, "该账号尝试登录次数过多，请稍后再试")
			return
		}
	}

	var user model.User
	if err := h.db.Where("username = ?", req.Username).Preload("Dept").Preload("Roles").First(&user).Error; err != nil {
		if h.loginLimiter != nil {
			h.loginLimiter.RecordFail(userKey)
		}
		utils.Unauthorized(c, "用户名或密码错误")
		return
	}

	if user.Status == 0 {
		utils.Forbidden(c, "账号已被禁用")
		return
	}

	if !utils.CheckPassword(user.Password, req.Password) {
		if h.loginLimiter != nil {
			h.loginLimiter.RecordFail(userKey)
			if locked, remaining := h.loginLimiter.IsLocked(userKey); locked {
				utils.Error(c, 429, fmt.Sprintf("密码错误次数过多，账号已锁定 %d 分钟", int(remaining.Minutes())+1))
				return
			}
		}
		utils.Unauthorized(c, "用户名或密码错误")
		return
	}

	// Collect permission codes from all roles
	permissions, err := h.loadUserPermissions(user.ID, user.IsAdmin)
	if err != nil {
		utils.InternalError(c, "加载用户权限失败")
		return
	}

	token, err := utils.GenerateToken(
		&h.cfg.JWT,
		user.ID,
		user.Username,
		user.DeptID,
		user.IsAdmin,
		permissions,
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
		"permissions": permissions,
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
	jti := c.GetString("jti")
	expVal, _ := c.Get("exp")
	if jti != "" {
		if exp, ok := expVal.(time.Time); ok {
			utils.GetTokenBlacklist().Revoke(jti, exp)
		} else {
			utils.GetTokenBlacklist().Revoke(jti, time.Now().Add(time.Hour))
		}
	}
	utils.Success(c, nil)
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	userID := middleware.GetUserID(c)
	username := c.GetString("username")
	deptID := middleware.GetDeptID(c)
	isAdmin := middleware.IsAdmin(c)
	permissions := middleware.GetPermissions(c)

	token, err := utils.GenerateToken(&h.cfg.JWT, userID, username, deptID, isAdmin, permissions)
	if err != nil {
		utils.InternalError(c, "刷新令牌失败")
		return
	}

	utils.Success(c, gin.H{"token": token})
}

// loadUserPermissions queries all permission codes granted to the user via roles.
// Admin users receive EVERY permission code in the database so they can operate
// across all business modules (DeptScope still limits them to their own department).
func (h *AuthHandler) loadUserPermissions(userID uint, isAdmin bool) ([]string, error) {
	if isAdmin {
		var codes []string
		if err := h.db.Table("permissions").Select("code").Scan(&codes).Error; err != nil {
			return nil, err
		}
		return codes, nil
	}
	var codes []string
	err := h.db.Table("permissions").
		Select("permissions.code").
		Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Joins("JOIN user_roles ON user_roles.role_id = role_permissions.role_id").
		Where("user_roles.user_id = ?", userID).
		Distinct().
		Scan(&codes).Error
	if err != nil {
		return nil, err
	}
	return codes, nil
}
