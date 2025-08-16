// Copyright 2025 JC-Lab
// SPDX-License-Identifier: AGPL-3.0-or-later

package admin

// PaginationResponse represents pagination details for a list response.
type PaginationResponse struct {
	Page  int `json:"page"`
	Total int `json:"total"`
	Limit int `json:"limit"`
}

// PaginatedListResponse is the response for a list of items with pagination.
type PaginatedListResponse[T any] struct {
	Data       []T                `json:"data"`
	Pagination PaginationResponse `json:"pagination"`
}

// DeleteResponse is the standard response for a delete operation.
type DeleteResponse struct {
	Deleted bool   `json:"deleted"`
	Message string `json:"message"`
}
