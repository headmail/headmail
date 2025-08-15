package service

import (
	"context"
	"encoding/json"
	"log"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/headmail/headmail/pkg/domain"
	"github.com/headmail/headmail/pkg/mailer"
	"github.com/headmail/headmail/pkg/queue"
	"github.com/headmail/headmail/pkg/receiver"
	"github.com/headmail/headmail/pkg/repository"
)

type deliveryQueueData struct {
	DeliveryID string `json:"delivery_id"`
}

// DeliveryServiceProvider defines the interface for a delivery service.
type DeliveryServiceProvider interface {
	CreateDelivery(ctx context.Context, delivery *domain.Delivery) error
	GetDelivery(ctx context.Context, id string) (*domain.Delivery, error)
	UpdateDelivery(ctx context.Context, delivery *domain.Delivery) error
	ListDeliveries(ctx context.Context, filter repository.DeliveryFilter, pagination repository.Pagination) ([]*domain.Delivery, int, error)
	GetDeliveriesByCampaign(ctx context.Context, campaignID string, pagination repository.Pagination) ([]*domain.Delivery, int, error)
	UpdateDeliveryStatus(ctx context.Context, id string, status string) error

	HandleDeliveryQueuedItem(ctx context.Context, workerID string, item *queue.QueueItem) error

	HandleBouncedMail(ctx context.Context, event *receiver.Event) error
}

// DeliveryService provides business logic for delivery management.
type DeliveryService struct {
	db           repository.DB
	repo         repository.DeliveryRepository
	eventRepo    repository.EventRepository
	queue        queue.Queue
	mailer       mailer.Mailer
	trackingHost string
	maxAttempts  int
}

// NewDeliveryService creates a new DeliveryService.
func NewDeliveryService(db repository.DB, q queue.Queue, m mailer.Mailer, trackingHost string, maxAttempts int) *DeliveryService {
	return &DeliveryService{
		db:           db,
		repo:         db.DeliveryRepository(),
		eventRepo:    db.EventRepository(),
		queue:        q,
		mailer:       m,
		trackingHost: trackingHost,
		maxAttempts:  maxAttempts,
	}
}

// CreateDelivery creates a new delivery and enqueues immediate deliveries.
func (s *DeliveryService) CreateDelivery(ctx context.Context, delivery *domain.Delivery) error {
	return repository.Transactional0(s.db, ctx, func(txCtx context.Context) error {
		// prepare delivery
		delivery.ID = uuid.New().String()
		if delivery.CreatedAt == 0 {
			delivery.CreatedAt = time.Now().Unix()
		}
		// default status
		if delivery.Status == "" {
			delivery.Status = domain.DeliveryStatusScheduled
		}

		// create delivery
		if err := s.repo.Create(txCtx, delivery); err != nil {
			return err
		}

		// if no scheduled_at then enqueue immediately within the same transaction
		if delivery.ScheduledAt == nil {
			var err error
			data := &deliveryQueueData{
				DeliveryID: delivery.ID,
			}
			unique := "delivery:" + delivery.ID
			item := &queue.QueueItem{
				ID:        uuid.New().String(),
				Type:      "delivery",
				UniqueKey: &unique,
				Status:    "pending",
				CreatedAt: time.Now().Unix(),
			}
			item.Payload, err = json.Marshal(data)
			if err != nil {
				return err
			}
			if err := s.queue.Enqueue(txCtx, item); err != nil {
				return err
			}
		}

		return nil
	})
}

// GetDelivery retrieves a delivery by its ID.
func (s *DeliveryService) GetDelivery(ctx context.Context, id string) (*domain.Delivery, error) {
	return s.repo.GetByID(ctx, id)
}

// UpdateDelivery updates an existing delivery.
func (s *DeliveryService) UpdateDelivery(ctx context.Context, delivery *domain.Delivery) error {
	return s.repo.Update(ctx, delivery)
}

// ListDeliveries lists all deliveries.
func (s *DeliveryService) ListDeliveries(ctx context.Context, filter repository.DeliveryFilter, pagination repository.Pagination) ([]*domain.Delivery, int, error) {
	return s.repo.List(ctx, filter, pagination)
}

// GetDeliveriesByCampaign retrieves all deliveries for a campaign.
func (s *DeliveryService) GetDeliveriesByCampaign(ctx context.Context, campaignID string, pagination repository.Pagination) ([]*domain.Delivery, int, error) {
	return s.repo.GetByCampaignID(ctx, campaignID, pagination)
}

