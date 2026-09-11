package routes

import (
	"lims-backend/internal/config"
	"lims-backend/internal/handler"
	"lims-backend/internal/middleware"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func RegisterSamplingSheetRoutes(r *gin.RouterGroup, cfg *config.Config, logger *zap.Logger, db *gorm.DB) {
	h := handler.NewSamplingSheetHandler(logger, db)
	group := r.Group("/sampling-sheets")
	group.Use(middleware.AuthMiddleware(cfg))
	group.Use(middleware.GormContextMiddleware(db))
	{
		group.GET("/templates", h.ListTemplates)
		group.POST("/templates", h.CreateTemplate)
		group.GET("/templates/:id", h.GetTemplate)
		group.PUT("/templates/:id", h.UpdateTemplate)
		group.DELETE("/templates/:id", h.DeleteTemplate)
		group.GET("/templates/by-sample-type/:sampleType", h.GetTemplateBySampleType)

		group.GET("", h.List)
		group.POST("", h.Create)
		group.GET("/:id", h.Get)
		group.PUT("/:id", h.Update)
		group.DELETE("/:id", h.Delete)
	}
}
