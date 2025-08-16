// Copyright 2025 JC-Lab
// SPDX-License-Identifier: AGPL-3.0-or-later

package admin

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/headmail/headmail/pkg/api/admin/dto"

	"github.com/go-chi/chi/v5"
	"github.com/headmail/headmail/pkg/domain"
	"github.com/headmail/headmail/pkg/repository"
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
	r.Route("/lists", func(r chi.Router) {
		r.Post("/", h.createList)
		r.Get("/", h.listLists)

		// Subscribers management under a list
		r.Route("/{listID}/subscribers", func(r chi.Router) {
			r.Post("/", h.addSubscribers)
			r.Get("/", h.listSubscribersOfList)
			r.Patch("/", h.patchSubscribersOfList)
			r.Put("/", h.replaceSubscribersOfList)
		})

		r.Route("/{listID}", func(r chi.Router) {
			r.Get("/", h.getList)
			r.Put("/", h.updateList)
			r.Delete("/", h.deleteList)
		})
	})
}

// @Summary Create a new mailing list
// @Description Create a new mailing list
// @Tags lists
// @Accept  json
// @Produce  json
// @Param   list  body  dto.CreateListRequest  true  "List to create"
// @Success 201 {object} domain.List
// @Router /lists [post]
func (h *ListHandler) createList(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateListRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	list := &domain.List{
		Name:        req.Name,
		Description: req.Description,
		Tags:        req.Tags,
	}

	if err := h.service.CreateList(r.Context(), list); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJson(w, http.StatusCreated, list)
}

