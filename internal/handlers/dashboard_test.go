package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/babykart/gozone/internal/middleware"
	"github.com/babykart/gozone/internal/models"
)

func TestDashboard(t *testing.T) {
	h := newTestHandler(t)

	user := &models.User{ID: 1, Username: "admin", Role: "admin"}
	ctx := context.WithValue(context.Background(), middleware.UserContextKey, user)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	r = r.WithContext(ctx)
	h.Dashboard(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestGetRecentActivityLogs_Empty(t *testing.T) {
	h := newTestHandler(t)

	logs := h.getRecentActivityLogs(10)
	if len(logs) != 0 {
		t.Errorf("expected 0 logs, got %d", len(logs))
	}
}
