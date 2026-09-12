package router

import (
	"lims-backend/internal/model"
	"os"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

func RunAutoMigrate(logger *zap.Logger, db *gorm.DB) {
	logger.Info("running AutoMigrate to create/update tables")

	strictCleanup := os.Getenv("LIMS_STRICT_CLEANUP") != "false"
	preMigrateCleanup(db, logger, strictCleanup)

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
	)
}
