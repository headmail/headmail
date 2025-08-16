// Copyright 2025 JC-Lab
// SPDX-License-Identifier: AGPL-3.0-or-later

package sqlite

import (
	"context"
	"testing"
	"time"

	"github.com/headmail/headmail/pkg/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSubscriberRepository_CreateAndGet(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSubscriberRepository(&DB{db})

	ctx := context.Background()
	now := time.Now().Unix()
	testSubscriber := &domain.Subscriber{
		ID:     "test-subscriber-id",
		Email:  "test@example.com",
		Name:   "Test User",
		Status: domain.SubscriberStatusEnabled,
		Lists: []domain.SubscriberList{
			{
				ListID:       "test-list-id",
				Status:       domain.SubscriberListStatusConfirmed,
				SubscribedAt: &now,
			},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := repo.Create(ctx, testSubscriber)
	require.NoError(t, err)

	// Test GetByID
	retrievedByID, err := repo.GetByID(ctx, "test-subscriber-id")
	require.NoError(t, err)
	assert.Equal(t, "test@example.com", retrievedByID.Email)
	assert.Equal(t, "Test User", retrievedByID.Name)
	require.Len(t, retrievedByID.Lists, 1)
	assert.Equal(t, "test-list-id", retrievedByID.Lists[0].ListID)

	// Test GetByEmail
	retrievedByEmail, err := repo.GetByEmail(ctx, "test@example.com")
	require.NoError(t, err)
	assert.Equal(t, "test-subscriber-id", retrievedByEmail.ID)
	assert.Equal(t, "Test User", retrievedByEmail.Name)
	require.Len(t, retrievedByEmail.Lists, 1)
	assert.Equal(t, "test-list-id", retrievedByEmail.Lists[0].ListID)
}
