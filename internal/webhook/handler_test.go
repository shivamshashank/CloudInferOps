package webhook

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestStartServer(t *testing.T) {
	// Start the server in a goroutine on a specific high port
	port := "58223"
	go func() {
		_ = StartServer(port)
	}()

	// Give the server a moment to start
	time.Sleep(200 * time.Millisecond)

	baseURL := "http://localhost:" + port

	// 1. Test /health
	resp, err := http.Get(baseURL + "/health")
	if err != nil {
		t.Fatalf("failed to query /health: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if !strings.Contains(string(body), "healthy") {
		t.Errorf("expected body to contain 'healthy', got %s", body)
	}

	// 2. Test /incidents GET
	resp, err = http.Get(baseURL + "/incidents")
	if err != nil {
		t.Fatalf("failed to query /incidents: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", resp.StatusCode)
	}
	_ = resp.Body.Close()

	// 3. Test /incidents invalid method POST
	resp, err = http.Post(baseURL+"/incidents", "application/json", nil)
	if err != nil {
		t.Fatalf("failed to post /incidents: %v", err)
	}
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("expected 405 Method Not Allowed, got %d", resp.StatusCode)
	}
	_ = resp.Body.Close()

	// 4. Test /webhook/alertmanager invalid method GET
	resp, err = http.Get(baseURL + "/webhook/alertmanager")
	if err != nil {
		t.Fatalf("failed to get /webhook/alertmanager: %v", err)
	}
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("expected 405 Method Not Allowed, got %d", resp.StatusCode)
	}
	_ = resp.Body.Close()

	// 5. Test /webhook/alertmanager valid POST
	payloadJSON := `{
		"status": "firing",
		"alerts": [
			{
				"status": "firing",
				"labels": {"alertname": "TestServerAlert"},
				"annotations": {"summary": "test"},
				"fingerprint": "12345"
			}
		]
	}`
	resp, err = http.Post(baseURL+"/webhook/alertmanager", "application/json", bytes.NewBuffer([]byte(payloadJSON)))
	if err != nil {
		t.Fatalf("failed to post /webhook/alertmanager: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", resp.StatusCode)
	}
	_ = resp.Body.Close()

	// Verify the incident is now stored
	resp, err = http.Get(baseURL + "/incidents")
	if err != nil {
		t.Fatalf("failed to query /incidents: %v", err)
	}
	body, _ = io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if !strings.Contains(string(body), "TestServerAlert") {
		t.Errorf("expected TestServerAlert in incidents, got %s", body)
	}

	// 6. Test /webhook/alertmanager invalid JSON
	resp, err = http.Post(baseURL+"/webhook/alertmanager", "application/json", bytes.NewBuffer([]byte("invalid json")))
	if err != nil {
		t.Fatalf("failed to post invalid JSON: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request, got %d", resp.StatusCode)
	}
	_ = resp.Body.Close()
}
