package sqlite

import (
	"context"
	"encoding/json"

	"github.com/headmail/headmail/pkg/domain"
	"github.com/headmail/headmail/pkg/repository"
)

type eventRepository struct {
	db *DB
}

func NewEventRepository(db *DB) repository.EventRepository {
	return &eventRepository{db: db}
}

func entityToDeliveryEventDomain(e *DeliveryEvent) (*domain.DeliveryEvent, error) {
	var data map[string]interface{}
	if len(e.EventData) > 0 {
		if err := json.Unmarshal(e.EventData, &data); err != nil {
			return nil, err
		}
	} else {
		data = map[string]interface{}{}
	}
	evType := domain.EventType(e.EventType)
	return &domain.DeliveryEvent{
		ID:         e.ID,
		DeliveryID: e.DeliveryID,
		EventType:  evType,
		EventData:  data,
		UserAgent:  e.UserAgent,
		IPAddress:  e.IPAddress,
		URL:        e.URL,
		CreatedAt:  e.CreatedAt,
	}, nil
}

func domainToDeliveryEventEntity(d *domain.DeliveryEvent) (*DeliveryEvent, error) {
	dataJSON, err := json.Marshal(d.EventData)
	if err != nil {
		return nil, err
	}
	return &DeliveryEvent{
		ID:         d.ID,
		DeliveryID: d.DeliveryID,
		EventType:  d.EventType,
		EventData:  dataJSON,
		UserAgent:  d.UserAgent,
		IPAddress:  d.IPAddress,
		URL:        d.URL,
		CreatedAt:  d.CreatedAt,
	}, nil
}

// Create stores a new delivery event.
func (r *eventRepository) Create(ctx context.Context, event *domain.DeliveryEvent) error {
	entity, err := domainToDeliveryEventEntity(event)
	if err != nil {
		return err
	}
	db := extractTx(ctx, r.db.DB)
	return db.WithContext(ctx).Create(entity).Error
}

// ListByCampaignAndRange returns events for given campaign IDs within [from, to] (unix timestamps).
func (r *eventRepository) ListByCampaignAndRange(ctx context.Context, campaignIDs []string, from int64, to int64) ([]*domain.DeliveryEvent, error) {
	var entities []DeliveryEvent
	db := extractTx(ctx, r.db.DB)

	// Use join to filter by campaign_id on deliveries
	query := db.WithContext(ctx).Model(&DeliveryEvent{}).
		Joins("JOIN deliveries ON deliveries.id = delivery_events.delivery_id").
		Where("deliveries.campaign_id IN ?", campaignIDs).
		Where("delivery_events.created_at >= ? AND delivery_events.created_at <= ?", from, to).
		Order("delivery_events.created_at ASC")

	if err := query.Find(&entities).Error; err != nil {
		return nil, err
	}

	var events []*domain.DeliveryEvent
	for _, e := range entities {
		ev, err := entityToDeliveryEventDomain(&e)
		if err != nil {
			return nil, err
		}
		events = append(events, ev)
	}
	return events, nil
}

// CountByCampaignAndRange returns aggregated event counts grouped by campaign and bucket time.
func (r *eventRepository) CountByCampaignAndRange(ctx context.Context, campaignIDs []string, from int64, to int64, granularity string) (map[string]map[int64]int64, error) {
	db := extractTx(ctx, r.db.DB)

	var bucketSeconds int64 = 3600
	if granularity == "day" {
		bucketSeconds = 86400
	}

	// Raw SQL: join delivery_events -> deliveries to get campaign_id, compute bucket = (created_at / bucketSeconds) * bucketSeconds
	rows, err := db.WithContext(ctx).Raw(
		`SELECT deliveries.campaign_id, ((delivery_events.created_at / ?) * ?) as bucket, COUNT(*) as cnt
		 FROM delivery_events
		 JOIN deliveries ON deliveries.id = delivery_events.delivery_id
		 WHERE deliveries.campaign_id IN ? AND delivery_events.created_at BETWEEN ? AND ?
		 GROUP BY deliveries.campaign_id, bucket`, bucketSeconds, bucketSeconds, campaignIDs, from, to).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := map[string]map[int64]int64{}
	for rows.Next() {
		var campaignID string
		var bucket int64
		var cnt int64
		if err := rows.Scan(&campaignID, &bucket, &cnt); err != nil {
			return nil, err
		}
		if _, ok := result[campaignID]; !ok {
			result[campaignID] = map[int64]int64{}
		}
		result[campaignID][bucket] = cnt
	}
	return result, nil
}
