package routes

import (
	"lims-backend/internal/config"
	"lims-backend/internal/handler"
	"lims-backend/internal/middleware"
	"lims-backend/internal/workflow"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func RegisterBusinessRoutes(r *gin.RouterGroup, cfg *config.Config, logger *zap.Logger, db *gorm.DB) {
	taskOrderH := handler.NewTaskOrderHandler(logger, db)
	contractReviewH := handler.NewContractReviewHandler(logger, db)
	qcTaskH := handler.NewQCTaskHandler(logger, db)
	samplingScheduleH := handler.NewSamplingScheduleHandler(logger, db)
	fieldSamplingH := handler.NewFieldSamplingRecordHandler(logger, db)
	sampleReceivingH := handler.NewSampleReceivingHandler(logger, db)
	taskAssignH := handler.NewTaskAssignHandler(logger, db)
	dataEntryH := handler.NewDataEntryHandler(logger, db)
	dataReviewH := handler.NewDataReviewHandler(logger, db)
	dataAuditH := handler.NewDataAuditHandler(logger, db)
	reportPrepareH := handler.NewReportPrepareHandler(logger, db)
	reportReviewH := handler.NewReportReviewHandler(logger, db)
	reportAuditH := handler.NewReportAuditHandler(logger, db)
	reportSignH := handler.NewReportSignHandler(logger, db)
	reportPrintH := handler.NewReportPrintHandler(logger, db)
	projectArchiveH := handler.NewProjectArchiveHandler(logger, db)

	group := r.Group("/business")
	group.Use(middleware.AuthMiddleware(cfg, db))
	group.Use(middleware.GormContextMiddleware(db))
	{
		g := group.Group("/task-orders")
		g.Use(middleware.PermissionMiddleware(db, "business:task-order"))
		{
			g.GET("", taskOrderH.List)
			g.POST("", taskOrderH.Create)
			g.GET("/:id", taskOrderH.Get)
			g.PUT("/:id", taskOrderH.Update)
			g.DELETE("/:id", taskOrderH.Delete)

			deptG := g.Group("")
			deptG.Use(middleware.DeptScopeMiddleware(db, workflow.DeptBusiness))
			deptG.POST("/:id/submit", taskOrderH.Submit)
		}

		g = group.Group("/contract-reviews")
		g.Use(middleware.PermissionMiddleware(db, "business:contract-review"))
		{
			g.GET("", contractReviewH.List)
			g.POST("", contractReviewH.Create)
			g.GET("/:id", contractReviewH.Get)
			g.PUT("/:id", contractReviewH.Update)
			g.DELETE("/:id", contractReviewH.Delete)

			deptG := g.Group("")
			deptG.Use(middleware.DeptScopeMiddleware(db, workflow.DeptTech))
			deptG.POST("/:id/approve", contractReviewH.Approve)
			deptG.POST("/:id/reject", contractReviewH.Reject)
		}

		g = group.Group("/qc-tasks")
		g.Use(middleware.PermissionMiddleware(db, "business:qc-task"))
		{
			g.GET("", qcTaskH.List)
			g.POST("", qcTaskH.Create)
			g.GET("/:id", qcTaskH.Get)
			g.PUT("/:id", qcTaskH.Update)
			g.DELETE("/:id", qcTaskH.Delete)

			deptG := g.Group("")
			deptG.Use(middleware.DeptScopeMiddleware(db, workflow.DeptQC))
			deptG.POST("/:id/approve", qcTaskH.Approve)
			deptG.POST("/:id/reject", qcTaskH.Reject)
		}

		g = group.Group("/sampling-schedules")
		g.Use(middleware.PermissionMiddleware(db, "business:sampling-schedule"))
		{
			g.GET("", samplingScheduleH.List)
			g.POST("", samplingScheduleH.Create)
			g.GET("/:id", samplingScheduleH.Get)
			g.PUT("/:id", samplingScheduleH.Update)
			g.DELETE("/:id", samplingScheduleH.Delete)

			deptG := g.Group("")
			deptG.Use(middleware.DeptScopeMiddleware(db, workflow.DeptField))
			deptG.POST("/:id/approve", samplingScheduleH.Approve)
			deptG.POST("/:id/reject", samplingScheduleH.Reject)
		}

		g = group.Group("/field-sampling")
		g.Use(middleware.PermissionMiddleware(db, "business:field-sampling"))
		{
			g.GET("", fieldSamplingH.List)
			g.POST("", fieldSamplingH.Create)
			g.GET("/:id", fieldSamplingH.Get)
			g.PUT("/:id", fieldSamplingH.Update)
			g.DELETE("/:id", fieldSamplingH.Delete)

			deptG := g.Group("")
			deptG.Use(middleware.DeptScopeMiddleware(db, workflow.DeptField))
			deptG.POST("/:id/approve", fieldSamplingH.Approve)
			deptG.POST("/:id/reject", fieldSamplingH.Reject)
		}

		g = group.Group("/sample-receiving")
		g.Use(middleware.PermissionMiddleware(db, "business:sample-receiving"))
		{
			g.GET("", sampleReceivingH.List)
			g.POST("", sampleReceivingH.Create)
			g.GET("/:id", sampleReceivingH.Get)
			g.PUT("/:id", sampleReceivingH.Update)
			g.DELETE("/:id", sampleReceivingH.Delete)

			deptG := g.Group("")
			deptG.Use(middleware.DeptScopeMiddleware(db, workflow.DeptSample))
			deptG.POST("/:id/approve", sampleReceivingH.Approve)
			deptG.POST("/:id/reject", sampleReceivingH.Reject)
		}

		g = group.Group("/task-assign")
		g.Use(middleware.PermissionMiddleware(db, "business:task-assign"))
		{
			g.GET("", taskAssignH.List)
			g.POST("", taskAssignH.Create)
			g.GET("/:id", taskAssignH.Get)
			g.PUT("/:id", taskAssignH.Update)
			g.DELETE("/:id", taskAssignH.Delete)

			deptG := g.Group("")
			deptG.Use(middleware.DeptScopeMiddleware(db, workflow.DeptLab))
			deptG.POST("/:id/approve", taskAssignH.Approve)
			deptG.POST("/:id/reject", taskAssignH.Reject)
		}

		g = group.Group("/data-entry")
		g.Use(middleware.PermissionMiddleware(db, "business:data-entry"))
		{
			g.GET("", dataEntryH.List)
			g.POST("", dataEntryH.Create)
			g.GET("/:id", dataEntryH.Get)
			g.PUT("/:id", dataEntryH.Update)
			g.DELETE("/:id", dataEntryH.Delete)

			deptG := g.Group("")
			deptG.Use(middleware.DeptScopeMiddleware(db, workflow.DeptLab))
			deptG.POST("/:id/approve", dataEntryH.Approve)
			deptG.POST("/:id/reject", dataEntryH.Reject)
		}

		g = group.Group("/data-review")
		g.Use(middleware.PermissionMiddleware(db, "business:data-review"))
		{
			g.GET("", dataReviewH.List)
			g.POST("", dataReviewH.Create)
			g.GET("/:id", dataReviewH.Get)
			g.PUT("/:id", dataReviewH.Update)
			g.DELETE("/:id", dataReviewH.Delete)

			deptG := g.Group("")
			deptG.Use(middleware.DeptScopeMiddleware(db, workflow.DeptLab))
			deptG.POST("/:id/approve", dataReviewH.Approve)
			deptG.POST("/:id/reject", dataReviewH.Reject)
		}

		g = group.Group("/data-audit")
		g.Use(middleware.PermissionMiddleware(db, "business:data-audit"))
		{
			g.GET("", dataAuditH.List)
			g.POST("", dataAuditH.Create)
			g.GET("/:id", dataAuditH.Get)
			g.PUT("/:id", dataAuditH.Update)
			g.DELETE("/:id", dataAuditH.Delete)

			deptG := g.Group("")
			deptG.Use(middleware.DeptScopeMiddleware(db, workflow.DeptLab))
			deptG.POST("/:id/approve", dataAuditH.Approve)
			deptG.POST("/:id/reject", dataAuditH.Reject)
		}

		g = group.Group("/report-prepare")
		g.Use(middleware.PermissionMiddleware(db, "business:report-prepare"))
		{
			g.GET("", reportPrepareH.List)
			g.POST("", reportPrepareH.Create)
			g.GET("/:id", reportPrepareH.Get)
			g.PUT("/:id", reportPrepareH.Update)
			g.DELETE("/:id", reportPrepareH.Delete)

			deptG := g.Group("")
			deptG.Use(middleware.DeptScopeMiddleware(db, workflow.DeptReport))
			deptG.POST("/:id/approve", reportPrepareH.Approve)
			deptG.POST("/:id/reject", reportPrepareH.Reject)
		}

		g = group.Group("/report-review")
		g.Use(middleware.PermissionMiddleware(db, "business:report-review"))
		{
			g.GET("", reportReviewH.List)
			g.POST("", reportReviewH.Create)
			g.GET("/:id", reportReviewH.Get)
			g.PUT("/:id", reportReviewH.Update)
			g.DELETE("/:id", reportReviewH.Delete)

			deptG := g.Group("")
			deptG.Use(middleware.DeptScopeMiddleware(db, workflow.DeptLab))
			deptG.POST("/:id/approve", reportReviewH.Approve)
			deptG.POST("/:id/reject", reportReviewH.Reject)
		}

		g = group.Group("/report-audit")
		g.Use(middleware.PermissionMiddleware(db, "business:report-audit"))
		{
			g.GET("", reportAuditH.List)
			g.POST("", reportAuditH.Create)
			g.GET("/:id", reportAuditH.Get)
			g.PUT("/:id", reportAuditH.Update)
			g.DELETE("/:id", reportAuditH.Delete)

			deptG := g.Group("")
			deptG.Use(middleware.DeptScopeMiddleware(db, workflow.DeptQC))
			deptG.POST("/:id/approve", reportAuditH.Approve)
			deptG.POST("/:id/reject", reportAuditH.Reject)
		}

		g = group.Group("/report-sign")
		g.Use(middleware.PermissionMiddleware(db, "business:report-sign"))
		{
			g.GET("", reportSignH.List)
			g.POST("", reportSignH.Create)
			g.GET("/:id", reportSignH.Get)
			g.PUT("/:id", reportSignH.Update)
			g.DELETE("/:id", reportSignH.Delete)

			deptG := g.Group("")
			deptG.Use(middleware.DeptScopeMiddleware(db, workflow.DeptTech))
			deptG.POST("/:id/approve", reportSignH.Approve)
			deptG.POST("/:id/reject", reportSignH.Reject)
		}

		g = group.Group("/report-print")
		g.Use(middleware.PermissionMiddleware(db, "business:report-print"))
		{
			g.GET("", reportPrintH.List)
			g.POST("", reportPrintH.Create)
			g.GET("/:id", reportPrintH.Get)
			g.PUT("/:id", reportPrintH.Update)
			g.DELETE("/:id", reportPrintH.Delete)

			deptG := g.Group("")
			deptG.Use(middleware.DeptScopeMiddleware(db, workflow.DeptBusiness))
			deptG.POST("/:id/approve", reportPrintH.Approve)
			deptG.POST("/:id/reject", reportPrintH.Reject)
		}

		g = group.Group("/project-archive")
		g.Use(middleware.PermissionMiddleware(db, "business:project-archive"))
		{
			g.GET("", projectArchiveH.List)
			g.POST("", projectArchiveH.Create)
			g.GET("/:id", projectArchiveH.Get)
			g.PUT("/:id", projectArchiveH.Update)
			g.DELETE("/:id", projectArchiveH.Delete)

			deptG := g.Group("")
			deptG.Use(middleware.DeptScopeMiddleware(db, workflow.DeptReport))
			deptG.POST("/:id/approve", projectArchiveH.Approve)
			deptG.POST("/:id/reject", projectArchiveH.Reject)
		}
	}
}
