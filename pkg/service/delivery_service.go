// Copyright 2025 JC-Lab
// SPDX-License-Identifier: AGPL-3.0-or-later

package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/headmail/headmail/pkg/template"
	"golang.org/x/net/html"

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
	UpdateDeliveryStatus(ctx context.Context, id string, status domain.DeliveryStatus) error

	EnqueueDelivery(txCtx context.Context, delivery *domain.Delivery) error

	// Handle queued work
	HandleDeliveryQueuedItem(ctx context.Context, workerID string, item *queue.QueueItem) error

	// Handle bounced mail from receiver
	HandleBouncedMail(ctx context.Context, event *receiver.Event) error

	// SendNow performs an immediate synchronous send attempt for the specified delivery ID.
	SendNow(ctx context.Context, deliveryID string) (*domain.Delivery, error)

	// Retry performs an immediate retry of the specified delivery (resets attempts/flags and sends now).
	Retry(ctx context.Context, deliveryID string) (*domain.Delivery, error)
}

// DeliveryService provides business logic for delivery management.
type DeliveryService struct {
	db              repository.DB
	templateService *template.Service
	repo            repository.DeliveryRepository
	eventRepo       repository.EventRepository
	queue           queue.Queue
	mailer          mailer.Mailer
	trackingHost    string
	maxAttempts     int
}

// NewDeliveryService creates a new DeliveryService.
func NewDeliveryService(db repository.DB, templateService *template.Service, q queue.Queue, m mailer.Mailer, trackingHost string, maxAttempts int) *DeliveryService {
	return &DeliveryService{
		db:              db,
		templateService: templateService,
		repo:            db.DeliveryRepository(),
		eventRepo:       db.EventRepository(),
		queue:           q,
		mailer:          m,
		trackingHost:    trackingHost,
		maxAttempts:     maxAttempts,
	}
}

// CreateDelivery creates a new delivery and enqueues immediate deliveries.
func (s *DeliveryService) CreateDelivery(ctx context.Context, delivery *domain.Delivery) error {
	var err error

	delivery.ID = uuid.New().String()

	// prepare delivery
	if delivery.CreatedAt == 0 {
		delivery.CreatedAt = time.Now().Unix()
	}
	// default status
	if delivery.Status == "" {
		return errors.New("invalid status")
	}

	// Prepare data for templating
	templateData := make(map[string]interface{})
	if delivery.Data != nil {
		for k, v := range delivery.Data {
			templateData[k] = v
		}
	}
	templateData["deliveryId"] = delivery.ID
	templateData["name"] = delivery.Name
	templateData["email"] = delivery.Email

	// Render subject
	delivery.Subject, err = s.templateService.Render(delivery.Subject, templateData)
	if err != nil {
		return err
	}

	// Render HTML content
	delivery.BodyHTML, err = s.templateService.Render(delivery.BodyHTML, templateData)
	if err != nil {
		return err
	}

	// Render text content
	delivery.BodyText, err = s.templateService.Render(delivery.BodyText, templateData)
	if err != nil {
		return err
	}

	return repository.Transactional0(s.db, ctx, func(txCtx context.Context) error {
		// create delivery
		if err := s.repo.Create(txCtx, delivery); err != nil {
			return err
		}

		if delivery.Status == domain.DeliveryStatusScheduled && delivery.ScheduledAt == nil {
			return s.EnqueueDelivery(txCtx, delivery)
		}

		return nil
	})
}

func (s *DeliveryService) EnqueueDelivery(txCtx context.Context, delivery *domain.Delivery) error {
	var err error
	data := &deliveryQueueData{
		DeliveryID: delivery.ID,
	}
	unique := fmt.Sprintf("delivery:%s:%d", delivery.ID, delivery.Attempts)
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
	return nil
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
func (s *DeliveryService) UpdateDeliveryStatus(ctx context.Context, id string, status domain.DeliveryStatus) error {
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
			d.ScheduledAt = &nextScheduledAt
		}

		log.Printf("worker %s: mail send failed for delivery %s: %v", workerID, d.ID, err)
	} else {
		d.SentAt = &now
		d.Status = domain.DeliveryStatusSent
	}

	if err := s.repo.Update(ctx, d); err != nil {
		return err
	}

	if d.CampaignID != nil && *d.CampaignID != "" {
		if d.Status == domain.DeliveryStatusSent {
			// delivered: transition to Sent
			if err := s.db.CampaignRepository().IncrementStats(ctx, *d.CampaignID, 0, 1, 0, 0, 0, 0); err != nil {
				log.Printf("worker %s: failed to increment campaign delivered count for %s: %v", workerID, *d.CampaignID, err)
			}
		} else if d.Status == domain.DeliveryStatusFailed {
			// failed: transition to Failed
			if err := s.db.CampaignRepository().IncrementStats(ctx, *d.CampaignID, 0, 0, 1, 0, 0, 0); err != nil {
				log.Printf("worker %s: failed to increment campaign failed count for %s: %v", workerID, *d.CampaignID, err)
			}
		}
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
	isFirstBounce, err := s.repo.IncrementCount(ctx, data.DeliveryID, domain.EventTypeBounced)
	if err != nil {
		log.Printf("imap receiver: failed to increment bounce count: %v", err)
	} else if isFirstBounce {
		// If this is the first bounce for the delivery, increment campaign-level bounce counter
		if d, derr := s.repo.GetByID(ctx, data.DeliveryID); derr == nil {
			if d.CampaignID != nil && *d.CampaignID != "" {
				if cerr := s.db.CampaignRepository().IncrementStats(ctx, *d.CampaignID, 0, 0, 0, 0, 0, 1); cerr != nil {
					log.Printf("imap receiver: failed to increment campaign bounce count for %s: %v", *d.CampaignID, cerr)
				}
			}
		} else {
			log.Printf("imap receiver: failed to load delivery %s for campaign bounce increment: %v", data.DeliveryID, derr)
		}
	}

	// write synchronously
	if err := s.eventRepo.Create(ctx, ev); err != nil {
		log.Printf("imap receiver: failed to save event: %v", err)
	}

	return s.UpdateDeliveryStatus(ctx, data.DeliveryID, domain.DeliveryStatusBounced)
}

