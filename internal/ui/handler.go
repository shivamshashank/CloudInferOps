package ui

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type Handler struct {
	service *Service
	token   string
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service, token: strings.TrimSpace(os.Getenv("CLOUDINFEROPS_UI_TOKEN"))}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	if !h.authorized(r) {
		h.writeStatus(w, http.StatusUnauthorized, map[string]string{"status": "error", "message": "authentication required"})
		return
	}
	path := strings.TrimSuffix(r.URL.Path, "/")
	if path == "/api/logs/stream" && r.Method == http.MethodGet {
		h.streamLogs(w, r)
		return
	}
	switch {
	case path == "/api/health" && r.Method == http.MethodGet:
		h.writeJSON(w, h.service.Health())
	case path == "/api/overview" && r.Method == http.MethodGet:
		h.writeJSON(w, h.service.Overview())
	case path == "/api/deployments" && r.Method == http.MethodGet:
		h.writeJSON(w, h.service.Deployments())
	case path == "/api/models" && r.Method == http.MethodGet:
		h.writeJSON(w, h.service.Models())
	case path == "/api/observability" && r.Method == http.MethodGet:
		h.writeJSON(w, h.service.Observability())
	case path == "/api/alerts" && r.Method == http.MethodGet:
		h.writeJSON(w, h.service.Alerts())
	case path == "/api/pods" && r.Method == http.MethodGet:
		h.writeJSON(w, h.service.Pods())
	case path == "/api/logs" && r.Method == http.MethodGet:
		lines, _ := strconv.Atoi(r.URL.Query().Get("lines"))
		payload, err := h.service.Logs(r.URL.Query().Get("namespace"), r.URL.Query().Get("pod"), lines)
		if err != nil {
			h.writeError(w, http.StatusBadRequest, err)
			return
		}
		h.writeJSON(w, payload)
	case path == "/api/config" && r.Method == http.MethodGet:
		h.writeJSON(w, h.service.Config())
	case path == "/api/config" && r.Method == http.MethodPut:
		var payload PortalConfig
		if !h.decode(w, r, &payload) {
			return
		}
		if err := h.service.SaveConfig(payload); err != nil {
			h.writeError(w, http.StatusForbidden, err)
			return
		}
		h.writeJSON(w, map[string]string{"status": "ok", "message": "configuration saved"})
	case path == "/api/benchmark" && r.Method == http.MethodGet:
		h.writeJSON(w, h.service.Benchmark())
	case path == "/api/benchmark" && r.Method == http.MethodPost:
		var payload BenchmarkRequest
		if !h.decode(w, r, &payload) {
			return
		}
		result, err := h.service.RunBenchmark(payload)
		if err != nil {
			h.writeError(w, http.StatusForbidden, err)
			return
		}
		h.writeJSON(w, result)
	case path == "/api/actions/deploy" && r.Method == http.MethodPost:
		var payload struct {
			Name string `json:"name"`
		}
		if !h.decode(w, r, &payload) {
			return
		}
		if err := h.service.Deploy(strings.TrimSpace(payload.Name)); err != nil {
			h.writeError(w, http.StatusForbidden, err)
			return
		}
		h.writeJSON(w, map[string]string{"status": "ok", "message": "deployment reconciled for " + payload.Name})
	case path == "/api/actions/restart" && r.Method == http.MethodPost:
		var payload struct {
			Namespace string `json:"namespace"`
			Name      string `json:"name"`
		}
		if !h.decode(w, r, &payload) {
			return
		}
		if err := h.service.Restart(payload.Namespace, payload.Name); err != nil {
			h.writeError(w, http.StatusForbidden, err)
			return
		}
		h.writeJSON(w, map[string]string{"status": "ok", "message": "restart requested for " + payload.Name})
	default:
		if strings.HasPrefix(path, "/api/") {
			h.writeStatus(w, http.StatusNotFound, map[string]string{"status": "error", "message": "not found"})
			return
		}
		h.writeStatus(w, http.StatusMethodNotAllowed, map[string]string{"status": "error", "message": "method not allowed"})
	}
}

func (h *Handler) streamLogs(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		h.writeStatus(w, http.StatusNotImplemented, map[string]string{"status": "error", "message": "streaming unsupported"})
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()
	writer := flushWriter{writer: w, flusher: flusher}
	_ = h.service.StreamLogs(r.Context(), r.URL.Query().Get("namespace"), r.URL.Query().Get("pod"), writer)
}

type flushWriter struct {
	writer  http.ResponseWriter
	flusher http.Flusher
}

func (w flushWriter) Write(data []byte) (int, error) {
	n, err := w.writer.Write(data)
	w.flusher.Flush()
	return n, err
}

func (h *Handler) authorized(r *http.Request) bool {
	if h.token == "" {
		return true
	}
	token := strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
	return token == h.token
}
func (h *Handler) decode(w http.ResponseWriter, r *http.Request, payload any) bool {
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(payload); err != nil {
		h.writeStatus(w, http.StatusBadRequest, map[string]string{"status": "error", "message": "invalid request body"})
		return false
	}
	return true
}
func (h *Handler) writeJSON(w http.ResponseWriter, payload any) {
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(payload)
}
func (h *Handler) writeStatus(w http.ResponseWriter, status int, payload any) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
func (h *Handler) writeError(w http.ResponseWriter, status int, err error) {
	h.writeStatus(w, status, map[string]string{"status": "error", "message": err.Error()})
}
