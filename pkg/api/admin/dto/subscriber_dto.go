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

// PatchSubscribersRequest is used to add/remove subscribers from a list.
// add: list of subscriber IDs to add
// remove: list of subscriber IDs to remove
type PatchSubscribersRequest struct {
	Add    []string `json:"add,omitempty"`
	Remove []string `json:"remove,omitempty"`
}

// ReplaceSubscribersRequest replaces the list members with the provided subscriber IDs.
type ReplaceSubscribersRequest struct {
	Subscribers []string `json:"subscribers"`
}

// CreateSubscribersRequest is the request for creating multiple subscribers (used when adding to a list).
type CreateSubscribersRequest struct {
	Subscribers []*CreateSubscriberRequest `json:"subscribers"`
	Append      bool                       `json:"append"`
}

// EmptyResponse represents an empty response body.
type EmptyResponse struct{}
