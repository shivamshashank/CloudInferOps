package ui

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOverviewHandlerReturnsSummary(t *testing.T) {
	svc := NewService()
	h := NewHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/overview", nil)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	if rr.Body.String() == "" {
		t.Fatal("expected response body")
	}
}

func TestDeploymentsHandlerReturnsCollections(t *testing.T) {
	svc := NewService()
	h := NewHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/deployments", nil)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	if rr.Body.String() == "" {
		t.Fatal("expected deployments response")
	}
}

func TestHealthHandlerReturnsOk(t *testing.T) {
	svc := NewService()
	h := NewHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestLogsHandlerReturnsLogCollection(t *testing.T) {
	svc := NewService()
	h := NewHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/logs?service=cloudinferops-ui", nil)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestDeployActionHandlerRequiresTarget(t *testing.T) {
	svc := NewService()
	h := NewHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/api/actions/deploy", nil)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}
