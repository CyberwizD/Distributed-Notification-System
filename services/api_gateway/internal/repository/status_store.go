package repository

import (
	"gorm.io/gorm"
)

// NotificationStatus represents the status of a notification.
type NotificationStatus struct {
	RequestID string `gorm:"primaryKey"`
	Status    string
}

// StatusStore stores and retrieves notification statuses.
type StatusStore struct {
	db *gorm.DB
}

// NewStatusStore creates a new StatusStore.
func NewStatusStore(db *gorm.DB) *StatusStore {
	// Auto-migrate the schema
	db.AutoMigrate(&NotificationStatus{})
	return &StatusStore{db: db}
}

// SetStatus sets the status for a given request ID.
func (s *StatusStore) SetStatus(requestID, status string) error {
	return s.db.Create(&NotificationStatus{RequestID: requestID, Status: status}).Error
}

// GetStatus retrieves the status for a given request ID.
func (s *StatusStore) GetStatus(requestID string) (string, error) {
	var notificationStatus NotificationStatus
	if err := s.db.First(&notificationStatus, "request_id = ?", requestID).Error; err != nil {
		return "", err
	}
	return notificationStatus.Status, nil
}
