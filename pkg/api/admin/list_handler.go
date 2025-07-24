package admin

import (
	"encoding/json"
	"github.com/headmail/headmail/pkg/api/admin/dto"
	"net/http"
	"strconv"

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
