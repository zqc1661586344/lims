package middleware

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupSQLiteForDeptTest(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.Exec(`CREATE TABLE depts (id INTEGER PRIMARY KEY AUTOINCREMENT, code TEXT, name TEXT)`).Error)
	require.NoError(t, db.Exec(`INSERT INTO depts (code, name) VALUES ('business', '业务室')`).Error)
	require.NoError(t, db.Exec(`INSERT INTO depts (code, name) VALUES ('lab', '实验室')`).Error)
	require.NoError(t, db.Exec(`INSERT INTO depts (code, name) VALUES ('tech', '技术室')`).Error)
	return db
}

func ptrU(v uint) *uint { return &v }

func TestDeptScopeMiddleware_AdminBypasses(t *testing.T) {
	db := setupSQLiteForDeptTest(t)

	c, _ := newTestCtx()
	c.Set("is_admin", true)
	c.Set("dept_id", ptrU(999))
	DeptScopeMiddleware(db, "business")(c)
	assert.False(t, c.IsAborted(), "admin must bypass department scope")
}

func TestDeptScopeMiddleware_NonAdminMatchingDept(t *testing.T) {
	db := setupSQLiteForDeptTest(t)

	var deptID uint
	require.NoError(t, db.Raw(`SELECT id FROM depts WHERE code = 'lab'`).Scan(&deptID).Error)

	c, w := newTestCtx()
	c.Set("is_admin", false)
	c.Set("dept_id", ptrU(deptID))
	DeptScopeMiddleware(db, "lab", "business")(c)
	assert.False(t, c.IsAborted())
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDeptScopeMiddleware_NonAdminDisallowedDept(t *testing.T) {
	db := setupSQLiteForDeptTest(t)

	var deptID uint
	require.NoError(t, db.Raw(`SELECT id FROM depts WHERE code = 'tech'`).Scan(&deptID).Error)

	c, w := newTestCtx()
	c.Set("is_admin", false)
	c.Set("dept_id", ptrU(deptID))
	DeptScopeMiddleware(db, "lab", "business")(c)
	assert.True(t, c.IsAborted())
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestDeptScopeMiddleware_NoDeptIDReturns403(t *testing.T) {
	db := setupSQLiteForDeptTest(t)

	c, w := newTestCtx()
	c.Set("is_admin", false)
	c.Set("dept_id", nil)
	DeptScopeMiddleware(db, "lab")(c)
	assert.True(t, c.IsAborted())
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestDeptScopeMiddleware_EmptyAllowedCodesAllowsAdmin(t *testing.T) {
	db := setupSQLiteForDeptTest(t)
	c, _ := newTestCtx()
	c.Set("is_admin", true)
	c.Set("dept_id", ptrU(1))
	DeptScopeMiddleware(db)(c)
	assert.False(t, c.IsAborted())
}

func TestDeptScopeMiddleware_EmptyAllowedCodesBlocksNonAdmin(t *testing.T) {
	db := setupSQLiteForDeptTest(t)
	var deptID uint
	require.NoError(t, db.Raw(`SELECT id FROM depts WHERE code = 'lab'`).Scan(&deptID).Error)

	c, w := newTestCtx()
	c.Set("is_admin", false)
	c.Set("dept_id", ptrU(deptID))
	DeptScopeMiddleware(db)(c)
	assert.True(t, c.IsAborted(), "non-admin with no allowed codes should be blocked")
	assert.Equal(t, http.StatusForbidden, w.Code)
}
