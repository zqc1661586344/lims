package middleware

import (
	"net/http"
	"strings"

	"lims-backend/internal/config"
	"lims-backend/internal/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AuthMiddleware returns a Gin middleware that validates JWT tokens AND
// refreshes user state (status, is_admin, dept_id, permissions) from the
// database on every request. This ensures that disabling a user or removing
// a permission takes effect immediately — stale values in the JWT claims
// are never trusted.
func AuthMiddleware(cfg *config.Config, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.Unauthorized(c, "missing authorization header")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			utils.Unauthorized(c, "invalid authorization format")
			c.Abort()
			return
		}

		claims, err := utils.ParseToken(&cfg.JWT, parts[1])
		if err != nil {
			utils.Unauthorized(c, "invalid or expired token")
			c.Abort()
			return
		}

		if utils.GetTokenBlacklist().IsRevoked(claims.ID) {
			utils.Unauthorized(c, "token has been revoked")
			c.Abort()
			return
		}

		if db != nil {
			var row struct {
				ID      uint
				Status  int
				IsAdmin bool
				DeptID  *uint
			}
			if err := db.Raw(
				"SELECT id, status, is_admin, dept_id FROM users WHERE id = ?",
				claims.UserID,
			).Scan(&row).Error; err != nil || row.ID == 0 {
				utils.Unauthorized(c, "用户不存在")
				c.Abort()
				return
			}
			if row.Status == 0 {
				utils.Unauthorized(c, "账号已被禁用")
				c.Abort()
				return
			}

			claims.IsAdmin = row.IsAdmin
			claims.DeptID = row.DeptID

			var perms []string
			if row.IsAdmin {
				if err := db.Raw("SELECT code FROM permissions").Scan(&perms).Error; err != nil {
					utils.InternalError(c, "加载权限失败")
					c.Abort()
					return
				}
			} else {
				if err := db.Raw(`
					SELECT DISTINCT p.code
					FROM permissions p
					JOIN role_permissions rp ON rp.permission_id = p.id
					JOIN user_roles ur ON ur.role_id = rp.role_id
					WHERE ur.user_id = ?
				`, claims.UserID).Scan(&perms).Error; err != nil {
					utils.InternalError(c, "加载权限失败")
					c.Abort()
					return
				}
			}
			claims.Permissions = perms
		}

		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("dept_id", claims.DeptID)
		c.Set("is_admin", claims.IsAdmin)
		c.Set("permissions", claims.Permissions)
		c.Set("jti", claims.ID)
		if claims.ExpiresAt != nil {
			c.Set("exp", claims.ExpiresAt.Time)
		}

		c.Next()
	}
}

// GetUserID extracts the user ID from the Gin context.
func GetUserID(c *gin.Context) uint {
	id, _ := c.Get("user_id")
	if id == nil {
		return 0
	}
	return id.(uint)
}

// GetDeptID extracts the department ID from the Gin context.
func GetDeptID(c *gin.Context) *uint {
	deptID, _ := c.Get("dept_id")
	if deptID == nil {
		return nil
	}
	return deptID.(*uint)
}

// IsAdmin checks if the current user is an admin.
func IsAdmin(c *gin.Context) bool {
	admin, _ := c.Get("is_admin")
	if admin == nil {
		return false
	}
	return admin.(bool)
}

// RequireAdmin returns a middleware that enforces the caller to be an admin user.
// It must be placed after AuthMiddleware so that user info is available in context.
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !IsAdmin(c) {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "需要管理员权限才能执行此操作",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// GetPermissions extracts the permission codes from the Gin context.
func GetPermissions(c *gin.Context) []string {
	perms, _ := c.Get("permissions")
	if perms == nil {
		return nil
	}
	return perms.([]string)
}

func GetDeptIDVal(c *gin.Context) uint {
	d := GetDeptID(c)
	if d == nil {
		return 0
	}
	return *d
}
