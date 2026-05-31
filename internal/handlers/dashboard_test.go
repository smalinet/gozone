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

func TestDashboard_ServerStats(t *testing.T) {
	h, pdnsSrv := newTestHandlerWithPDNS(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[{"name":"uptime","type":"StatisticItem","value":"3600"},{"name":"questions","type":"StatisticItem","value":"12345"},{"name":"packetcache-hit","type":"StatisticItem","value":"5000"},{"name":"packetcache-miss","type":"StatisticItem","value":"100"},{"name":"query-cache-hit","type":"StatisticItem","value":"8000"},{"name":"query-cache-miss","type":"StatisticItem","value":"200"},{"name":"qsize-q","type":"StatisticItem","value":"0"}]`))
	})
	defer pdnsSrv.Close()

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
