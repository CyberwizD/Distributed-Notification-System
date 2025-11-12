package repository

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// NotificationStatus represents the status of a notification.
type NotificationStatus struct {
	RequestID string    `gorm:"primaryKey" json:"request_id"`
	Status    string    `json:"status"`
	Provider  string    `json:"provider"`
	Detail    string    `json:"detail"`
	UpdatedAt time.Time `json:"updated_at"`
}

// StatusStore stores and retrieves notification statuses.
type StatusStore struct {
	db *gorm.DB
}

// NewStatusStore creates a new StatusStore.
func NewStatusStore(db *gorm.DB) *StatusStore {
	// Auto-migrate the schema
	_ = db.AutoMigrate(&NotificationStatus{})
	return &StatusStore{db: db}
}

// SetStatus upserts the status for a given request ID.
func (s *StatusStore) SetStatus(requestID, status string) error {
	return s.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "request_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"status", "provider", "detail", "updated_at"}),
	}).Create(&NotificationStatus{
		RequestID: requestID,
		Status:    status,
		UpdatedAt: time.Now(),
	}).Error
}

// SetStatusWithProvider allows downstream services to include provider info.
func (s *StatusStore) SetStatusWithProvider(requestID, status, provider, detail string) error {
	return s.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "request_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"status", "provider", "detail", "updated_at"}),
	}).Create(&NotificationStatus{
		RequestID: requestID,
		Status:    status,
		Provider:  provider,
		Detail:    detail,
		UpdatedAt: time.Now(),
	}).Error
}

// GetStatus retrieves the status for a given request ID.
func (s *StatusStore) GetStatus(requestID string) (string, error) {
	var notificationStatus NotificationStatus
	if err := s.db.First(&notificationStatus, "request_id = ?", requestID).Error; err != nil {
		return "", err
	}
	return notificationStatus.Status, nil
}
