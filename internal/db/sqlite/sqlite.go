package sqlite

import (
	"context"
	"log"

	"github.com/glebarez/sqlite"
	"github.com/headmail/headmail/pkg/config"
	"github.com/headmail/headmail/pkg/repository"
	"gorm.io/gorm"
)

// DB holds the gorm database connection.
type DB struct {
	*gorm.DB
}

var _ repository.DB = (*DB)(nil)

// New opens a connection to the SQLite database using GORM.
func New(cfg config.DatabaseConfig) (*DB, error) {
	db, err := gorm.Open(sqlite.Open(cfg.URL), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Auto-migrate the schema
	if err := db.AutoMigrate(
		&List{},
		&Subscriber{},
		&Campaign{},
		&Delivery{},
		&DeliveryEvent{},
	); err != nil {
		return nil, err
	}

	log.Println("Database connection established and schema migrated.")
	return &DB{db}, nil
}

func (db *DB) ListRepository() repository.ListRepository {
	return NewListRepository(db.DB)
}

func (db *DB) SubscriberRepository() repository.SubscriberRepository {
	return NewSubscriberRepository(db.DB)
}

func (db *DB) CampaignRepository() repository.CampaignRepository {
	return NewCampaignRepository(db.DB)
}

func (db *DB) DeliveryRepository() repository.DeliveryRepository {
	return NewDeliveryRepository(db.DB)
}

func (db *DB) Begin(ctx context.Context) (context.Context, error) {
	tx := db.DB.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	return injectTx(ctx, tx), nil
}

func (db *DB) Commit(ctx context.Context) error {
	tx := extractTx(ctx, db.DB)
	return tx.Commit().Error
}

func (db *DB) Rollback(ctx context.Context) error {
	tx := extractTx(ctx, db.DB)
	return tx.Rollback().Error
}
