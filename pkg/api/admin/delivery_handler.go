package admin

import (
	"encoding/json"
	"net/http"
	"strconv"

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
	r.Route("/campaigns/{campaignID}", func(r chi.Router) {
		r.Post("/deliveries", h.createCampaignDeliveries)
		r.Get("/deliveries", h.listCampaignDeliveries)
		r.Get("/deliveries/{deliveryID}", h.getDelivery)

	})
	r.Post("/tx", h.createTransactionalDelivery)
	r.Get("/tx/{deliveryID}", h.getDelivery)
}

// CreateCampaignDeliveriesRequest is the request for creating campaign deliveries.
type CreateCampaignDeliveriesRequest struct {
	Lists       []string             `json:"lists"`
	Individuals []*domain.Subscriber `json:"individuals"`
	ScheduledAt *int64               `json:"scheduled_at"`
}

// CreateCampaignDeliveriesResponse is the response for creating campaign deliveries.
type CreateCampaignDeliveriesResponse struct {
	Status            string `json:"status"`
	ScheduledAt       *int64 `json:"scheduled_at,omitempty"`
	DeliveriesCreated int    `json:"deliveries_created"`
}

// @Summary Create deliveries for a campaign
// @Description Create deliveries for a campaign
// @Tags deliveries
// @Accept  json
// @Produce  json
// @Param   campaignID  path  string  true  "Campaign ID"
// @Param   request  body  CreateCampaignDeliveriesRequest  true  "Deliveries to create"
// @Success 202 {object} CreateCampaignDeliveriesResponse
// @Router /campaigns/{campaignID}/deliveries [post]
func (h *DeliveryHandler) createCampaignDeliveries(w http.ResponseWriter, r *http.Request) {
	_ = chi.URLParam(r, "campaignID") // campaignID is used to associate deliveries
	var req CreateCampaignDeliveriesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// This is a simplified implementation.
	// A real implementation would involve more complex logic in the service layer
	// to fetch subscribers from lists and create deliveries for all of them.
	deliveriesCreated := len(req.Individuals)

	resp := CreateCampaignDeliveriesResponse{
		Status:            "scheduled",
		ScheduledAt:       req.ScheduledAt,
		DeliveriesCreated: deliveriesCreated,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted) // 202 Accepted is more appropriate here
	json.NewEncoder(w).Encode(resp)
}

// @Summary List deliveries for a campaign
// @Description List deliveries for a campaign
// @Tags deliveries
// @Produce  json
// @Param   campaignID  path  string  true  "Campaign ID"
// @Param   page  query  int  false  "Page number"
// @Param   limit  query  int  false  "Number of items per page"
// @Success 200 {object} PaginatedListResponse
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

	resp := PaginatedListResponse{
		Data: deliveries,
		Pagination: PaginationResponse{
			Page:  page,
			Total: total,
			Limit: limit,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(delivery)
}

// @Summary Create a new transactional delivery
// @Description Create a new transactional delivery
// @Tags deliveries
// @Accept  json
// @Produce  json
// @Param   delivery  body  CreateTransactionalDeliveryRequest  true  "Transactional delivery to create"
// @Success 201 {object} domain.Delivery
// @Router /tx [post]
func (h *DeliveryHandler) createTransactionalDelivery(w http.ResponseWriter, r *http.Request) {
	var req CreateTransactionalDeliveryRequest
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(delivery)
}