// UpdateDeliveryStatus updates the status of a delivery.
func (s *DeliveryService) UpdateDeliveryStatus(ctx context.Context, id string, status string) error {
	return s.repo.UpdateStatus(ctx, id, status)
}

func (s *DeliveryService) HandleDeliveryQueuedItem(ctx context.Context, workerID string, item *queue.QueueItem) error {
	var payload deliveryQueueData
	if err := json.Unmarshal(item.Payload, &payload); err != nil {
		return err
	}

	// load delivery using repository (respects tx in ctx if present)
	d, err := s.repo.GetByID(ctx, payload.DeliveryID)
	if err != nil {
		// if not found, nothing to do
		log.Printf("worker %s: delivery %s not found: %v", workerID, payload.DeliveryID, err)
		return nil
	}

	// inject tracking into HTML before sending (rewrite links + add tracking pixel)
	if d.BodyHTML != "" && s.trackingHost != "" {
		d.BodyHTML = s.injectTracking(d.ID, d.BodyHTML)
	}
	err = s.mailer.Send(ctx, d)
	now := time.Now().Unix()
	if err != nil {
		d.FailedAt = &now
		d.Attempts++
		if d.Attempts >= s.maxAttempts {
			d.Status = domain.DeliveryStatusFailed
		} else {
			d.Status = domain.DeliveryStatusScheduled
			nextScheduledAt := now + 300 // 5 minutes
			d.SendScheduledAt = &nextScheduledAt
		}

		log.Printf("worker %s: mail send failed for delivery %s: %v", workerID, d.ID, err)
	} else {
		d.SentAt = &now
		d.Status = domain.DeliveryStatusSent
	}

	if err := s.repo.Update(ctx, d); err != nil {
		return err
	}

	return nil
}

func (s *DeliveryService) HandleBouncedMail(ctx context.Context, data *receiver.Event) error {
	now := time.Now()
	ev := &domain.DeliveryEvent{
		DeliveryID: data.DeliveryID,
		CreatedAt:  now.Unix(),
		EventType:  domain.EventTypeBounced,
		EventData: map[string]interface{}{
			"recipients": data.BouncedRecipients,
			"subject":    data.Subject,
			"message_id": data.MessageID,
			"reason":     data.Reason,
		},
	}

	// Atomically increment bounce count (repository will handle race conditions)
	if err := s.repo.IncrementCount(ctx, data.DeliveryID, domain.EventTypeBounced); err != nil {
		log.Printf("imap receiver: failed to increment bounce count: %v", err)
		// continue to record event even if increment failed
	}

	// write synchronously
	if err := s.eventRepo.Create(ctx, ev); err != nil {
		log.Printf("imap receiver: failed to save event: %v", err)
	}

	return nil
}

func (s *DeliveryService) sendMail(d *domain.Delivery) error {
	// Delegate to configured mailer implementation.
	// Mailer implementations are responsible for building headers/body and sending.
	return s.mailer.Send(context.Background(), d)
}

// injectTracking rewrites all links in the provided HTML to route through the click
// tracker and appends an open-tracking 1x1 image pointing to the open tracker.
func (s *DeliveryService) injectTracking(deliveryID, html string) string {
	// rewrite links: href="..." -> href="https://{host}/r/{deliveryID}/c?u={urlencoded}"
	re := regexp.MustCompile(`(?i)href="([^"#]+[^"]*)"`)
	newHTML := re.ReplaceAllStringFunc(html, func(m string) string {
		matches := re.FindStringSubmatch(m)
		if len(matches) < 2 {
			return m
		}
		orig := matches[1]
		encoded := url.QueryEscape(orig)
		trackingURL := s.trackingHost
		// ensure scheme present in trackingHost
		if !strings.HasPrefix(trackingURL, "http://") && !strings.HasPrefix(trackingURL, "https://") {
			trackingURL = "https://" + strings.TrimRight(trackingURL, "/")
		}
		newHref := trackingURL + "/r/" + deliveryID + "/c?u=" + encoded
		return `href="` + newHref + `"`
	})

	// append tracking pixel before </body> or at end
	pixelURL := s.trackingHost
	if !strings.HasPrefix(pixelURL, "http://") && !strings.HasPrefix(pixelURL, "https://") {
		pixelURL = "https://" + strings.TrimRight(pixelURL, "/")
	}
	pixel := `<img src="` + pixelURL + `/r/` + deliveryID + `/o" width="1" height="1" style="display:none" alt="">`

	if idx := strings.LastIndex(strings.ToLower(newHTML), "</body>"); idx != -1 {
		return newHTML[:idx] + pixel + newHTML[idx:]
	}
	return newHTML + pixel
}
