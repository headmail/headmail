package dto

// CreateListRequest is the request for creating a list.
type CreateListRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
}

// UpdateListRequest is the request for updating a list.
type UpdateListRequest = CreateListRequest
