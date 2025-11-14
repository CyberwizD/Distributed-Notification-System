package repository

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// NotificationStatus mirrors the schema used by the API gateway.
type NotificationStatus struct {
	RequestID string `gorm:"primaryKey"`
	Status    string
	UpdatedAt time.Time
	Provider  string
	Detail    string
}

type StatusStore struct {
	db        *gorm.DB
	tableName string
}

func NewStatusStore(db *gorm.DB, tableName string) *StatusStore {
	if tableName == "" {
		tableName = "notification_statuses"
	}

	if err := db.Table(tableName).AutoMigrate(&NotificationStatus{}); err != nil {
		// AutoMigrate error is ignored here to keep constructor signature simple.
		// The caller is expected to have validated connectivity beforehand.
	}

	return &StatusStore{
		db:        db,
		tableName: tableName,
	}
}

func (s *StatusStore) UpdateStatus(ctx context.Context, requestID, status, provider, detail string) error {
	ns := NotificationStatus{
		RequestID: requestID,
		Status:    status,
		UpdatedAt: time.Now(),
		Provider:  provider,
		Detail:    detail,
	}
	return s.db.WithContext(ctx).Table(s.tableName).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "request_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"status", "updated_at", "provider", "detail"}),
		}).Create(&ns).Error
}
