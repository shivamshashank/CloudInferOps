package ui

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func testService(t *testing.T) *Service {
	t.Helper()
	runner := func(_ context.Context, name string, args ...string) ([]byte, error) {
		joined := name + " " + strings.Join(args, " ")
		switch {
		case strings.Contains(joined, "get deployments"):
			return []byte(`{"items":[{"metadata":{"name":"gateway-deployment","namespace":"inference","creationTimestamp":"2026-01-01T00:00:00Z"},"spec":{"replicas":1},"status":{"readyReplicas":1,"availableReplicas":1}}]}`), nil
		case strings.Contains(joined, "get pods"):
			return []byte(`{"items":[{"metadata":{"name":"gateway-1","namespace":"inference"},"status":{"phase":"Running"}}]}`), nil
		case strings.Contains(joined, "logs"):
			return []byte("ready\nhealthy"), nil
		case strings.Contains(joined, "get configmap"):
			return []byte(`{"data":{"provider":"ollama","model":"llama3"}}`), nil
		default:
			return []byte("ok"), nil
		}
	}
	svc := NewServiceWithRunner(runner)
	return svc
}

func request(t *testing.T, h http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	return rr
}

func TestReadEndpoints(t *testing.T) {
	h := NewHandler(testService(t))
	for _, path := range []string{"/api/health", "/api/overview", "/api/deployments", "/api/observability", "/api/pods", "/api/config"} {
		rr := request(t, h, http.MethodGet, path, "")
		if rr.Code != http.StatusOK {
			t.Fatalf("%s: expected 200, got %d", path, rr.Code)
		}
	}
}

func TestLogsHandler(t *testing.T) {
	rr := request(t, NewHandler(testService(t)), http.MethodGet, "/api/logs?namespace=inference&pod=gateway-1", "")
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	var payload LogResponse
	if json.Unmarshal(rr.Body.Bytes(), &payload) != nil || len(payload.Lines) != 2 {
		t.Fatal("expected parsed log lines")
	}
}

func TestActionsDisabledByDefault(t *testing.T) {
	t.Setenv("CLOUDINFEROPS_UI_ENABLE_ACTIONS", "")
	h := NewHandler(testService(t))
	rr := request(t, h, http.MethodPost, "/api/actions/deploy", `{"name":"inference"}`)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
}

func TestInvalidAndUnknownRequests(t *testing.T) {
	h := NewHandler(testService(t))
	if rr := request(t, h, http.MethodPost, "/api/actions/deploy", ""); rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
	if rr := request(t, h, http.MethodGet, "/api/missing", ""); rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestTokenGuard(t *testing.T) {
	t.Setenv("CLOUDINFEROPS_UI_TOKEN", "secret")
	h := NewHandler(testService(t))
	rr := request(t, h, http.MethodGet, "/api/health", "")
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}
