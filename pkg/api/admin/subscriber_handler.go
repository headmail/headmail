// Copyright 2025 JC-Lab
// SPDX-License-Identifier: AGPL-3.0-or-later

package admin

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/headmail/headmail/pkg/api/admin/dto"

	"github.com/go-chi/chi/v5"
	"github.com/headmail/headmail/pkg/domain"
	"github.com/headmail/headmail/pkg/repository"
	"github.com/headmail/headmail/pkg/service"
)

// SubscriberHandler handles HTTP requests for subscribers.
type SubscriberHandler struct {
	service service.ListServiceProvider
}

// NewSubscriberHandler creates a new SubscriberHandler.
func NewSubscriberHandler(service service.ListServiceProvider) *SubscriberHandler {
	return &SubscriberHandler{service: service}
}

// RegisterRoutes registers the subscriber routes to the router.
func (h *SubscriberHandler) RegisterRoutes(r chi.Router) {
	r.Get("/subscribers", h.listSubscribers)
	r.Route("/subscribers/{subscriberID}", func(r chi.Router) {
		r.Get("/", h.getSubscriber)
		r.Put("/", h.updateSubscriber)
		r.Delete("/", h.deleteSubscriber)
	})
}

// CreateSubscribersRequest is the request for creating subscribers.
type CreateSubscribersRequest struct {
	Subscribers []*dto.CreateSubscriberRequest `json:"subscribers"`
	Append      bool                           `json:"append"`
}

type EmptyResponse struct{}

// @Summary Add subscribers to a list
// @Description Add subscribers to a list
// @Tags subscribers
// @Accept  json
// @Produce  json
// @Param   listID  path  string  true  "List ID"
// @Param   request  body  CreateSubscribersRequest  true  "Subscribers to add"
// @Success 201 {object} EmptyResponse
// @Router /lists/{listID}/subscribers [post]
func (h *SubscriberHandler) addSubscriber(w http.ResponseWriter, r *http.Request) {
	listID := chi.URLParam(r, "listID")
	var req CreateSubscribersRequest
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

// @Summary Get a subscriber by ID
// @Description Get a subscriber by ID
// @Tags subscribers
// @Produce  json
// @Param   subscriberID  path  string  true  "Subscriber ID"
// @Success 200 {object} domain.Subscriber
// @Router /subscribers/{subscriberID} [get]
func (h *SubscriberHandler) getSubscriber(w http.ResponseWriter, r *http.Request) {
	subscriberID := chi.URLParam(r, "subscriberID")

	subscriber, err := h.service.GetSubscriber(r.Context(), subscriberID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJson(w, http.StatusOK, subscriber)
}

// @Summary Update a subscriber
// @Description Update a subscriber
// @Tags subscribers
// @Accept  json
// @Produce  json
// @Param   subscriberID  path  string  true  "Subscriber ID"
// @Param   subscriber  body  dto.UpdateSubscriberRequest  true  "Subscriber to update"
// @Success 200 {object} domain.Subscriber
// @Router /subscribers/{subscriberID} [put]
func (h *SubscriberHandler) updateSubscriber(w http.ResponseWriter, r *http.Request) {
	subscriberID := chi.URLParam(r, "subscriberID")

	var req dto.UpdateSubscriberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	subscriber := &domain.Subscriber{
		ID:     subscriberID,
		Email:  req.Email,
		Name:   req.Name,
		Status: domain.SubscriberStatusEnabled,
	}
	if req.Status != nil && *req.Status != "" {
		subscriber.Status = *req.Status
	}

	if err := h.service.UpdateSubscriber(r.Context(), subscriber); err != nil {
		var e *repository.ErrUniqueConstraintFailed
		if errors.As(err, &e) {
			http.Error(w, "Email address already exists", http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJson(w, http.StatusOK, subscriber)
}

// @Summary Delete a subscriber
// @Description Delete a subscriber
// @Tags subscribers
// @Produce  json
// @Param   subscriberID  path  string  true  "Subscriber ID"
// @Success 200 {object} DeleteResponse
// @Router /subscribers/{subscriberID} [delete]
func (h *SubscriberHandler) deleteSubscriber(w http.ResponseWriter, r *http.Request) {
	subscriberID := chi.URLParam(r, "subscriberID")

	if err := h.service.DeleteSubscriber(r.Context(), subscriberID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := DeleteResponse{
		Deleted: true,
		Message: "Subscriber deleted successfully",
	}
	writeJson(w, http.StatusOK, resp)
}

// @Summary List subscribers in a list
// @Description List subscribers in a list
// @Tags subscribers
// @Produce  json
// @Param   listID  path  string  true  "List ID"
// @Param   page  query  int  false  "Page number"
// @Param   limit  query  int  false  "Number of items per page"
// @Param   status  query  string  false  "Filter by status"
// @Param   search  query  string  false  "Search term"
// @Success 200 {object} PaginatedListResponse[domain.Subscriber]
// @Router /subscribers [get]
func (h *SubscriberHandler) listSubscribersOfList(w http.ResponseWriter, r *http.Request) {
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

// @Summary List subscribers
// @Description List subscribers
// @Tags subscribers
// @Produce  json
// @Param   page  query  int  false  "Page number"
// @Param   limit  query  int  false  "Number of items per page"
// @Param   status  query  string  false  "Filter by status"
// @Param   search  query  string  false  "Search term"
// @Success 200 {object} PaginatedListResponse[domain.Subscriber]
// @Router /subscribers [get]
func (h *SubscriberHandler) listSubscribers(w http.ResponseWriter, r *http.Request) {
	h.listSubscribersOfList(w, r)
}
