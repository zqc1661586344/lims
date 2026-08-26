package middleware

import (
	"strings"

	"lims-backend/internal/config"
	"lims-backend/internal/utils"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware returns a Gin middleware that validates JWT tokens.
func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.Unauthorized(c, "missing authorization header")
			c.Abort()
			return
		}

		// Expect "Bearer <token>"
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

		// Store user info in context
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("dept_id", claims.DeptID)
		c.Set("is_admin", claims.IsAdmin)

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