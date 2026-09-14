package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func newTestCtx() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	return c, w
}

func setAdminCtx(c *gin.Context, admin bool) {
	c.Set("is_admin", admin)
	c.Set("permissions", []string{})
}

func setPermCtx(c *gin.Context, perms []string) {
	c.Set("is_admin", false)
	c.Set("permissions", perms)
}

func testOKHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func TestPermissionMiddleware_EmptyCodesAllowsAll(t *testing.T) {
	c, w := newTestCtx()
	setPermCtx(c, nil)
	PermissionMiddleware(nil)(c)
	assert.False(t, c.IsAborted())
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestPermissionMiddleware_AdminBypassesAllChecks(t *testing.T) {
	for _, codes := range [][]string{
		{"workflow:task"},
		{"business:task-order", "workflow:view"},
		{"totally:nonexistent:permission"},
	} {
		t.Run("admin bypasses "+codes[0], func(t *testing.T) {
			c, _ := newTestCtx()
			setAdminCtx(c, true)
			mw := PermissionMiddleware(nil, codes...)
			mw(c)
			assert.False(t, c.IsAborted(), "admin must never be blocked")
		})
	}
}

func TestPermissionMiddleware_NonAdminWithMatchingPerm(t *testing.T) {
	codes := []string{"workflow:view", "workflow:task"}
	for _, userPerm := range codes {
		t.Run("has "+userPerm, func(t *testing.T) {
			c, w := newTestCtx()
			setPermCtx(c, []string{userPerm})
			mw := PermissionMiddleware(nil, codes...)
			mw(c)
			assert.False(t, c.IsAborted())
			assert.Equal(t, http.StatusOK, w.Code)
		})
	}
}

func TestPermissionMiddleware_NonAdminWithoutPermGets403(t *testing.T) {
	c, w := newTestCtx()
	setPermCtx(c, []string{"other:perm"})
	mw := PermissionMiddleware(nil, "workflow:task")
	mw(c)
	assert.True(t, c.IsAborted())
	assert.Equal(t, http.StatusForbidden, w.Code)
	var body map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &body)
	assert.Equal(t, float64(403), body["code"])
}

func TestPermissionMiddleware_NonAdminNoPermsGets403(t *testing.T) {
	c, w := newTestCtx()
	setPermCtx(c, nil)
	mw := PermissionMiddleware(nil, "workflow:view")
	mw(c)
	assert.True(t, c.IsAborted())
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestPermissionMiddleware_NilDBDoesNotPanic(t *testing.T) {
	c, w := newTestCtx()
	setPermCtx(c, []string{"x"})
	assert.NotPanics(t, func() {
		PermissionMiddleware((*gorm.DB)(nil), "x")(c)
	})
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestIsAdmin_FalseWhenNoKey(t *testing.T) {
	c, _ := newTestCtx()
	assert.False(t, IsAdmin(c))
}

func TestIsAdmin_FalseForExplicitFalse(t *testing.T) {
	c, _ := newTestCtx()
	c.Set("is_admin", false)
	assert.False(t, IsAdmin(c))
}

func TestIsAdmin_True(t *testing.T) {
	c, _ := newTestCtx()
	c.Set("is_admin", true)
	assert.True(t, IsAdmin(c))
}
