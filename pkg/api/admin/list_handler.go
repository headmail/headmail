package admin

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/headmail/headmail/pkg/domain"
	"github.com/headmail/headmail/pkg/service"
)

// ListHandler handles HTTP requests for lists.
type ListHandler struct {
	service service.ListServiceProvider
}

// NewListHandler creates a new ListHandler.
func NewListHandler(service service.ListServiceProvider) *ListHandler {
	return &ListHandler{service: service}
}

// RegisterRoutes registers the list routes to the router.
func (h *ListHandler) RegisterRoutes(r chi.Router) {
	r.Post("/lists", h.createList)
	// TODO: Register other list routes
}

// createList handles the creation of a new mailing list.
func (h *ListHandler) createList(w http.ResponseWriter, r *http.Request) {
	var list domain.List
	if err := json.NewDecoder(r.Body).Decode(&list); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.service.CreateList(r.Context(), &list); err != nil {
		// In a real app, you'd check for specific error types
		// (e.g., validation error vs. database error)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(list)
}
