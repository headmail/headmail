// Copyright 2025 JC-Lab
// SPDX-License-Identifier: AGPL-3.0-or-later

package admin

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/headmail/headmail/pkg/api/admin/dto"

	"github.com/go-chi/chi/v5"
	"github.com/headmail/headmail/pkg/domain"
	"github.com/headmail/headmail/pkg/repository"
	"github.com/headmail/headmail/pkg/service"
)

// UpdateCampaignStatusRequest is the request for updating a campaign's status.
type UpdateCampaignStatusRequest struct {
	Status domain.CampaignStatus `json:"status"`
}

// CampaignHandler handles HTTP requests for campaigns.
type CampaignHandler struct {
	service service.CampaignServiceProvider
}

// NewCampaignHandler creates a new CampaignHandler.
func NewCampaignHandler(service service.CampaignServiceProvider) *CampaignHandler {
	return &CampaignHandler{service: service}
}

// RegisterRoutes registers the campaign routes to the router.
func (h *CampaignHandler) RegisterRoutes(r chi.Router) {
	r.Post("/campaigns", h.createCampaign)
	// Allow creating a campaign with a pre-defined ID. Accepts optional ?upsert=true to update existing.
	r.Post("/campaigns/{campaignID}", h.createCampaignWithID)
	r.Get("/campaigns", h.listCampaigns)
	r.Get("/campaigns/{campaignID}", h.getCampaign)
	r.Put("/campaigns/{campaignID}", h.updateCampaign)
	r.Delete("/campaigns/{campaignID}", h.deleteCampaign)
	r.Patch("/campaigns/{campaignID}/status", h.updateCampaignStatus)
	r.Post("/campaigns/{campaignID}/deliveries", h.createCampaignDeliveries)
	r.Get("/campaigns/stats", h.getCampaignsStats)
	r.Get("/campaigns/{campaignID}/stats", h.getCampaignStats)
}

