// Copyright 2025 JC-Lab
// SPDX-License-Identifier: AGPL-3.0-or-later

package sqlite

import (
	"context"
	"testing"

	"github.com/headmail/headmail/pkg/domain"
	"github.com/headmail/headmail/pkg/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestTransaction(t *testing.T) {
	gormDB := setupTestDB(t)
	db := &DB{gormDB}
	repo := NewCampaignRepository(db)

	ctx := context.Background()

	t.Run("rollback", func(t *testing.T) {
		txCtx, err := db.Begin(ctx)
		require.NoError(t, err)

		campaign := &domain.Campaign{
			ID:   "tx-campaign-1",
			Name: "TX Campaign 1",
		}
		err = repo.Create(txCtx, campaign)
		require.NoError(t, err)

		err = db.Rollback(txCtx)
		require.NoError(t, err)

		_, err = repo.GetByID(ctx, "tx-campaign-1")
		assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	})

	t.Run("commit", func(t *testing.T) {
		txCtx, err := db.Begin(ctx)
		require.NoError(t, err)

		campaign := &domain.Campaign{
			ID:   "tx-campaign-2",
			Name: "TX Campaign 2",
		}
		err = repo.Create(txCtx, campaign)
		require.NoError(t, err)

		err = db.Commit(txCtx)
		require.NoError(t, err)

		retrieved, err := repo.GetByID(ctx, "tx-campaign-2")
		require.NoError(t, err)
		assert.Equal(t, "TX Campaign 2", retrieved.Name)
	})
}

func TestCampaignRepository_List_WithTags(t *testing.T) {
	db := setupTestDB(t)
	repo := NewCampaignRepository(&DB{db})

	ctx := context.Background()

	// Create test campaigns
	campaign1 := &domain.Campaign{
		ID:   "campaign-1",
		Name: "Campaign A",
		Tags: []string{"go", "test"},
	}
	campaign2 := &domain.Campaign{
		ID:   "campaign-2",
		Name: "Campaign B",
		Tags: []string{"rust", "test"},
	}
	campaign3 := &domain.Campaign{
		ID:   "campaign-3",
		Name: "Campaign C",
		Tags: []string{"go", "dev"},
	}

	require.NoError(t, repo.Create(ctx, campaign1))
	require.NoError(t, repo.Create(ctx, campaign2))
	require.NoError(t, repo.Create(ctx, campaign3))

	t.Run("filter by single tag 'go'", func(t *testing.T) {
		filter := repository.CampaignFilter{
			Tags: []string{"go"},
		}
		pagination := repository.Pagination{Page: 1, Limit: 10}

		campaigns, total, err := repo.List(ctx, filter, pagination)
		require.NoError(t, err)
		assert.Equal(t, 2, total)
		assert.Len(t, campaigns, 2)
		assert.ElementsMatch(t, []string{"campaign-1", "campaign-3"}, []string{campaigns[0].ID, campaigns[1].ID})
	})

	t.Run("filter by single tag 'test'", func(t *testing.T) {
		filter := repository.CampaignFilter{
			Tags: []string{"test"},
		}
		pagination := repository.Pagination{Page: 1, Limit: 10}

		campaigns, total, err := repo.List(ctx, filter, pagination)
		require.NoError(t, err)
		assert.Equal(t, 2, total)
		assert.Len(t, campaigns, 2)
		assert.ElementsMatch(t, []string{"campaign-1", "campaign-2"}, []string{campaigns[0].ID, campaigns[1].ID})
	})

	t.Run("filter by multiple tags 'go' and 'test'", func(t *testing.T) {
		filter := repository.CampaignFilter{
			Tags: []string{"go", "test"},
		}
		pagination := repository.Pagination{Page: 1, Limit: 10}

		campaigns, total, err := repo.List(ctx, filter, pagination)
		require.NoError(t, err)
		assert.Equal(t, 1, total)
		assert.Len(t, campaigns, 1)
		assert.Equal(t, "campaign-1", campaigns[0].ID)
	})

	t.Run("filter by non-existing tag", func(t *testing.T) {
		filter := repository.CampaignFilter{
			Tags: []string{"python"},
		}
		pagination := repository.Pagination{Page: 1, Limit: 10}

		campaigns, total, err := repo.List(ctx, filter, pagination)
		require.NoError(t, err)
		assert.Equal(t, 0, total)
		assert.Len(t, campaigns, 0)
	})
}
