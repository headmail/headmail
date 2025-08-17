// Copyright 2025 JC-Lab
// SPDX-License-Identifier: AGPL-3.0-or-later

package admin

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/headmail/headmail/pkg/api/admin/dto"

	"github.com/go-chi/chi/v5"
	"github.com/headmail/headmail/pkg/domain"
	"github.com/headmail/headmail/pkg/repository"
	"github.com/headmail/headmail/pkg/service"
)

// DeliveryHandler handles HTTP requests for deliveries.
type DeliveryHandler struct {
	service         service.DeliveryServiceProvider
	templateService service.TemplateServiceProvider
}

// NewDeliveryHandler creates a new DeliveryHandler.
func NewDeliveryHandler(service service.DeliveryServiceProvider, templateService service.TemplateServiceProvider) *DeliveryHandler {
	return &DeliveryHandler{service: service, templateService: templateService}
}

// RegisterRoutes registers the delivery routes to the router.
func (h *DeliveryHandler) RegisterRoutes(r chi.Router) {
	r.Get("/campaigns/{campaignID}/deliveries", h.listCampaignDeliveries)
	r.Get("/campaigns/{campaignID}/deliveries/{deliveryID}", h.getDelivery)
	r.Post("/tx", h.createTransactionalDelivery)
	r.Get("/tx/{deliveryID}", h.getDelivery)

	// Immediate send / retry endpoints for a specific delivery (synchronous)
	r.Post("/deliveries/{deliveryID}/send-now", h.sendNow)
	r.Post("/deliveries/{deliveryID}/retry", h.retry)
}

// @Summary List deliveries for a campaign
// @Description List deliveries for a campaign
// @Tags deliveries
// @Produce  json
// @Param   campaignID  path  string  true  "Campaign ID"
// @Param   page  query  int  false  "Page number"
// @Param   limit  query  int  false  "Number of items per page"
// @Success 200 {object} PaginatedListResponse[domain.Delivery]
// @Router /campaigns/{campaignID}/deliveries [get]
func (h *DeliveryHandler) listCampaignDeliveries(w http.ResponseWriter, r *http.Request) {
	campaignID := chi.URLParam(r, "campaignID")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page == 0 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 {
		limit = 20
	}

	filter := repository.DeliveryFilter{
		CampaignID: campaignID,
	}
	pagination := repository.Pagination{
		Page:  page,
		Limit: limit,
	}

	deliveries, total, err := h.service.ListDeliveries(r.Context(), filter, pagination)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := &PaginatedListResponse[*domain.Delivery]{
		Data: deliveries,
		Pagination: PaginationResponse{
			Page:  page,
			Total: total,
			Limit: limit,
		},
	}

	writeJson(w, http.StatusOK, resp)
}

// @Summary Get a delivery by ID
// @Description Get a delivery by ID
// @Tags deliveries
// @Produce  json
// @Param   deliveryID  path  string  true  "Delivery ID"
// @Success 200 {object} domain.Delivery
// @Router /campaigns/{campaignID}/deliveries/{deliveryID} [get]
// @Router /tx/{deliveryID} [get]
func (h *DeliveryHandler) getDelivery(w http.ResponseWriter, r *http.Request) {
	deliveryID := chi.URLParam(r, "deliveryID")

	delivery, err := h.service.GetDelivery(r.Context(), deliveryID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJson(w, http.StatusOK, delivery)
}

// @Summary Create a new transactional delivery
// @Description Create a new transactional delivery
// @Tags deliveries
// @Accept  json
// @Produce  json
// @Param   delivery  body  dto.CreateTransactionalDeliveryRequest  true  "Transactional delivery to create"
// @Success 201 {object} domain.Delivery
// @Router /tx [post]
func (h *DeliveryHandler) createTransactionalDelivery(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateTransactionalDeliveryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Prepare subject/body from request and/or template
	subject := ""
	if req.Subject != nil {
		subject = *req.Subject
	}
	bodyHTML := ""
	bodyText := ""

	// If a template_id is provided, load template and fill missing parts from it.
	if req.TemplateID != nil {
		tmpl, err := h.templateService.GetTemplate(r.Context(), *req.TemplateID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if subject == "" {
			subject = tmpl.Subject
		}
		bodyHTML = tmpl.BodyHTML
		bodyText = tmpl.BodyText
	} else {
		if req.TemplateHTML != nil {
			bodyHTML = *req.TemplateHTML
		}
		if req.TemplateText != nil {
			bodyText = *req.TemplateText
		}
	}

	delivery := &domain.Delivery{
		Type:     domain.DeliveryTypeTransaction,
		Name:     req.Name,
		Email:    req.Email,
		Subject:  subject,
		BodyHTML: bodyHTML,
		BodyText: bodyText,
		Data:     req.Data,
		Headers:  req.Headers,
		Tags:     req.Tags,
	}

	// Keep template id reference in data for auditing/rendering if provided
	if req.TemplateID != nil && *req.TemplateID != "" {
		if delivery.Data == nil {
			delivery.Data = map[string]interface{}{}
		}
		delivery.Data["template_id"] = *req.TemplateID
	}

	if err := h.service.CreateDelivery(r.Context(), delivery); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJson(w, http.StatusCreated, delivery)
}

// @Summary Send a delivery immediately (synchronous)
// @Description Perform an immediate send attempt for the specified delivery ID. This runs synchronously and returns the updated delivery object.
// @Tags deliveries
// @Accept  json
// @Produce  json
// @Param   deliveryID  path  string  true  "Delivery ID"
// @Success 200 {object} domain.Delivery
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /deliveries/{deliveryID}/send-now [post]
func (h *DeliveryHandler) sendNow(w http.ResponseWriter, r *http.Request) {
	deliveryID := chi.URLParam(r, "deliveryID")
	if deliveryID == "" {
		http.Error(w, "missing deliveryID path param", http.StatusBadRequest)
		return
	}

	delivery, err := h.service.SendNow(r.Context(), deliveryID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJson(w, http.StatusOK, delivery)
}

// @Summary Retry a delivery immediately (synchronous)
// @Description Reset attempt metadata and perform an immediate retry for the specified delivery ID. Returns the updated delivery object.
// @Tags deliveries
// @Accept  json
// @Produce  json
// @Param   deliveryID  path  string  true  "Delivery ID"
// @Success 200 {object} domain.Delivery
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /deliveries/{deliveryID}/retry [post]
func (h *DeliveryHandler) retry(w http.ResponseWriter, r *http.Request) {
	deliveryID := chi.URLParam(r, "deliveryID")
	if deliveryID == "" {
		http.Error(w, "missing deliveryID path param", http.StatusBadRequest)
		return
	}

	delivery, err := h.service.Retry(r.Context(), deliveryID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJson(w, http.StatusOK, delivery)
}
