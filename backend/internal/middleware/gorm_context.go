package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ctxKey string

const (
	CtxUserID   ctxKey = "user_id"
	CtxUsername ctxKey = "username"
	CtxIsAdmin  ctxKey = "is_admin"
)

// GormContextMiddleware propagates Gin context values (user_id, username,
// is_admin) into the GORM DB session context so audit hooks and workflow
// engines can extract the operator & their role.
func GormContextMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		username, _ := c.Get("username")
		isAdmin, _ := c.Get("is_admin")

		uid, _ := userID.(uint)
		uname, _ := username.(string)
		admin, _ := isAdmin.(bool)

		ctx := context.WithValue(c.Request.Context(), CtxUserID, uid)
		ctx = context.WithValue(ctx, CtxUsername, uname)
		ctx = context.WithValue(ctx, CtxIsAdmin, admin)

		c.Request = c.Request.WithContext(ctx)
		c.Set("db", db.WithContext(ctx))

		c.Next()
	}
}

// GetDB retrieves the contextualized GORM DB from the Gin context.
func GetDB(c *gin.Context) *gorm.DB {
	db, _ := c.Get("db")
	if db == nil {
		return nil
	}
	return db.(*gorm.DB)
}
