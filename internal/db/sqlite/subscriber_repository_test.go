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
	repo := NewSubscriberRepository(db)

	ctx := context.Background()
	now := time.Now().Unix()
	testSubscriber := &domain.Subscriber{
		ID:           "test-subscriber-id",
		Email:        "test@example.com",
		Name:         "Test User",
		Status:       "active",
		SubscribedAt: now,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	err := repo.Create(ctx, testSubscriber)
	require.NoError(t, err)

	// Test GetByID
	retrievedByID, err := repo.GetByID(ctx, "test-subscriber-id")
	require.NoError(t, err)
	assert.Equal(t, "test@example.com", retrievedByID.Email)
	assert.Equal(t, "Test User", retrievedByID.Name)

	// Test GetByEmail
	retrievedByEmail, err := repo.GetByEmail(ctx, "test@example.com")
	require.NoError(t, err)
	assert.Equal(t, "test-subscriber-id", retrievedByEmail.ID)
	assert.Equal(t, "Test User", retrievedByEmail.Name)
}
