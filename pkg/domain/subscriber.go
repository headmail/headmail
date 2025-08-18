// Copyright 2025 JC-Lab
// SPDX-License-Identifier: AGPL-3.0-or-later

package domain

type SubscriberStatus string

const (
	SubscriberStatusEnabled  SubscriberStatus = "enabled"
	SubscriberStatusDisabled SubscriberStatus = "disabled"
	SubscriberStatusDeleted  SubscriberStatus = "deleted"
)

type SubscriberListStatus string

const (
	SubscriberListStatusConfirmed    SubscriberListStatus = "confirmed"
	SubscriberListStatusUnsubscribed SubscriberListStatus = "unsubscribed"
	SubscriberListStatusBounced      SubscriberListStatus = "bounced"
	SubscriberListStatusComplained   SubscriberListStatus = "complained"
)

// Subscriber represents a unique subscriber
type Subscriber struct {
	ID        string           `json:"id"`         // UUID
	CreatedAt int64            `json:"created_at"` // Unix timestamp in seconds
	UpdatedAt int64            `json:"updated_at"` // Unix timestamp in seconds
	Email     string           `json:"email"`      // Unique email address
	Name      string           `json:"name"`       // Name of the subscriber
	Status    SubscriberStatus `json:"status"`
	Lists     []SubscriberList `json:"lists"`
}

type SubscriberList struct {
	ListID         string               `json:"list_id"`    // Foreign key to the List
	CreatedAt      int64                `json:"created_at"` // Unix timestamp in seconds
	UpdatedAt      int64                `json:"updated_at"` // Unix timestamp in seconds
	Status         SubscriberListStatus `json:"status"`
	SubscribedAt   *int64               `json:"subscribed_at"`   // Unix timestamp of subscription
	UnsubscribedAt *int64               `json:"unsubscribed_at"` // Unix timestamp of unsubscription
}
