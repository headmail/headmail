package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/smtp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/headmail/headmail/pkg/config"
	"github.com/headmail/headmail/pkg/domain"
	"github.com/headmail/headmail/pkg/queue"
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
}

// DeliveryService provides business logic for delivery management.
type DeliveryService struct {
	db      repository.DB
	repo    repository.DeliveryRepository
	queue   queue.Queue
	smtpCfg config.SMTPConfig
}

// NewDeliveryService creates a new DeliveryService.
func NewDeliveryService(db repository.DB, q queue.Queue, smtpCfg config.SMTPConfig) *DeliveryService {
	return &DeliveryService{
		db:      db,
		repo:    db.DeliveryRepository(),
		queue:   q,
		smtpCfg: smtpCfg,
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

	err = s.sendMail(d)
	now := time.Now().Unix()
	if err != nil {
		d.FailedAt = &now
		d.Attempts++
		if d.Attempts >= s.smtpCfg.Send.Attempts {
			d.Status = domain.DeliveryStatusFailed
		} else {
			d.Status = domain.DeliveryStatusScheduled
			nextScheduledAt := now + 300 // 5 minutes
			d.SendScheduledAt = &nextScheduledAt
		}

		log.Printf("worker %s: smtp send failed for delivery %s: %v", workerID, d.ID, err)
	} else {
		d.SentAt = &now
		d.Status = domain.DeliveryStatusSent
	}

	if err := s.repo.Update(ctx, d); err != nil {
		return err
	}

	return nil
}

func (s *DeliveryService) sendMail(d *domain.Delivery) error {
	// Build email message
	fromHeader := fmt.Sprintf("%s <%s>", s.smtpCfg.From.Name, s.smtpCfg.From.Email)
	toHeader := d.Email
	subject := d.Subject
	var body string
	var contentType string
	if d.BodyText != "" && d.BodyHTML != "" {
		// simple prefer HTML
		contentType = "text/html; charset=\"utf-8\""
		body = d.BodyHTML
	} else if d.BodyHTML != "" {
		contentType = "text/html; charset=\"utf-8\""
		body = d.BodyHTML
	} else {
		contentType = "text/plain; charset=\"utf-8\""
		body = d.BodyText
	}

	headers := make(map[string]string)
	headers["From"] = fromHeader
	headers["To"] = toHeader
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = contentType

	var msg strings.Builder
	for k, v := range headers {
		msg.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	msg.WriteString("\r\n")
	msg.WriteString(body)

	// Prepare SMTP auth if credentials provided
	addr := fmt.Sprintf("%s:%d", s.smtpCfg.Host, s.smtpCfg.Port)
	var auth smtp.Auth
	if s.smtpCfg.Username != "" {
		auth = smtp.PlainAuth("", s.smtpCfg.Username, s.smtpCfg.Password, s.smtpCfg.Host)
	}

	// Attempt to send email (this is the actual send step; errors will be returned to caller)
	if err := smtp.SendMail(addr, auth, s.smtpCfg.From.Email, []string{d.Email}, []byte(msg.String())); err != nil {
		return err
	}

	return nil
}
