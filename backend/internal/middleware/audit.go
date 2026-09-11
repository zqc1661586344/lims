package middleware

import (
	"context"
	"encoding/json"
	"lims-backend/internal/model"
	"reflect"
	"strings"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// normalizeJSON ensures a value written to a jsonb column is valid JSON.
// An empty string is not valid JSON in PostgreSQL — use "null" instead.
func normalizeJSON(s string) string {
	if strings.TrimSpace(s) == "" {
		return "null"
	}
	return s
}

// AuditPlugin implements GORM's plugin interface to automatically
// record data changes (CREATE/UPDATE/DELETE) in the audit_logs table.
type AuditPlugin struct {
	logger *zap.Logger
}

// NewAuditPlugin creates a new AuditPlugin.
func NewAuditPlugin(logger *zap.Logger) *AuditPlugin {
	return &AuditPlugin{logger: logger}
}

// Name returns the plugin name.
func (p *AuditPlugin) Name() string {
	return "audit:logger"
}

// Initialize registers GORM callbacks for Create, Update, and Delete.
func (p *AuditPlugin) Initialize(db *gorm.DB) error {
	// Skip audit logging for the audit_logs table itself.
	skipTables := map[string]bool{
		"audit_logs": true,
	}

	// Register AFTER CREATE callback.
	if err := db.Callback().Create().After("gorm:after_create").Register("audit:after_create", p.afterCreate(skipTables)); err != nil {
		return err
	}

	// Register BEFORE UPDATE callback to capture old data.
	if err := db.Callback().Update().Before("gorm:before_update").Register("audit:before_update", p.beforeUpdateCacheOld(skipTables)); err != nil {
		return err
	}
	// Register AFTER UPDATE callback.
	if err := db.Callback().Update().After("gorm:after_update").Register("audit:after_update", p.afterUpdate(skipTables)); err != nil {
		return err
	}

	// Register AFTER DELETE callback.
	if err := db.Callback().Delete().After("gorm:after_delete").Register("audit:after_delete", p.afterDelete(skipTables)); err != nil {
		return err
	}

	return nil
}

// afterCreate returns a callback function that logs record creation.
func (p *AuditPlugin) afterCreate(skipTables map[string]bool) func(db *gorm.DB) {
	return func(db *gorm.DB) {
		if db.Statement == nil || db.Statement.Table == "" {
			return
		}
		if skipTables[db.Statement.Table] {
			return
		}

		// Skip if no rows were affected.
		if db.Statement.RowsAffected == 0 {
			return
		}

		operatorID, operator := extractOperator(db)
		recordID := extractRecordID(db)

		newData, _ := json.Marshal(db.Statement.Dest)

		entry := model.AuditLog{
			AffectedTable: db.Statement.Table,
			RecordID:      recordID,
			Action:        "CREATE",
			OperatorID:    operatorID,
			Operator:      operator,
			OldData:       "null", // CREATE has no previous state; jsonb requires valid JSON
			NewData:       normalizeJSON(string(newData)),
		}

		if err := db.Session(&gorm.Session{NewDB: true, SkipHooks: true}).Create(&entry).Error; err != nil {
			p.logger.Warn("Failed to write audit log",
				zap.String("table", db.Statement.Table),
				zap.Error(err),
			)
		}
	}
}

// afterUpdate returns a callback function that logs record updates.
func (p *AuditPlugin) afterUpdate(skipTables map[string]bool) func(db *gorm.DB) {
	return func(db *gorm.DB) {
		if db.Statement == nil || db.Statement.Table == "" {
			return
		}
		if skipTables[db.Statement.Table] {
			return
		}
		if db.Statement.RowsAffected == 0 {
			return
		}

		operatorID, operator := extractOperator(db)
		recordID := extractRecordID(db)

		oldData := []byte("null")
		if db.Statement.Context != nil {
			if cached, ok := db.Statement.Context.Value(auditOldDataKey).([]byte); ok && cached != nil {
				oldData = cached
			}
		}

		newData, _ := json.Marshal(db.Statement.Dest)

		entry := model.AuditLog{
			AffectedTable: db.Statement.Table,
			RecordID:      recordID,
			Action:        "UPDATE",
			OperatorID:    operatorID,
			Operator:      operator,
			OldData:       normalizeJSON(string(oldData)),
			NewData:       normalizeJSON(string(newData)),
		}

		if err := db.Session(&gorm.Session{NewDB: true, SkipHooks: true}).Create(&entry).Error; err != nil {
			p.logger.Warn("Failed to write audit log",
				zap.String("table", db.Statement.Table),
				zap.Error(err),
			)
		}
	}
}

// afterDelete returns a callback function that logs record deletion.
func (p *AuditPlugin) afterDelete(skipTables map[string]bool) func(db *gorm.DB) {
	return func(db *gorm.DB) {
		if db.Statement == nil || db.Statement.Table == "" {
			return
		}
		if skipTables[db.Statement.Table] {
			return
		}
		if db.Statement.RowsAffected == 0 {
			return
		}

		operatorID, operator := extractOperator(db)
		recordID := extractRecordID(db)

		// Capture the deleted data.
		var oldData []byte
		if db.Statement.Model != nil && recordID > 0 {
			oldData, _ = json.Marshal(db.Statement.Model)
		} else if db.Statement.Dest != nil {
			oldData, _ = json.Marshal(db.Statement.Dest)
		}

		entry := model.AuditLog{
			AffectedTable: db.Statement.Table,
			RecordID:      recordID,
			Action:        "DELETE",
			OperatorID:    operatorID,
			Operator:      operator,
			OldData:       normalizeJSON(string(oldData)),
			NewData:       "null", // DELETE has no new state; jsonb requires valid JSON
		}

		if err := db.Session(&gorm.Session{NewDB: true, SkipHooks: true}).Create(&entry).Error; err != nil {
			p.logger.Warn("Failed to write audit log",
				zap.String("table", db.Statement.Table),
				zap.Error(err),
			)
		}
	}
}

// extractOperator reads the operator info from the gorm statement context.
// The auth middleware stores user info in the request context, which is
// carried through to GORM via db.Statement.Context.
func extractOperator(db *gorm.DB) (uint, string) {
	if db.Statement.Context == nil {
		return 0, "system"
	}

	type ctxKey string
	const (
		keyUserID   ctxKey = "user_id"
		keyUsername ctxKey = "username"
	)

	uid, ok := db.Statement.Context.Value(keyUserID).(uint)
	if !ok {
		return 0, "system"
	}

	uname, _ := db.Statement.Context.Value(keyUsername).(string)
	return uid, uname
}

// extractRecordID extracts the primary key value from the gorm statement.
func extractRecordID(db *gorm.DB) uint {
	// Try primary key from clause.
	if db.Statement.Schema != nil && db.Statement.Schema.PrioritizedPrimaryField != nil {
		if field := db.Statement.ReflectValue; field.IsValid() && field.Kind() == reflect.Struct {
			v := field.FieldByName(db.Statement.Schema.PrioritizedPrimaryField.Name)
			if v.IsValid() && v.CanUint() {
				return uint(v.Uint())
			}
		}
	}

	// Fallback: check if dest is a struct with ID field.
	if db.Statement.Dest != nil {
		data, _ := json.Marshal(db.Statement.Dest)
		var m map[string]interface{}
		if err := json.Unmarshal(data, &m); err == nil {
			if id, ok := m["id"]; ok {
				switch v := id.(type) {
				case float64:
					return uint(v)
				}
			}
		}
	}

	return 0
}

func cloneModel(m interface{}) interface{} {
	if m == nil {
		return nil
	}
	v := reflect.ValueOf(m)
	if v.Kind() == reflect.Ptr {
		newPtr := reflect.New(v.Elem().Type())
		newPtr.Elem().Set(v.Elem())
		return newPtr.Interface()
	}
	return m
}

type auditCtxKey string

const auditOldDataKey auditCtxKey = "audit_old_data"

func (p *AuditPlugin) beforeUpdateCacheOld(skipTables map[string]bool) func(db *gorm.DB) {
	return func(db *gorm.DB) {
		if db.Statement == nil || db.Statement.Table == "" {
			return
		}
		if skipTables[db.Statement.Table] {
			return
		}
		if db.Statement.Model == nil {
			return
		}
		recordID := extractRecordID(db)
		if recordID == 0 {
			return
		}
		oldModel := cloneModel(db.Statement.Model)
		if err := db.Session(&gorm.Session{NewDB: true, SkipHooks: true}).
			Model(oldModel).
			Where("id = ?", recordID).
			First(oldModel).Error; err != nil {
			return
		}
		oldData, _ := json.Marshal(oldModel)
		if db.Statement.Context != nil {
			ctx := context.WithValue(db.Statement.Context, auditOldDataKey, oldData)
			db.Statement.Context = ctx
		}
	}
}
