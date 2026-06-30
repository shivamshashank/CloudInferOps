package webhook

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

var (
	recentIncidents []Incident
	incidentsMutex  sync.RWMutex
)

// StartServer registers incident HTTP endpoints and boots the REST gateway daemon
func StartServer(port string) error {
	mux := http.NewServeMux()

	// Probes
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"healthy"}`))
	})

	// Incidents list query
	mux.HandleFunc("/incidents", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		incidentsMutex.RLock()
		data, err := json.Marshal(recentIncidents)
		incidentsMutex.RUnlock()

		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	})

	// Alertmanager Webhook Router
	mux.HandleFunc("/webhook/alertmanager", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var payload AlertmanagerPayload
		err := json.NewDecoder(r.Body).Decode(&payload)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid JSON body: %v", err), http.StatusBadRequest)
			return
		}

		// Parse the alerts and generate Incident instances
		newIncidents := ParseAlertmanagerPayload(payload)

		// Append to in-memory store
		incidentsMutex.Lock()
		recentIncidents = append(recentIncidents, newIncidents...)
		// Cap store size at 100 recent alerts to prevent unbounded memory growth
		if len(recentIncidents) > 100 {
			recentIncidents = recentIncidents[len(recentIncidents)-100:]
		}
		incidentsMutex.Unlock()

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"success","message":"webhook processed successfully"}`))
	})

	fmt.Printf("Incident webhook handler daemon listening on port :%s...\n", port)
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	return server.ListenAndServe()
}
