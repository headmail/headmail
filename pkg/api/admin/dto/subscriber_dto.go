package dto

import "github.com/headmail/headmail/pkg/domain"

// CreateSubscriberRequest is the request for creating a subscriber.
type CreateSubscriberRequest struct {
	Email  string                   `json:"email"`
	Name   string                   `json:"name"`
	Status *domain.SubscriberStatus `json:"status,omitempty"`
}

// UpdateSubscriberRequest is the request for updating a subscriber.
type UpdateSubscriberRequest = CreateSubscriberRequest
