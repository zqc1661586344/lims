package middleware

import (
	"lims-backend/internal/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// PermissionMiddleware checks if the current user has the required permission code.
// Pass one or more permission codes — user needs at least one to pass.
//
// Permission codes are read from the Gin context (populated by AuthMiddleware
// on every request from the DB, so role changes take effect immediately).
//
// Admin users are NOT exempt — they must also hold the permission code.
// AuthMiddleware grants admins every permission code, so admins will always
// pass this check, but DeptScopeMiddleware (department) and workflow SoD rules
// still apply to them.
func PermissionMiddleware(db *gorm.DB, codes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if len(codes) == 0 {
			c.Next()
			return
		}

		userPerms := GetPermissions(c)
		for _, code := range codes {
			for _, p := range userPerms {
				if p == code {
					c.Next()
					return
				}
			}
		}

		utils.Forbidden(c, "无权限访问")
		c.Abort()
	}
}
