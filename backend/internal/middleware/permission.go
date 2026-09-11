package middleware

import (
	"lims-backend/internal/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// PermissionMiddleware checks if the current user has the required permission code.
// Pass one or more permission codes — user needs at least one to pass.
// Permissions are read directly from the JWT claims (no DB query).
//
// Admin users bypass this middleware (they are expected to hold every business
// permission). DeptScopeMiddleware still limits admins to their own department
// so CNAS duty-separation rules are preserved.
func PermissionMiddleware(db *gorm.DB, codes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if len(codes) == 0 {
			c.Next()
			return
		}

		if IsAdmin(c) {
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
