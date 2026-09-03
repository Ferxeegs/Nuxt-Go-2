package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func newTestRouter(setContext func(c *gin.Context), guard gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/protected", func(c *gin.Context) {
		setContext(c)
		c.Next()
	}, guard, func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	return r
}

func TestRequirePermission_Allowed(t *testing.T) {
	r := newTestRouter(func(c *gin.Context) {
		c.Set(ContextPermissions, []string{"setting.update", "setting.view"})
	}, RequirePermission("setting.update"))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestRequirePermission_Denied(t *testing.T) {
	r := newTestRouter(func(c *gin.Context) {
		c.Set(ContextPermissions, []string{"setting.view"})
	}, RequirePermission("setting.delete"))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestRequirePermission_MissingContext(t *testing.T) {
	r := newTestRouter(func(c *gin.Context) {
		// no permissions set in context, simulating a bypass of Auth middleware
	}, RequirePermission("setting.view"))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestRequireRole_Allowed(t *testing.T) {
	r := newTestRouter(func(c *gin.Context) {
		c.Set(ContextRole, "SuperAdmin")
	}, RequireRole("SuperAdmin", "Admin"))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestRequireRole_Denied(t *testing.T) {
	r := newTestRouter(func(c *gin.Context) {
		c.Set(ContextRole, "Editor")
	}, RequireRole("SuperAdmin", "Admin"))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}
