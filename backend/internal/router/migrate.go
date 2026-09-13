package router

import (
	"lims-backend/internal/model"
	"os"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

func RunAutoMigrate(logger *zap.Logger, db *gorm.DB) {
	logger.Info("running AutoMigrate to create/update tables")

	shouldDelete := os.Getenv("LIMS_CLEANUP_ORPHANS") == "true"
	if shouldDelete {
		logger.Warn("orphan row cleanup ENABLED — this will DELETE data; set LIMS_CLEANUP_ORPHANS=false to skip")
	} else {
		logger.Info("orphan row cleanup is disabled (set LIMS_CLEANUP_ORPHANS=true to enable)")
	}
	preMigrateCleanup(db, logger, shouldDelete)

	db.AutoMigrate(
		&model.User{},
		&model.Dept{},
		&model.Role{},
		&model.Permission{},
		&model.UserRole{},
		&model.RolePermission{},
		&model.AuditLog{},
		&model.TestItem{},
		&model.TestStandard{},
		&model.Equipment{},
		&model.Reagent{},
		&model.ProcessInstance{},
		&model.ProcessTask{},
		&model.TaskOrder{},
		&model.ContractReview{},
		&model.QCTask{},
		&model.SamplingSchedule{},
		&model.FieldSamplingRecord{},
		&model.SampleReceiving{},
		&model.TaskAssign{},
		&model.DataEntry{},
		&model.DataReview{},
		&model.DataAudit{},
		&model.ReportPrepare{},
		&model.ReportReview{},
		&model.ReportAudit{},
		&model.ReportSign{},
		&model.ReportPrint{},
		&model.ProjectArchive{},
		&model.LabSheetTemplate{},
		&model.LabSheet{},
		&model.SamplingSheetTemplate{},
		&model.SamplingSheet{},
		&model.Sample{},
		&model.TaskOrderTestItem{},
		&model.SamplingPoint{},
	)
}
