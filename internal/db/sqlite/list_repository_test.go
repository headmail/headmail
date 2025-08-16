// Copyright 2025 JC-Lab
// SPDX-License-Identifier: AGPL-3.0-or-later

package sqlite

import (
	"context"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/headmail/headmail/pkg/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&List{}, &Subscriber{}, &SubscriberList{}, &Campaign{}, &Delivery{}, &DeliveryEvent{})
	require.NoError(t, err)

	return db
}

func TestListRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewListRepository(&DB{db})

	ctx := context.Background()
	now := time.Now().Unix()
	testList := &domain.List{
		ID:          "test-id",
		Name:        "Test List",
		Description: "A list for testing",
		Tags:        []string{"test", "go"},
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	err := repo.Create(ctx, testList)
	require.NoError(t, err)

	var entity List
	err = db.First(&entity, "id = ?", "test-id").Error
	require.NoError(t, err)

	assert.Equal(t, "Test List", entity.Name)
	assert.Equal(t, "A list for testing", entity.Description)
	assert.Equal(t, JSON(`["test","go"]`), entity.Tags)
}
