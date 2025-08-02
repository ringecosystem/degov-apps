package services

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/ringecosystem/degov-apps/internal/database"
	"github.com/ringecosystem/degov-apps/models"
)

type UserInteractionService struct {
	db *gorm.DB
}

func NewUserInteractionService() *UserInteractionService {
	return &UserInteractionService{
		db: database.GetDB(),
	}
}

// UserLikedDao methods
func (s *UserInteractionService) LikeDao(userID, daoCode string) (*models.UserLikedDao, error) {
	// check if already liked
	var existing models.UserLikedDao
	err := s.db.Where("user_id = ? AND dao_code = ?", userID, daoCode).First(&existing).Error
	if err == nil {
		return nil, errors.New("DAO already liked by user")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("error checking existing like: %w", err)
	}

	// generate like ID
	likeID := fmt.Sprintf("like_%d", s.generateLikeID())

	like := &models.UserLikedDao{
		ID:      likeID,
		DaoCode: daoCode,
		UserID:  userID,
		CTime:   time.Now(),
	}

	if err := s.db.Create(like).Error; err != nil {
		return nil, fmt.Errorf("error creating like: %w", err)
	}

	return like, nil
}

func (s *UserInteractionService) UnlikeDao(userID, daoCode string) error {
	result := s.db.Where("user_id = ? AND dao_code = ?", userID, daoCode).Delete(&models.UserLikedDao{})
	if result.Error != nil {
		return fmt.Errorf("error removing like: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.New("like not found")
	}
	return nil
}

func (s *UserInteractionService) GetUserLikedDaos(userID string) ([]*models.UserLikedDao, error) {
	var likes []*models.UserLikedDao
	err := s.db.Where("user_id = ?", userID).Find(&likes).Error
	if err != nil {
		return nil, fmt.Errorf("error getting user liked DAOs: %w", err)
	}
	return likes, nil
}

// UserFollowedDao methods
func (s *UserInteractionService) FollowDao(userID, daoCode string, chainID int, enableNewProposal, enableVotingEndReminder int) (*models.UserFollowedDao, error) {
	// check if already following
	var existing models.UserFollowedDao
	err := s.db.Where("user_id = ? AND dao_code = ?", userID, daoCode).First(&existing).Error
	if err == nil {
		return nil, errors.New("DAO already followed by user")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("error checking existing follow: %w", err)
	}

	// generate follow ID
	followID := fmt.Sprintf("follow_%d", s.generateFollowID())

	follow := &models.UserFollowedDao{
		ID:                      followID,
		ChainID:                 chainID,
		DaoCode:                 daoCode,
		UserID:                  userID,
		EnableNewProposal:       enableNewProposal,
		EnableVotingEndReminder: enableVotingEndReminder,
		CTime:                   time.Now(),
	}

	if err := s.db.Create(follow).Error; err != nil {
		return nil, fmt.Errorf("error creating follow: %w", err)
	}

	return follow, nil
}

func (s *UserInteractionService) UnfollowDao(userID, daoCode string) error {
	result := s.db.Where("user_id = ? AND dao_code = ?", userID, daoCode).Delete(&models.UserFollowedDao{})
	if result.Error != nil {
		return fmt.Errorf("error removing follow: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.New("follow not found")
	}
	return nil
}

func (s *UserInteractionService) GetUserFollowedDaos(userID string) ([]*models.UserFollowedDao, error) {
	var follows []*models.UserFollowedDao
	err := s.db.Where("user_id = ?", userID).Find(&follows).Error
	if err != nil {
		return nil, fmt.Errorf("error getting user followed DAOs: %w", err)
	}
	return follows, nil
}

func (s *UserInteractionService) UpdateFollowSettings(userID, daoCode string, enableNewProposal, enableVotingEndReminder int) (*models.UserFollowedDao, error) {
	var follow models.UserFollowedDao
	err := s.db.Where("user_id = ? AND dao_code = ?", userID, daoCode).First(&follow).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("follow not found")
		}
		return nil, fmt.Errorf("error finding follow: %w", err)
	}

	follow.EnableNewProposal = enableNewProposal
	follow.EnableVotingEndReminder = enableVotingEndReminder

	if err := s.db.Save(&follow).Error; err != nil {
		return nil, fmt.Errorf("error updating follow settings: %w", err)
	}

	return &follow, nil
}

func (s *UserInteractionService) generateLikeID() int64 {
	var count int64
	s.db.Model(&models.UserLikedDao{}).Count(&count)
	return count + 1
}

func (s *UserInteractionService) generateFollowID() int64 {
	var count int64
	s.db.Model(&models.UserFollowedDao{}).Count(&count)
	return count + 1
}
