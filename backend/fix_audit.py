import sys

fp = "internal/middleware/audit.go"
with open(fp, "r") as f:
    s = f.read()

# 1) Add context import
if '"context"' not in s:
    s = s.replace(
        'import (\n\t"encoding/json"\n\t"reflect"',
        'import (\n\t"context"\n\t"encoding/json"\n\t"reflect"',
    )

# 2) Fix cloneModel + add beforeUpdate method + const
old_clone = """// cloneModel creates a shallow copy of the model pointer.
func cloneModel(m interface{}) interface{} {
	if m == nil {
		return nil
	}
	// Return the same pointer — GORM.Session creates a new DB session so
	// we are reading from a separate query context not mutating the original.
	return m
}"""

new_clone = """func cloneModel(m interface{}) interface{} {
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
}"""

s = s.replace(old_clone, new_clone)

# 3) Register before_update callback in Initialize
old_init = '''// Register AFTER UPDATE callback.
	if err := db.Callback().Update().After("gorm:after_update").Register("audit:after_update"'''
new_init = '''// Register BEFORE UPDATE callback to capture old data.
	if err := db.Callback().Update().Before("gorm:before_update").Register("audit:before_update", p.beforeUpdateCacheOld(skipTables)); err != nil {
		return err
	}
	// Register AFTER UPDATE callback.
	if err := db.Callback().Update().After("gorm:after_update").Register("audit:after_update"'''

s = s.replace(old_init, new_init)

# 4) Fix afterUpdate to read oldData from context instead of re-querying
old_after_update_olddata = """		// Capture old data from the model before updates were applied.
		var oldData []byte
		if db.Statement.Model != nil && recordID > 0 {
			oldModel := cloneModel(db.Statement.Model)
			tx := db.Session(&gorm.Session{NewDB: true, SkipHooks: true}).
				Model(oldModel).
				Where("id = ?", recordID).
				First(oldModel)
			if tx.Error == nil {
				oldData, _ = json.Marshal(oldModel)
			}
		}"""

new_after_update_olddata = """		oldData := []byte("null")
		if db.Statement.Context != nil {
			if cached, ok := db.Statement.Context.Value(auditOldDataKey).([]byte); ok && cached != nil {
				oldData = cached
			}
		}"""

s = s.replace(old_after_update_olddata, new_after_update_olddata)

with open(fp, "w") as f:
    f.write(s)

print("audit.go patched")
