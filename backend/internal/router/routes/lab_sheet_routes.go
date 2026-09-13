package routes

import (
	"lims-backend/internal/config"
	"lims-backend/internal/handler"
	"lims-backend/internal/middleware"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func RegisterLabSheetRoutes(r *gin.RouterGroup, cfg *config.Config, logger *zap.Logger, db *gorm.DB) {
	h := handler.NewLabSheetHandler(logger, db)
	group := r.Group("/lab-sheets")
	group.Use(middleware.AuthMiddleware(cfg, db))
	group.Use(middleware.GormContextMiddleware(db))
	{
		// Templates
		group.GET("/templates", h.ListTemplates)
		group.POST("/templates", h.CreateTemplate)
		group.GET("/templates/:id", h.GetTemplate)
		group.PUT("/templates/:id", h.UpdateTemplate)
		group.DELETE("/templates/:id", h.DeleteTemplate)
		group.GET("/templates/by-item/:testItemID", h.GetTemplateByItem)

		// Instances
		group.GET("", h.List)
		group.POST("", h.Create)
		group.GET("/:id", h.Get)
		group.PUT("/:id", h.Update)
		group.DELETE("/:id", h.Delete)
	}
}
