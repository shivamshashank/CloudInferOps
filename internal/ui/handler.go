package ui

import (
	"encoding/json"
	"net/http"
	"strings"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.URL.Path {
	case "/api/health":
		h.writeJSON(w, h.service.Health())
	case "/api/overview":
		h.writeJSON(w, h.service.Overview())
	case "/api/deployments":
		h.writeJSON(w, h.service.Deployments())
	case "/api/models":
		h.writeJSON(w, h.service.Models())
	case "/api/alerts":
		h.writeJSON(w, h.service.Alerts())
	case "/api/logs":
		service := r.URL.Query().Get("service")
		h.writeJSON(w, h.service.LogSummary(service))
	case "/api/actions/deploy":
		if r.Method != http.MethodPost {
			h.writeStatus(w, http.StatusMethodNotAllowed, map[string]string{"status": "error", "message": "method not allowed"})
			return
		}
		payload := struct {
			Name string `json:"name"`
		}{}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			h.writeStatus(w, http.StatusBadRequest, map[string]string{"status": "error", "message": "invalid request body"})
			return
		}
		if strings.TrimSpace(payload.Name) == "" {
			h.writeStatus(w, http.StatusBadRequest, map[string]string{"status": "error", "message": "deployment target is required"})
			return
		}
		h.writeJSON(w, h.service.DeployAction(payload.Name))
	default:
		h.writeJSON(w, map[string]string{"status": "not_found"})
	}
}

func (h *Handler) writeJSON(w http.ResponseWriter, payload any) {
	_ = json.NewEncoder(w).Encode(payload)
}

func (h *Handler) writeStatus(w http.ResponseWriter, status int, payload any) {
	w.WriteHeader(status)
	h.writeJSON(w, payload)
}