/*
SendNow performs an immediate synchronous send attempt for the specified delivery ID.
It will perform the send in-process (not via queue), update delivery status and campaign counters
similarly to the worker logic.
*/
func (s *DeliveryService) SendNow(ctx context.Context, deliveryID string) (*domain.Delivery, error) {
	// load delivery
	d, err := s.repo.GetByID(ctx, deliveryID)
	if err != nil {
		return nil, err
	}

	prevStatus := d.Status

	// inject tracking before sending
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
			nextScheduledAt := now + 300
			d.ScheduledAt = &nextScheduledAt
		}
	} else {
		d.SentAt = &now
		d.Status = domain.DeliveryStatusSent
	}

	// persist delivery changes
	if perr := s.repo.Update(ctx, d); perr != nil {
		return nil, perr
	}

	// update campaign-level counters only on state transitions
	if d.CampaignID != nil && *d.CampaignID != "" {
		cid := *d.CampaignID
		if prevStatus != domain.DeliveryStatusSent && d.Status == domain.DeliveryStatusSent {
			if cerr := s.db.CampaignRepository().IncrementStats(ctx, cid, 0, 1, 0, 0, 0, 0); cerr != nil {
				log.Printf("SendNow: failed to increment campaign delivered count for %s: %v", cid, cerr)
			}
		}
		if prevStatus != domain.DeliveryStatusFailed && d.Status == domain.DeliveryStatusFailed {
			if cerr := s.db.CampaignRepository().IncrementStats(ctx, cid, 0, 0, 1, 0, 0, 0); cerr != nil {
				log.Printf("SendNow: failed to increment campaign failed count for %s: %v", cid, cerr)
			}
		}
	}

	return d, err
}

/*
Retry resets attempt-related fields and performs an immediate synchronous send attempt.
*/
func (s *DeliveryService) Retry(ctx context.Context, deliveryID string) (*domain.Delivery, error) {
	// load delivery
	d, err := s.repo.GetByID(ctx, deliveryID)
	if err != nil {
		return nil, err
	}

	// reset attempt metadata so send can be retried immediately
	d.Attempts = 0
	d.FailedAt = nil
	d.FailureReason = nil
	d.ScheduledAt = nil
	// mark as scheduled so SendNow logic treats it appropriately
	d.Status = domain.DeliveryStatusScheduled

	if err := s.repo.Update(ctx, d); err != nil {
		return nil, err
	}

	// perform immediate synchronous send
	return s.SendNow(ctx, deliveryID)
}

// sendMail - helper retained for compatibility with other code paths
func (s *DeliveryService) sendMail(d *domain.Delivery) error {
	// Delegate to configured mailer implementation.
	// Mailer implementations are responsible for building headers/body and sending.
	return s.mailer.Send(context.Background(), d)
}

// injectTracking rewrites only anchor tag href attributes in the provided HTML
// to route through the click tracker and appends an open-tracking 1x1 image.
// Uses an HTML parser to modify only <a> href attributes.
func (s *DeliveryService) injectTracking(deliveryID, htmlStr string) string {
	// parse HTML document
	doc, err := html.Parse(strings.NewReader(htmlStr))
	if err != nil {
		// fallback: append pixel and return original
		return s.appendPixel(htmlStr, deliveryID)
	}

	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && strings.EqualFold(n.Data, "a") {
			for i, a := range n.Attr {
				if strings.EqualFold(a.Key, "href") {
					orig := a.Val
					lower := strings.ToLower(orig)
					if !strings.HasPrefix(lower, "http") || strings.Contains(orig, "/r/"+deliveryID+"/c") {
						continue
					}
					encoded := url.QueryEscape(orig)
					trackingURL := s.trackingHost
					if !strings.HasPrefix(trackingURL, "http://") && !strings.HasPrefix(trackingURL, "https://") {
						trackingURL = "https://" + strings.TrimRight(trackingURL, "/")
					}
					newHref := trackingURL + "/r/" + deliveryID + "/c?u=" + encoded
					n.Attr[i].Val = newHref
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}

	walk(doc)

	var buf bytes.Buffer
	if err := html.Render(&buf, doc); err != nil {
		return s.appendPixel(htmlStr, deliveryID)
	}
	out := buf.String()
	// append pixel
	return s.appendPixel(out, deliveryID)
}

// appendPixel appends the tracking pixel before </body> if present, otherwise at end.
func (s *DeliveryService) appendPixel(htmlStr, deliveryID string) string {
	pixelURL := s.trackingHost
	if !strings.HasPrefix(pixelURL, "http://") && !strings.HasPrefix(pixelURL, "https://") {
		pixelURL = "https://" + strings.TrimRight(pixelURL, "/")
	}
	pixel := `<img src="` + pixelURL + `/r/` + deliveryID + `/o" width="1" height="1" style="display:none" alt="">`
	lower := strings.ToLower(htmlStr)
	if idx := strings.LastIndex(lower, "</body>"); idx != -1 {
		return htmlStr[:idx] + pixel + htmlStr[idx:]
	}
	return htmlStr + pixel
}
