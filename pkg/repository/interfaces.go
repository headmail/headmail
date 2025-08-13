package repository

import (
	"context"

	"github.com/headmail/headmail/pkg/domain"
	"github.com/headmail/headmail/pkg/queue"
)

// DB defines the interface for a database connection that can provide repositories.
type DB interface {
	Transactionable
	ListRepository() ListRepository
	SubscriberRepository() SubscriberRepository
	CampaignRepository() CampaignRepository
	DeliveryRepository() DeliveryRepository
	TemplateRepository() TemplateRepository
	// QueueRepository returns an implementation of a generic queue backed by the DB.
	QueueRepository() queue.Queue
	// EventRepository returns an implementation for storing delivery events (opens/clicks).
	EventRepository() EventRepository
}

// Transactionable defines the interface for transaction management.
type Transactionable interface {
	Begin(ctx context.Context) (context.Context, error)
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

// ListRepository defines the interface for list storage.
type ListRepository interface {
	Create(ctx context.Context, list *domain.List) error
	GetByID(ctx context.Context, id string) (*domain.List, error)
	Update(ctx context.Context, list *domain.List) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filter ListFilter, pagination Pagination) ([]*domain.List, int, error)
	GetSubscriberCount(ctx context.Context, listID string) (int, error)
	GetSubscribers(ctx context.Context) (chan *domain.Subscriber, error)

	// AddSubscribers adds existing subscribers to the specified list (ids).
	AddSubscribers(ctx context.Context, listID string, subscriberIDs []string) error

	// RemoveSubscribers removes subscribers from the specified list.
	RemoveSubscribers(ctx context.Context, listID string, subscriberIDs []string) error

	// ReplaceSubscribers replaces the subscribers of the given list with the provided list (atomic).
	ReplaceSubscribers(ctx context.Context, listID string, subscriberIDs []string) error
}

// SubscriberRepository defines the interface for subscriber storage.
type SubscriberRepository interface {
	Create(ctx context.Context, subscriber *domain.Subscriber) error
	GetByID(ctx context.Context, id string) (*domain.Subscriber, error)
	GetByEmail(ctx context.Context, email string) (*domain.Subscriber, error)
	Update(ctx context.Context, subscriber *domain.Subscriber) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filter SubscriberFilter, pagination Pagination) ([]*domain.Subscriber, int, error)
	ListStream(ctx context.Context, filter SubscriberFilter) (chan *domain.Subscriber, error)
	BulkUpsert(ctx context.Context, subscribers []*domain.Subscriber) error
}

// CampaignRepository defines the interface for campaign storage.
type CampaignRepository interface {
	Create(ctx context.Context, campaign *domain.Campaign) error
	GetByID(ctx context.Context, id string) (*domain.Campaign, error)
	Update(ctx context.Context, campaign *domain.Campaign) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filter CampaignFilter, pagination Pagination) ([]*domain.Campaign, int, error)
	UpdateStatus(ctx context.Context, id string, status domain.CampaignStatus) error
}

// DeliveryRepository defines the interface for delivery storage.
type DeliveryRepository interface {
	Create(ctx context.Context, delivery *domain.Delivery) error
	BulkCreate(ctx context.Context, deliveries []*domain.Delivery) error
	GetByID(ctx context.Context, id string) (*domain.Delivery, error)
	Update(ctx context.Context, delivery *domain.Delivery) error
	List(ctx context.Context, filter DeliveryFilter, pagination Pagination) ([]*domain.Delivery, int, error)
	// ListScheduledBefore returns deliveries whose scheduled_at is non-null and <= ts.
	// The result is limited by `limit`.
	ListScheduledBefore(ctx context.Context, ts int64, limit int) ([]*domain.Delivery, error)
	GetByCampaignID(ctx context.Context, campaignID string, pagination Pagination) ([]*domain.Delivery, int, error)
	UpdateStatus(ctx context.Context, id string, status string) error
}

// TemplateRepository defines the interface for template storage.
type TemplateRepository interface {
	Create(ctx context.Context, template *domain.Template) error
	GetByID(ctx context.Context, id string) (*domain.Template, error)
	Update(ctx context.Context, template *domain.Template) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, pagination Pagination) ([]*domain.Template, int, error)
}

// EventRepository defines the interface for delivery event storage (opens, clicks, etc).
type EventRepository interface {
	// Create stores a new delivery event.
	Create(ctx context.Context, event *domain.DeliveryEvent) error
	// ListByCampaignAndRange returns events for given campaign IDs within [from, to] (unix timestamps).
	ListByCampaignAndRange(ctx context.Context, campaignIDs []string, from int64, to int64) ([]*domain.DeliveryEvent, error)
	// CountByCampaignAndRange returns aggregated event counts grouped by campaign and bucket time.
	// Implementations may provide optimized aggregation helpers if needed.
	CountByCampaignAndRange(ctx context.Context, campaignIDs []string, from int64, to int64, granularity string) (map[string]map[int64]int64, error)
}

// Filter Types

type ListFilter struct {
	Search string   `json:"search,omitempty"`
	Tags   []string `json:"tags,omitempty"`
}

type SubscriberFilter struct {
	ListID     string                      `json:"list_id,omitempty"`
	ListStatus domain.SubscriberListStatus `json:"list_status,omitempty"`
	Status     domain.SubscriberStatus     `json:"status,omitempty"`
	Search     string                      `json:"search,omitempty"`
}

type CampaignFilter struct {
	Status []domain.CampaignStatus `json:"status,omitempty"`
	Search string                  `json:"search,omitempty"`
	Tags   []string                `json:"tags,omitempty"`
}

type DeliveryFilter struct {
	CampaignID string `json:"campaign_id,omitempty"`
	Type       string `json:"type,omitempty"`
	Status     string `json:"status,omitempty"`
	Email      string `json:"email,omitempty"`
}

type Pagination struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
}
