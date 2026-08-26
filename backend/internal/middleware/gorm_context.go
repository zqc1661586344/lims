package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ctxKey string

const (
	ctxUserID   ctxKey = "user_id"
	ctxUsername ctxKey = "username"
)

// GormContextMiddleware propagates Gin context values (user_id, username)
// into the GORM DB session context so audit hooks can extract the operator.
func GormContextMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		username, _ := c.Get("username")

		uid, _ := userID.(uint)
		uname, _ := username.(string)

		ctx := context.WithValue(c.Request.Context(), ctxUserID, uid)
		ctx = context.WithValue(ctx, ctxUsername, uname)

		// Replace the request context and store a contextualized DB in Gin.
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