// @Summary Get a mailing list by ID
// @Description Get a mailing list by ID
// @Tags lists
// @Produce  json
// @Param   listID  path  string  true  "List ID"
// @Success 200 {object} domain.List
// @Router /lists/{listID} [get]
func (h *ListHandler) getList(w http.ResponseWriter, r *http.Request) {
	listID := chi.URLParam(r, "listID")

	list, err := h.service.GetList(r.Context(), listID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	count, err := h.service.GetSubscriberCount(r.Context(), listID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	list.SubscriberCount = count

	writeJson(w, http.StatusOK, list)
}

// @Summary Update a mailing list
// @Description Update a mailing list
// @Tags lists
// @Accept  json
// @Produce  json
// @Param   listID  path  string  true  "List ID"
// @Param   list  body  dto.UpdateListRequest  true  "List to update"
// @Success 200 {object} domain.List
// @Router /lists/{listID} [put]
func (h *ListHandler) updateList(w http.ResponseWriter, r *http.Request) {
	listID := chi.URLParam(r, "listID")

	var req dto.UpdateListRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	list := &domain.List{
		ID:          listID,
		Name:        req.Name,
		Description: req.Description,
		Tags:        req.Tags,
	}

	if err := h.service.UpdateList(r.Context(), list); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	count, err := h.service.GetSubscriberCount(r.Context(), listID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	list.SubscriberCount = count

	writeJson(w, http.StatusOK, list)
}

// @Summary Delete a mailing list
// @Description Delete a mailing list
// @Tags lists
// @Produce  json
// @Param   listID  path  string  true  "List ID"
// @Success 200 {object} DeleteResponse
// @Router /lists/{listID} [delete]
func (h *ListHandler) deleteList(w http.ResponseWriter, r *http.Request) {
	listID := chi.URLParam(r, "listID")

	if err := h.service.DeleteList(r.Context(), listID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := DeleteResponse{
		Deleted: true,
		Message: "List deleted successfully",
	}

	writeJson(w, http.StatusOK, resp)
}

// @Summary Add subscribers to a list
// @Description Add subscribers to a list
// @Tags lists
// @Accept  json
// @Produce  json
// @Param   listID  path  string  true  "List ID"
// @Param   request  body  dto.CreateSubscribersRequest  true  "Subscribers to add"
// @Success 201 {object} EmptyResponse
// @Router /lists/{listID}/subscribers [post]
func (h *ListHandler) addSubscribers(w http.ResponseWriter, r *http.Request) {
	listID := chi.URLParam(r, "listID")
	var req dto.CreateSubscribersRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var subscribers []*domain.Subscriber
	for _, subReq := range req.Subscribers {
		sub := &domain.Subscriber{
			Email:  subReq.Email,
			Name:   subReq.Name,
			Status: domain.SubscriberStatusEnabled,
			Lists: []domain.SubscriberList{
				{
					ListID: listID,
					Status: domain.SubscriberListStatusConfirmed,
				},
			},
		}
		if subReq.Status != nil && *subReq.Status != "" {
			sub.Status = *subReq.Status
		}
		subscribers = append(subscribers, sub)
	}

	if err := h.service.AddSubscribers(r.Context(), subscribers); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJson(w, http.StatusCreated, &EmptyResponse{})
}

// @Summary List subscribers in a list
// @Description List subscribers in a list
// @Tags lists
// @Produce  json
// @Param   listID  path  string  true  "List ID"
// @Param   page  query  int  false  "Page number"
// @Param   limit  query  int  false  "Number of items per page"
// @Param   status  query  string  false  "Filter by status"
// @Param   search  query  string  false  "Search term"
// @Success 200 {object} PaginatedListResponse[domain.Subscriber]
// @Router /lists/{listID}/subscribers [get]
func (h *ListHandler) listSubscribersOfList(w http.ResponseWriter, r *http.Request) {
	listID := chi.URLParam(r, "listID")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page == 0 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 {
		limit = 20
	}

	filter := repository.SubscriberFilter{
		ListID: listID,
		Status: domain.SubscriberStatus(r.URL.Query().Get("status")),
		Search: r.URL.Query().Get("search"),
	}
	pagination := repository.Pagination{
		Page:  page,
		Limit: limit,
	}

	subscribers, total, err := h.service.ListSubscribers(r.Context(), filter, pagination)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := &PaginatedListResponse[*domain.Subscriber]{
		Data: subscribers,
		Pagination: PaginationResponse{
			Page:  page,
			Total: total,
			Limit: limit,
		},
	}
	writeJson(w, http.StatusOK, resp)
}

// @Summary Patch subscribers in a list (add/remove)
// @Description Add or remove subscribers to/from a list
// @Tags lists
// @Accept  json
// @Produce  json
// @Param   listID  path  string  true  "List ID"
// @Param   request  body  dto.PatchSubscribersRequest  true  "Patch request"
// @Success 204
// @Router /lists/{listID}/subscribers [patch]
func (h *ListHandler) patchSubscribersOfList(w http.ResponseWriter, r *http.Request) {
	listID := chi.URLParam(r, "listID")
	var req dto.PatchSubscribersRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ls, ok := h.service.(interface {
		PatchSubscribersInList(ctx context.Context, listID string, add []string, remove []string) error
	})
	if !ok {
		http.Error(w, "service does not support patch operation", http.StatusInternalServerError)
		return
	}

	if err := ls.PatchSubscribersInList(r.Context(), listID, req.Add, req.Remove); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// @Summary Replace subscribers in a list
// @Description Replace list members atomically
// @Tags lists
// @Accept  json
// @Produce  json
// @Param   listID  path  string  true  "List ID"
// @Param   request  body  dto.ReplaceSubscribersRequest  true  "Replace request"
// @Success 204
// @Router /lists/{listID}/subscribers [put]
func (h *ListHandler) replaceSubscribersOfList(w http.ResponseWriter, r *http.Request) {
	listID := chi.URLParam(r, "listID")
	var req dto.ReplaceSubscribersRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ls, ok := h.service.(interface {
		ReplaceSubscribersInList(ctx context.Context, listID string, subscriberIDs []string) error
	})
	if !ok {
		http.Error(w, "service does not support replace operation", http.StatusInternalServerError)
		return
	}

	if err := ls.ReplaceSubscribersInList(r.Context(), listID, req.Subscribers); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// @Summary List all mailing lists
// @Description List all mailing lists
// @Tags lists
// @Produce  json
// @Param   page  query  int  false  "Page number"
// @Param   limit  query  int  false  "Number of items per page"
// @Param   search  query  string  false  "Search term"
// @Param   tags[]  query  []string  false  "Tags to filter by"
// @Success 200 {object} PaginatedListResponse[domain.List]
// @Router /lists [get]
func (h *ListHandler) listLists(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page == 0 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 {
		limit = 20
	}

	filter := repository.ListFilter{
		Search: r.URL.Query().Get("search"),
		Tags:   r.URL.Query()["tags[]"],
	}
	pagination := repository.Pagination{
		Page:  page,
		Limit: limit,
	}

	lists, total, err := h.service.ListLists(r.Context(), filter, pagination)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, list := range lists {
		count, err := h.service.GetSubscriberCount(r.Context(), list.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		list.SubscriberCount = count
	}

	resp := &PaginatedListResponse[*domain.List]{
		Data: lists,
		Pagination: PaginationResponse{
			Page:  page,
			Total: total,
			Limit: limit,
		},
	}

	writeJson(w, http.StatusOK, resp)
}
