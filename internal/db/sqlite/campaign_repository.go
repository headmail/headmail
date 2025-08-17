// Copyright 2025 JC-Lab
// SPDX-License-Identifier: AGPL-3.0-or-later

package sqlite

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/headmail/headmail/pkg/domain"
	"github.com/headmail/headmail/pkg/repository"
)

type campaignRepository struct {
	db *DB
}

func NewCampaignRepository(db *DB) repository.CampaignRepository {
	return &campaignRepository{db: db}
}

func domainToCampaignEntity(d *domain.Campaign) (*Campaign, error) {
	dataJSON, err := json.Marshal(d.Data)
	if err != nil {
		return nil, err
	}
	tagsJSON, err := json.Marshal(d.Tags)
	if err != nil {
		return nil, err
	}
	headersJSON, err := json.Marshal(d.Headers)
	if err != nil {
		return nil, err
	}
	utmParamsJSON, err := json.Marshal(d.UTMParams)
	if err != nil {
		return nil, err
	}

	return &Campaign{
		ID:             d.ID,
		Name:           d.Name,
		Status:         d.Status,
		FromName:       d.FromName,
		FromEmail:      d.FromEmail,
		Subject:        d.Subject,
		TemplateID:     d.TemplateID,
		TemplateHTML:   d.TemplateHTML,
		TemplateText:   d.TemplateText,
		Data:           dataJSON,
		Tags:           tagsJSON,
		Headers:        headersJSON,
		UTMParams:      utmParamsJSON,
		ScheduledAt:    d.ScheduledAt,
		SentAt:         d.SentAt,
		CreatedAt:      d.CreatedAt,
		UpdatedAt:      d.UpdatedAt,
		DeletedAt:      d.DeletedAt,
		RecipientCount: d.RecipientCount,
		DeliveredCount: d.DeliveredCount,
		FailedCount:    d.FailedCount,
		OpenCount:      d.OpenCount,
		ClickCount:     d.ClickCount,
		BounceCount:    d.BounceCount,
	}, nil
}

func entityToCampaignDomain(e *Campaign) (*domain.Campaign, error) {
	var data map[string]interface{}
	if err := json.Unmarshal(e.Data, &data); err != nil {
		return nil, err
	}
	var tags []string
	if err := json.Unmarshal(e.Tags, &tags); err != nil {
		return nil, err
	}
	var headers map[string]string
	if err := json.Unmarshal(e.Headers, &headers); err != nil {
		return nil, err
	}
	var utmParams map[string]string
	if err := json.Unmarshal(e.UTMParams, &utmParams); err != nil {
		return nil, err
	}

	return &domain.Campaign{
		ID:             e.ID,
		Name:           e.Name,
		Status:         e.Status,
		FromName:       e.FromName,
		FromEmail:      e.FromEmail,
		Subject:        e.Subject,
		TemplateID:     e.TemplateID,
		TemplateHTML:   e.TemplateHTML,
		TemplateText:   e.TemplateText,
		Data:           data,
		Tags:           tags,
		Headers:        headers,
		UTMParams:      utmParams,
		ScheduledAt:    e.ScheduledAt,
		SentAt:         e.SentAt,
		CreatedAt:      e.CreatedAt,
		UpdatedAt:      e.UpdatedAt,
		DeletedAt:      e.DeletedAt,
		RecipientCount: e.RecipientCount,
		DeliveredCount: e.DeliveredCount,
		FailedCount:    e.FailedCount,
		OpenCount:      e.OpenCount,
		ClickCount:     e.ClickCount,
		BounceCount:    e.BounceCount,
	}, nil
}

func (r *campaignRepository) Create(ctx context.Context, campaign *domain.Campaign) error {
	entity, err := domainToCampaignEntity(campaign)
	if err != nil {
		return err
	}
	db := extractTx(ctx, r.db.DB)
	return db.WithContext(ctx).Create(entity).Error
}

