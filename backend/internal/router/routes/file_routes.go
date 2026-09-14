package routes

import (
	"lims-backend/internal/config"
	"lims-backend/internal/handler"
	"lims-backend/internal/middleware"
	"lims-backend/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func RegisterFileRoutes(r *gin.RouterGroup, cfg *config.Config, logger *zap.Logger, db *gorm.DB, storage service.StorageService) {
	fh := handler.NewFileHandler(db, logger, storage)

	files := r.Group("/files")
	files.Use(middleware.AuthMiddleware(cfg, db))
	{
		files.POST("/upload", fh.Upload)
		files.GET("", fh.List)
		files.GET("/:id", fh.Get)
		files.GET("/:id/download", fh.Download)
		files.DELETE("/:id", fh.Delete)
	}
}
