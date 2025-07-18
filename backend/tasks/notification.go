package tasks

import (
	"log/slog"
	"time"

	"gorm.io/gorm"

	"github.com/ringecosystem/degov-apps/internal/database"
)

type NotificationTask struct {
	db *gorm.DB
}

// NewNotificationTask creates a new notification task
func NewNotificationTask() *NotificationTask {
	return &NotificationTask{
		db: database.GetDB(),
	}
}

// Name returns the task name
func (t *NotificationTask) Name() string {
	return "notification-cleanup"
}

// Execute performs notification cleanup and processing
func (t *NotificationTask) Execute() error {
	startTime := time.Now()
	slog.Info("Starting notification cleanup task", "timestamp", startTime.Format(time.RFC3339))

	// Example: Clean up old notification records (older than 30 days)
	if err := t.cleanupOldNotifications(); err != nil {
		return err
	}

	// Example: Process pending notifications
	if err := t.processPendingNotifications(); err != nil {
		return err
	}

	duration := time.Since(startTime)
	slog.Info("Notification cleanup completed",
		"duration", duration.String(),
		"timestamp", time.Now().Format(time.RFC3339))

	return nil
}

// cleanupOldNotifications removes notification records older than 30 days
func (t *NotificationTask) cleanupOldNotifications() error {
	cutoffTime := time.Now().AddDate(0, 0, -30) // 30 days ago

	result := t.db.Exec("DELETE FROM dgv_notification_record WHERE ctime < ?", cutoffTime)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected > 0 {
		slog.Info("Cleaned up old notifications", "deleted_count", result.RowsAffected)
	}

	return nil
}

// processPendingNotifications handles failed notifications for retry
func (t *NotificationTask) processPendingNotifications() error {
	// Example: Find notifications that failed and need retry
	var failedNotifications []struct {
		ID         string `gorm:"column:id"`
		RetryTimes int    `gorm:"column:retry_times"`
	}

	err := t.db.Table("dgv_notification_record").
		Select("id, retry_times").
		Where("status = ? AND retry_times < ?", "SENT_FAIL", 3).
		Find(&failedNotifications).Error

	if err != nil {
		return err
	}

	if len(failedNotifications) > 0 {
		slog.Info("Found notifications to retry", "count", len(failedNotifications))

		// Here you would implement the actual retry logic
		// For now, just log that we found them
		for _, notification := range failedNotifications {
			slog.Debug("Notification ready for retry",
				"id", notification.ID,
				"retry_count", notification.RetryTimes)
		}
	}

	return nil
}
