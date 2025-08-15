package admin

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/headmail/headmail/pkg/api/admin/dto"
	"github.com/headmail/headmail/pkg/domain"
	"github.com/headmail/headmail/pkg/repository"
	"github.com/headmail/headmail/pkg/service"
	tmpl "github.com/headmail/headmail/pkg/template"
)

// TemplateHandler handles HTTP requests for templates.
type TemplateHandler struct {
	service service.TemplateServiceProvider
}

// NewTemplateHandler creates a new TemplateHandler.
func NewTemplateHandler(service service.TemplateServiceProvider) *TemplateHandler {
	return &TemplateHandler{service: service}
}

// RegisterRoutes registers the template routes to the router.
func (h *TemplateHandler) RegisterRoutes(r chi.Router) {
	r.Route("/templates", func(r chi.Router) {
		// server-side preview endpoint used by the editor to render templates with sample data
		r.Post("/preview", h.previewTemplate)

		r.Post("/", h.createTemplate)
		r.Get("/", h.listTemplates)
		r.Route("/{templateID}", func(r chi.Router) {
			r.Get("/", h.getTemplate)
			r.Put("/", h.updateTemplate)
			r.Delete("/", h.deleteTemplate)
		})
	})
}

// @Summary Create a new template
// @Description Create a new template
// @Tags templates
// @Accept  json
// @Produce  json
// @Param   template  body  dto.CreateTemplateRequest  true  "Template to create"
// @Success 201 {object} domain.Template
// @Router /templates [post]
func (h *TemplateHandler) createTemplate(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	template := &domain.Template{
		Name:     req.Name,
		BodyHTML: req.BodyHTML,
		BodyText: req.BodyText,
		BodyMJML: req.BodyMJML,
	}

	if err := h.service.CreateTemplate(r.Context(), template); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJson(w, http.StatusCreated, template)
}

// @Summary Render template preview
// @Description Renders provided template HTML/text/subject with sample data (name, email) and returns rendered output.
// @Tags templates
// @Accept  json
// @Produce  json
// @Param   preview  body  dto.PreviewTemplateRequest  true  "Preview request"
// @Success 200 {object} dto.PreviewTemplateResponse
// @Router /templates/preview [post]
func (h *TemplateHandler) previewTemplate(w http.ResponseWriter, r *http.Request) {
	var req dto.PreviewTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Use package template service for rendering
	svc := tmpl.NewService()

	// Sample data (matching createDeliveryFromCampaign defaults)
	data := map[string]interface{}{
		"name":       req.Name,
		"email":      req.Email,
		"deliveryId": uuid.New().String(),
	}

	renderedHTML, err := svc.Render(req.TemplateHTML, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var renderedText string
	if req.TemplateText != "" {
		if t, err := svc.Render(req.TemplateText, data); err == nil {
			renderedText = t
		}
	}

	var renderedSubject string
	if req.Subject != "" {
		if s, err := svc.Render(req.Subject, data); err == nil {
			renderedSubject = s
		}
	}

	resp := &dto.PreviewTemplateResponse{
		HTML:    renderedHTML,
		Text:    renderedText,
		Subject: renderedSubject,
	}
	writeJson(w, http.StatusOK, resp)
}

// @Summary Get a template by ID
// @Description Get a template by ID
// @Tags templates
// @Produce  json
// @Param   templateID  path  string  true  "Template ID"
// @Success 200 {object} domain.Template
// @Router /templates/{templateID} [get]
func (h *TemplateHandler) getTemplate(w http.ResponseWriter, r *http.Request) {
	templateID := chi.URLParam(r, "templateID")

	template, err := h.service.GetTemplate(r.Context(), templateID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJson(w, http.StatusOK, template)
}

// @Summary Update a template
// @Description Update a template
// @Tags templates
// @Accept  json
// @Produce  json
// @Param   templateID  path  string  true  "Template ID"
// @Param   template  body  dto.UpdateTemplateRequest  true  "Template to update"
// @Success 200 {object} domain.Template
// @Router /templates/{templateID} [put]
func (h *TemplateHandler) updateTemplate(w http.ResponseWriter, r *http.Request) {
	templateID := chi.URLParam(r, "templateID")

	var req dto.UpdateTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	template := &domain.Template{
		ID:       templateID,
		Name:     req.Name,
		BodyHTML: req.BodyHTML,
		BodyText: req.BodyText,
		BodyMJML: req.BodyMJML,
	}

	if err := h.service.UpdateTemplate(r.Context(), template); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJson(w, http.StatusOK, template)
}

// @Summary Delete a template
// @Description Delete a template
// @Tags templates
// @Produce  json
// @Param   templateID  path  string  true  "Template ID"
// @Success 200 {object} DeleteResponse
// @Router /templates/{templateID} [delete]
func (h *TemplateHandler) deleteTemplate(w http.ResponseWriter, r *http.Request) {
	templateID := chi.URLParam(r, "templateID")

	if err := h.service.DeleteTemplate(r.Context(), templateID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := DeleteResponse{
		Deleted: true,
		Message: "Template deleted successfully",
	}

	writeJson(w, http.StatusOK, resp)
}

// @Summary List all templates
// @Description List all templates
// @Tags templates
// @Produce  json
// @Param   page  query  int  false  "Page number"
// @Param   limit  query  int  false  "Number of items per page"
// @Success 200 {object} PaginatedListResponse[domain.Template]
// @Router /templates [get]
func (h *TemplateHandler) listTemplates(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page == 0 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 {
		limit = 20
	}

	pagination := repository.Pagination{
		Page:  page,
		Limit: limit,
	}

	templates, total, err := h.service.ListTemplates(r.Context(), pagination)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := &PaginatedListResponse[*domain.Template]{
		Data: templates,
		Pagination: PaginationResponse{
			Page:  page,
			Total: total,
			Limit: limit,
		},
	}

	writeJson(w, http.StatusOK, resp)
}
