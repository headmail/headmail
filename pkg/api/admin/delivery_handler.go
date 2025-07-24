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
	service service.DeliveryServiceProvider
}

// NewDeliveryHandler creates a new DeliveryHandler.
func NewDeliveryHandler(service service.DeliveryServiceProvider) *DeliveryHandler {
	return &DeliveryHandler{service: service}
}

// RegisterRoutes registers the delivery routes to the router.
func (h *DeliveryHandler) RegisterRoutes(r chi.Router) {
	r.Get("/campaigns/{campaignID}/deliveries", h.listCampaignDeliveries)
	r.Get("/campaigns/{campaignID}/deliveries/{deliveryID}", h.getDelivery)
	r.Post("/tx", h.createTransactionalDelivery)
	r.Get("/tx/{deliveryID}", h.getDelivery)
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

	delivery := &domain.Delivery{
		Type:    "transactional",
		Name:    req.Name,
		Email:   req.Email,
		Subject: req.Subject,
		Data:    req.Data,
		Headers: req.Headers,
		Tags:    req.Tags,
	}

	if err := h.service.CreateDelivery(r.Context(), delivery); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJson(w, http.StatusCreated, delivery)
}