func (r *campaignRepository) GetByID(ctx context.Context, id string) (*domain.Campaign, error) {
	var entity Campaign
	db := extractTx(ctx, r.db.DB)
	if err := db.WithContext(ctx).First(&entity, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return entityToCampaignDomain(&entity)
}

func (r *campaignRepository) Update(ctx context.Context, campaign *domain.Campaign) error {
	entity, err := domainToCampaignEntity(campaign)
	if err != nil {
		return err
	}
	db := extractTx(ctx, r.db.DB)

	// Only update the allowed fields to avoid overwriting other columns.
	updates := map[string]interface{}{
		"name":          entity.Name,
		"status":        entity.Status,
		"from_name":     entity.FromName,
		"from_email":    entity.FromEmail,
		"subject":       entity.Subject,
		"template_id":   entity.TemplateID,
		"template_html": entity.TemplateHTML,
		"template_text": entity.TemplateText,
		"data":          entity.Data,
		"tags":          entity.Tags,
		"headers":       entity.Headers,
		"utm_params":    entity.UTMParams,
		"scheduled_at":  entity.ScheduledAt,
		"updated_at":    entity.UpdatedAt,
	}

	return db.WithContext(ctx).Model(&Campaign{}).Where("id = ?", entity.ID).Updates(updates).Error
}

func (r *campaignRepository) Delete(ctx context.Context, id string) error {
	db := extractTx(ctx, r.db.DB)
	return db.WithContext(ctx).Delete(&Campaign{}, "id = ?", id).Error
}

func (r *campaignRepository) List(ctx context.Context, filter repository.CampaignFilter, pagination repository.Pagination) ([]*domain.Campaign, int, error) {
	var entities []Campaign
	var total int64

	db := extractTx(ctx, r.db.DB)
	query := db.WithContext(ctx).Model(&Campaign{})

	if len(filter.Status) > 0 {
		query = query.Where("status IN (?)", filter.Status)
	}
	if filter.Search != "" {
		query = query.Where("name LIKE ?", "%"+filter.Search+"%")
	}
	if len(filter.Tags) > 0 {
		for _, tag := range filter.Tags {
			query = query.Where("EXISTS (SELECT 1 FROM json_each(tags) WHERE value = ?)", tag)
		}
	}
	// TODO: json_each(frm.tags) where json_each.value where (?)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (pagination.Page - 1) * pagination.Limit
	if err := query.Order("created_at DESC").Offset(offset).Limit(pagination.Limit).Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	var campaigns []*domain.Campaign
	for _, entity := range entities {
		campaign, err := entityToCampaignDomain(&entity)
		if err != nil {
			return nil, 0, err // Or handle more gracefully
		}
		campaigns = append(campaigns, campaign)
	}

	return campaigns, int(total), nil
}

func (r *campaignRepository) UpdateStatus(ctx context.Context, id string, status domain.CampaignStatus) error {
	db := extractTx(ctx, r.db.DB)
	return db.WithContext(ctx).Model(&Campaign{}).Where("id = ?", id).Update("status", status).Error
}

// IncrementStats atomically increments per-campaign counters.
// Provide deltas for fields you want to change; pass 0 for no-op.
func (r *campaignRepository) IncrementStats(ctx context.Context, id string, recipientDelta int, deliveredDelta int, failedDelta int, openDelta int, clickDelta int, bounceDelta int) error {
	db := extractTx(ctx, r.db.DB)

	sets := make([]string, 0)
	args := make([]interface{}, 0)

	if recipientDelta != 0 {
		sets = append(sets, "recipient_count = recipient_count + ?")
		args = append(args, recipientDelta)
	}
	if deliveredDelta != 0 {
		sets = append(sets, "delivered_count = delivered_count + ?")
		args = append(args, deliveredDelta)
	}
	if failedDelta != 0 {
		sets = append(sets, "failed_count = failed_count + ?")
		args = append(args, failedDelta)
	}
	if openDelta != 0 {
		sets = append(sets, "open_count = open_count + ?")
		args = append(args, openDelta)
	}
	if clickDelta != 0 {
		sets = append(sets, "click_count = click_count + ?")
		args = append(args, clickDelta)
	}
	if bounceDelta != 0 {
		sets = append(sets, "bounce_count = bounce_count + ?")
		args = append(args, bounceDelta)
	}

	if len(sets) == 0 {
		return nil
	}

	query := "UPDATE campaigns SET " + strings.Join(sets, ", ") + " WHERE id = ?"
	args = append(args, id)

	return db.WithContext(ctx).Exec(query, args...).Error
}

// ListScheduledBefore returns campaigns whose scheduled_at is non-null and <= ts.
func (r *campaignRepository) ListScheduledBefore(ctx context.Context, ts int64) ([]*domain.Campaign, error) {
	var entities []Campaign
	db := extractTx(ctx, r.db.DB)
	query := db.WithContext(ctx).Model(&Campaign{}).
		Where("status = ? AND (scheduled_at <= ? OR scheduled_at IS NULL)", domain.CampaignStatusScheduled, ts)

	if err := query.Find(&entities).Error; err != nil {
		return nil, err
	}

	var campaigns []*domain.Campaign
	for _, e := range entities {
		c, err := entityToCampaignDomain(&e)
		if err != nil {
			return nil, err
		}
		campaigns = append(campaigns, c)
	}

	return campaigns, nil
}