// @Summary Create a new campaign
// @Description Create a new campaign
// @Tags campaigns
// @Accept  json
// @Produce  json
// @Param   campaign  body  dto.CreateCampaignRequest  true  "Campaign to create"
// @Success 201 {object} domain.Campaign
// @Router /campaigns [post]
func (h *CampaignHandler) createCampaign(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateCampaignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	campaign := &domain.Campaign{
		Name:         req.Name,
		Status:       req.Status,
		FromName:     req.FromName,
		FromEmail:    req.FromEmail,
		Subject:      req.Subject,
		TemplateID:   req.TemplateID,
		TemplateHTML: req.TemplateHTML,
		TemplateText: req.TemplateText,
		Data:         req.Data,
		Tags:         req.Tags,
		Headers:      req.Headers,
		UTMParams:    req.UTMParams,
		ScheduledAt:  req.ScheduledAt,
	}
	if campaign.Status == "" {
		campaign.Status = domain.CampaignStatusDraft
	}

	// Create normally (no upsert)
	if err := h.service.CreateCampaign(r.Context(), campaign, false); err != nil {
		// map unique constraint to 409
		if _, ok := err.(*repository.ErrUniqueConstraintFailed); ok {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJson(w, http.StatusCreated, campaign)
}

// createCampaignWithID handles POST /campaigns/{campaignID}?upsert=bool
// @Summary Create or upsert a campaign with given ID
// @Description Create a campaign specifying the ID in the path. Use ?upsert=true to update an existing campaign with the same ID.
// @Tags campaigns
// @Accept  json
// @Produce  json
// @Param   campaignID  path  string  true  "Campaign ID"
// @Param   upsert  query  bool  false  "Upsert if exists"
// @Param   campaign  body  dto.CreateCampaignRequest  true  "Campaign to create"
// @Success 201 {object} domain.Campaign
// @Failure 409 {object} map[string]string "Conflict - ID already exists"
// @Failure 400 {object} map[string]string "Bad request"
// @Router /campaigns/{campaignID} [post]
func (h *CampaignHandler) createCampaignWithID(w http.ResponseWriter, r *http.Request) {
	campaignID := chi.URLParam(r, "campaignID")
	if campaignID == "" {
		http.Error(w, "missing campaignID path param", http.StatusBadRequest)
		return
	}

	upsert := false
	if upStr := r.URL.Query().Get("upsert"); upStr != "" {
		if b, err := strconv.ParseBool(upStr); err == nil {
			upsert = b
		}
	}

	var req dto.CreateCampaignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	campaign := &domain.Campaign{
		ID:           campaignID,
		Name:         req.Name,
		Status:       req.Status,
		FromName:     req.FromName,
		FromEmail:    req.FromEmail,
		Subject:      req.Subject,
		TemplateID:   req.TemplateID,
		TemplateHTML: req.TemplateHTML,
		TemplateText: req.TemplateText,
		Data:         req.Data,
		Tags:         req.Tags,
		Headers:      req.Headers,
		UTMParams:    req.UTMParams,
		ScheduledAt:  req.ScheduledAt,
	}
	if campaign.Status == "" {
		campaign.Status = domain.CampaignStatusDraft
	}

	if err := h.service.CreateCampaign(r.Context(), campaign, upsert); err != nil {
		if _, ok := err.(*repository.ErrUniqueConstraintFailed); ok {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 201 Created for new resource, 200 OK if upserted existing (we can't easily tell without extra repo info).
	// For simplicity, return 201 when created; if upsert=true it's semantically acceptable to return 200,
	// but here we'll return 201 regardless.
	writeJson(w, http.StatusCreated, campaign)
}

// @Summary Get a campaign by ID
// @Description Get a campaign by ID
// @Tags campaigns
// @Produce  json
// @Param   campaignID  path  string  true  "Campaign ID"
// @Success 200 {object} domain.Campaign
// @Router /campaigns/{campaignID} [get]
func (h *CampaignHandler) getCampaign(w http.ResponseWriter, r *http.Request) {
	campaignID := chi.URLParam(r, "campaignID")

	campaign, err := h.service.GetCampaign(r.Context(), campaignID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJson(w, http.StatusOK, campaign)
}

// @Summary Update a campaign
// @Description Update a campaign
// @Tags campaigns
// @Accept  json
// @Produce  json
// @Param   campaignID  path  string  true  "Campaign ID"
// @Param   campaign  body  dto.UpdateCampaignRequest  true  "Campaign to update"
// @Success 200 {object} domain.Campaign
// @Router /campaigns/{campaignID} [put]
func (h *CampaignHandler) updateCampaign(w http.ResponseWriter, r *http.Request) {
	campaignID := chi.URLParam(r, "campaignID")

	var req dto.UpdateCampaignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	campaign := &domain.Campaign{
		ID:           campaignID,
		Name:         req.Name,
		Status:       req.Status,
		FromName:     req.FromName,
		FromEmail:    req.FromEmail,
		Subject:      req.Subject,
		TemplateID:   req.TemplateID,
		TemplateHTML: req.TemplateHTML,
		TemplateText: req.TemplateText,
		Data:         req.Data,
		Tags:         req.Tags,
		Headers:      req.Headers,
		UTMParams:    req.UTMParams,
		ScheduledAt:  req.ScheduledAt,
	}

	if err := h.service.UpdateCampaign(r.Context(), campaign); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJson(w, http.StatusOK, campaign)
}

// @Summary Delete a campaign
// @Description Delete a campaign
// @Tags campaigns
// @Produce  json
// @Param   campaignID  path  string  true  "Campaign ID"
// @Success 200 {object} DeleteResponse
// @Router /campaigns/{campaignID} [delete]
func (h *CampaignHandler) deleteCampaign(w http.ResponseWriter, r *http.Request) {
	campaignID := chi.URLParam(r, "campaignID")

	if err := h.service.DeleteCampaign(r.Context(), campaignID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := DeleteResponse{
		Deleted: true,
		Message: "Campaign deleted successfully",
	}

	writeJson(w, http.StatusOK, resp)
}

// @Summary List all campaigns
// @Description List all campaigns
// @Tags campaigns
// @Produce  json
// @Param   page  query  int  false  "Page number"
// @Param   limit  query  int  false  "Number of items per page"
// @Param   search  query  string  false  "Search term"
// @Param   tags[]  query  []string  false  "Tags to filter by"
// @Param   status[]  query  []string  false  "Status to filter by"
// @Success 200 {object} PaginatedListResponse[domain.Campaign]
// @Router /campaigns [get]
func (h *CampaignHandler) listCampaigns(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page == 0 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 {
		limit = 20
	}

	var statuses []domain.CampaignStatus
	for _, s := range r.URL.Query()["status[]"] {
		statuses = append(statuses, domain.CampaignStatus(s))
	}

	filter := repository.CampaignFilter{
		Search: r.URL.Query().Get("search"),
		Tags:   r.URL.Query()["tags[]"],
		Status: statuses,
	}
	pagination := repository.Pagination{
		Page:  page,
		Limit: limit,
	}

	campaigns, total, err := h.service.ListCampaigns(r.Context(), filter, pagination)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := &PaginatedListResponse[*domain.Campaign]{
		Data: campaigns,
		Pagination: PaginationResponse{
			Page:  page,
			Total: total,
			Limit: limit,
		},
	}

	writeJson(w, http.StatusOK, resp)
}

// @Summary Update campaign status
// @Description Update campaign status
// @Tags campaigns
// @Accept  json
// @Produce  json
// @Param   campaignID  path  string  true  "Campaign ID"
// @Param   statusUpdate  body  UpdateCampaignStatusRequest  true  "Status update"
// @Success 204 "No Content"
// @Router /campaigns/{campaignID}/status [patch]
func (h *CampaignHandler) updateCampaignStatus(w http.ResponseWriter, r *http.Request) {
	campaignID := chi.URLParam(r, "campaignID")

	var req UpdateCampaignStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateCampaignStatus(r.Context(), campaignID, req.Status); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// @Summary Create deliveries for a campaign
// @Description Create deliveries for a campaign
// @Tags campaigns
// @Accept  json
// @Produce  json
// @Param   campaignID  path  string  true  "Campaign ID"
// @Param   request  body  dto.CreateDeliveriesRequest  true  "Deliveries to create"
// @Success 202 {object} dto.CreateDeliveriesResponse
// @Router /campaigns/{campaignID}/deliveries [post]
func (h *CampaignHandler) createCampaignDeliveries(w http.ResponseWriter, r *http.Request) {
	campaignID := chi.URLParam(r, "campaignID")
	var req dto.CreateDeliveriesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	count, err := h.service.CreateDeliveries(r.Context(), campaignID, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := &dto.CreateDeliveriesResponse{
		Status:            "scheduled",
		ScheduledAt:       req.ScheduledAt,
		DeliveriesCreated: count,
	}

	writeJson(w, http.StatusCreated, resp)
}

// parseFromTo parses 'from' and 'to' query params as unix seconds. Returns defaults if not provided.
func parseFromTo(r *http.Request) (time.Time, time.Time, error) {
	now := time.Now()
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")
	var from time.Time
	var to time.Time
	if fromStr == "" {
		// default: 24h ago
		from = now.Add(-24 * time.Hour)
	} else {
		ts, err := strconv.ParseInt(fromStr, 10, 64)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
		from = time.Unix(ts, 0)
	}
	if toStr == "" {
		to = now
	} else {
		ts, err := strconv.ParseInt(toStr, 10, 64)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
		to = time.Unix(ts, 0)
	}
	return from, to, nil
}

// getCampaignsStats handles GET /campaigns/stats?campaign_ids=cid1,cid2&from=...&to=...&granularity=hour|day
func (h *CampaignHandler) getCampaignsStats(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("campaign_ids")
	if q == "" {
		http.Error(w, "missing campaign_ids query param", http.StatusBadRequest)
		return
	}
	campaignIDs := strings.Split(q, ",")
	from, to, err := parseFromTo(r)
	if err != nil {
		http.Error(w, "invalid from/to params", http.StatusBadRequest)
		return
	}
	gran := r.URL.Query().Get("granularity")
	if gran == "" {
		gran = "hour"
	}
	stats, err := h.service.GetCampaignStats(r.Context(), campaignIDs, from, to, gran)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJson(w, http.StatusOK, stats)
}

// getCampaignStats handles GET /campaigns/{campaignID}/stats?from=...&to=...&granularity=hour|day
func (h *CampaignHandler) getCampaignStats(w http.ResponseWriter, r *http.Request) {
	campaignID := chi.URLParam(r, "campaignID")
	if campaignID == "" {
		http.Error(w, "missing campaignID path param", http.StatusBadRequest)
		return
	}
	from, to, err := parseFromTo(r)
	if err != nil {
		http.Error(w, "invalid from/to params", http.StatusBadRequest)
		return
	}
	gran := r.URL.Query().Get("granularity")
	if gran == "" {
		gran = "hour"
	}
	stats, err := h.service.GetCampaignStats(r.Context(), []string{campaignID}, from, to, gran)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJson(w, http.StatusOK, stats)
}
