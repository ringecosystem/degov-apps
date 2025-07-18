package services

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/ringecosystem/degov-apps/internal/database"
	"github.com/ringecosystem/degov-apps/models"
)

type NotificationService struct {
	db *gorm.DB
}

func NewNotificationService() *NotificationService {
	return &NotificationService{
		db: database.GetDB(),
	}
}

// UserChannel methods
func (s *NotificationService) CreateUserChannel(userID, channelType, channelValue string, payload *string) (*models.UserChannel, error) {
	// check if channel already exists for user
	var existing models.UserChannel
	err := s.db.Where("user_id = ? AND channel_type = ? AND channel_value = ?", userID, channelType, channelValue).First(&existing).Error
	if err == nil {
		return nil, errors.New("channel already exists for user")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("error checking existing channel: %w", err)
	}

	// generate channel ID
	channelID := fmt.Sprintf("channel_%d", s.generateChannelID())

	channel := &models.UserChannel{
		ID:           channelID,
		UserID:       userID,
		Verified:     0, // not verified by default
		ChannelType:  channelType,
		ChannelValue: channelValue,
		Payload:      payload,
		CTime:        time.Now(),
	}

	if err := s.db.Create(channel).Error; err != nil {
		return nil, fmt.Errorf("error creating channel: %w", err)
	}

	return channel, nil
}

func (s *NotificationService) UpdateUserChannel(id, userID, channelType, channelValue string, payload *string) (*models.UserChannel, error) {
	var channel models.UserChannel
	err := s.db.First(&channel, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("channel not found")
		}
		return nil, fmt.Errorf("error finding channel: %w", err)
	}

	channel.UserID = userID
	channel.ChannelType = channelType
	channel.ChannelValue = channelValue
	channel.Payload = payload

	if err := s.db.Save(&channel).Error; err != nil {
		return nil, fmt.Errorf("error updating channel: %w", err)
	}

	return &channel, nil
}

func (s *NotificationService) DeleteUserChannel(id string) error {
	result := s.db.Delete(&models.UserChannel{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("error deleting channel: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.New("channel not found")
	}
	return nil
}

func (s *NotificationService) VerifyUserChannel(id string) (*models.UserChannel, error) {
	var channel models.UserChannel
	err := s.db.First(&channel, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("channel not found")
		}
		return nil, fmt.Errorf("error finding channel: %w", err)
	}

	channel.Verified = 1

	if err := s.db.Save(&channel).Error; err != nil {
		return nil, fmt.Errorf("error verifying channel: %w", err)
	}

	return &channel, nil
}

func (s *NotificationService) GetUserChannels(userID string) ([]*models.UserChannel, error) {
	var channels []*models.UserChannel
	err := s.db.Where("user_id = ?", userID).Find(&channels).Error
	if err != nil {
		return nil, fmt.Errorf("error getting user channels: %w", err)
	}
	return channels, nil
}

// NotificationRecord methods
func (s *NotificationService) CreateNotificationRecord(chainID int, chainName, daoName, daoCode, notificationType, targetID, userID, status string, message *string) (*models.NotificationRecord, error) {
	// generate notification ID
	notificationID := fmt.Sprintf("notification_%d", s.generateNotificationID())

	record := &models.NotificationRecord{
		ID:         notificationID,
		ChainID:    chainID,
		ChainName:  chainName,
		DaoName:    daoName,
		DaoCode:    daoCode,
		Type:       notificationType,
		TargetID:   &targetID,
		UserID:     userID,
		Status:     status,
		Message:    message,
		RetryTimes: 0,
		CTime:      time.Now(),
	}

	if err := s.db.Create(record).Error; err != nil {
		return nil, fmt.Errorf("error creating notification record: %w", err)
	}

	return record, nil
}

func (s *NotificationService) GetNotificationRecords(userID string) ([]*models.NotificationRecord, error) {
	var records []*models.NotificationRecord
	err := s.db.Where("user_id = ?", userID).Order("ctime DESC").Find(&records).Error
	if err != nil {
		return nil, fmt.Errorf("error getting notification records: %w", err)
	}
	return records, nil
}

func (s *NotificationService) GetNotificationRecordsByStatus(status string) ([]*models.NotificationRecord, error) {
	var records []*models.NotificationRecord
	err := s.db.Where("status = ?", status).Order("ctime ASC").Find(&records).Error
	if err != nil {
		return nil, fmt.Errorf("error getting notification records by status: %w", err)
	}
	return records, nil
}

func (s *NotificationService) UpdateNotificationStatus(id, status string, message *string) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if message != nil {
		updates["message"] = *message
	}

	result := s.db.Model(&models.NotificationRecord{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("error updating notification status: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.New("notification record not found")
	}
	return nil
}

func (s *NotificationService) IncrementRetryCount(id string) error {
	result := s.db.Model(&models.NotificationRecord{}).Where("id = ?", id).Update("retry_times", gorm.Expr("retry_times + 1"))
	if result.Error != nil {
		return fmt.Errorf("error incrementing retry count: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.New("notification record not found")
	}
	return nil
}

func (s *NotificationService) generateChannelID() int64 {
	var count int64
	s.db.Model(&models.UserChannel{}).Count(&count)
	return count + 1
}

func (s *NotificationService) generateNotificationID() int64 {
	var count int64
	s.db.Model(&models.NotificationRecord{}).Count(&count)
	return count + 1
}
