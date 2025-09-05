package services

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/patrickmn/go-cache"
	"gorm.io/gorm"

	"github.com/ringecosystem/degov-apps/database"
	dbmodels "github.com/ringecosystem/degov-apps/database/models"
	gqlmodels "github.com/ringecosystem/degov-apps/graph/models"
	"github.com/ringecosystem/degov-apps/internal/config"
	"github.com/ringecosystem/degov-apps/internal/utils"
	"github.com/ringecosystem/degov-apps/types"
)

type UserInteractionService struct {
	db              *gorm.DB
	daoService      *DaoService
	templateService *TemplateService
	otpCache        *cache.Cache
	rateLimitCache  *cache.Cache
	notifierService *NotifierService
	userService     *UserService
}

func NewUserInteractionService() *UserInteractionService {
	otpCache := cache.New(3*time.Minute, 5*time.Minute)
	rateLimitCache := cache.New(1*time.Minute, 2*time.Minute)

	return &UserInteractionService{
		db:              database.GetDB(),
		daoService:      NewDaoService(),
		templateService: NewTemplateService(),
		otpCache:        otpCache,
		rateLimitCache:  rateLimitCache,
		notifierService: NewNotifierService(),
		userService:     NewUserService(),
	}
}

func (s *UserInteractionService) ModifyLikeDao(baseInput types.BasicInput[gqlmodels.ModifyLikeDaoInput]) (bool, error) {
	user := baseInput.User
	input := baseInput.Input

	_, err := s.daoService.Inspect(types.BasicInput[string]{
		User:  user,
		Input: input.DaoCode,
	})
	if err != nil {
		return false, fmt.Errorf("error inspecting DAO: %w", err)
	}

	var existingLikedDao dbmodels.UserLikedDao
	err = s.db.Where("user_id = ? AND dao_code = ?", user.Id, input.DaoCode).First(&existingLikedDao).Error
	isNotFoundError := false
	if err != nil {
		isNotFoundError = errors.Is(err, gorm.ErrRecordNotFound)
		if !isNotFoundError {
			return false, fmt.Errorf("error checking existing like: %w", err)
		}
	}

	switch baseInput.Input.Action {
	case gqlmodels.LikeActionLike:
		if isNotFoundError {
			// Create new like
			like := &dbmodels.UserLikedDao{
				ID:          utils.NextIDString(),
				DaoCode:     input.DaoCode,
				UserID:      user.Id,
				UserAddress: user.Address,
				CTime:       time.Now(),
			}

			if err := s.db.Create(like).Error; err != nil {
				return false, fmt.Errorf("error creating like: %w", err)
			}
			return true, nil
		}
	case gqlmodels.LikeActionUnlike:
		if !isNotFoundError {
			// Remove existing like
			result := s.db.Where("user_id = ? AND dao_code = ?", user.Id, input.DaoCode).Delete(&dbmodels.UserLikedDao{})
			if result.Error != nil {
				return false, fmt.Errorf("error removing like: %w", result.Error)
			}
			if result.RowsAffected == 0 {
				return false, errors.New("like not found")
			}
			return true, nil
		}
	}

	return true, nil
}

func (s *UserInteractionService) BindNotificationChannel(baseInput types.BasicInput[gqlmodels.BindNotificationChannelInput]) (*gqlmodels.ResendOTPOutput, error) {
	user := baseInput.User
	input := baseInput.Input

	var existingChannel dbmodels.NotificationChannel
	err := s.db.Where("user_id = ? AND channel_type = ?", user.Id, input.Type).First(&existingChannel).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("error checking existing channel: %w", err)
	}

	if err == nil {
		if existingChannel.Verified == 1 {
			return nil, errors.New("channel type already exists for this user")
		}

		return s.resendOTPForChannel(types.BasicInput[*dbmodels.NotificationChannel]{
			User:  user,
			Input: &existingChannel,
		})
	}

	channelID := utils.NextIDString()

	channel := &dbmodels.NotificationChannel{
		ID:           channelID,
		UserID:       user.Id,
		UserAddress:  user.Address,
		Verified:     0,
		ChannelType:  dbmodels.NotificationChannelType(input.Type),
		ChannelValue: input.Value,
		CTime:        time.Now(),
	}

	if err := s.db.Create(channel).Error; err != nil {
		return nil, fmt.Errorf("error creating notification channel: %w", err)
	}

	return s.resendOTPForChannel(types.BasicInput[*dbmodels.NotificationChannel]{
		User:  user,
		Input: channel,
	})
}

func (s *UserInteractionService) VerifyNotificationChannel(baseInput types.BasicInput[gqlmodels.VerifyNotificationChannelInput]) (*gqlmodels.VerifyNotificationChannelOutput, error) {
	user := baseInput.User
	input := baseInput.Input

	cachedOTP, found := s.otpCache.Get(input.ID)
	if !found {
		return &gqlmodels.VerifyNotificationChannelOutput{
			Code:    1,
			Message: utils.StringPtr("OTP code has expired or does not exist"),
		}, nil
	}

	cachedOTPStr, ok := cachedOTP.(string)
	if !ok {
		return &gqlmodels.VerifyNotificationChannelOutput{
			Code:    1,
			Message: utils.StringPtr("Invalid OTP code format"),
		}, nil
	}

	if cachedOTPStr != input.OtpCode {
		return &gqlmodels.VerifyNotificationChannelOutput{
			Code:    1,
			Message: utils.StringPtr("Invalid OTP code"),
		}, nil
	}

	s.otpCache.Delete(input.ID)

	var channel dbmodels.NotificationChannel
	err := s.db.Where("id = ? AND user_id = ?", input.ID, user.Id).First(&channel).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &gqlmodels.VerifyNotificationChannelOutput{
				Code:    1,
				Message: utils.StringPtr("Notification channel not found"),
			}, nil
		}
		return nil, fmt.Errorf("error finding notification channel: %w", err)
	}

	if err := s.db.Model(&channel).Update("verified", 1).Error; err != nil {
		return nil, fmt.Errorf("error updating channel verification status: %w", err)
	}

	return &gqlmodels.VerifyNotificationChannelOutput{
		Code: 0,
	}, nil
}

func (s *UserInteractionService) ResendOTP(baseInput types.BasicInput[gqlmodels.ResendOTPInput]) (*gqlmodels.ResendOTPOutput, error) {
	user := baseInput.User
	input := baseInput.Input

	var existingChannel dbmodels.NotificationChannel
	err := s.db.Where("user_id = ? AND channel_type = ? and channel_value = ?", user.Id, input.Type, input.Value).First(&existingChannel).Error
	if err != nil {
		return nil, fmt.Errorf("error checking existing channel: %w", err)
	}
	return s.resendOTPForChannel(types.BasicInput[*dbmodels.NotificationChannel]{
		User:  user,
		Input: &existingChannel,
	})
}

func (s *UserInteractionService) resendOTPForChannel(baseInput types.BasicInput[*dbmodels.NotificationChannel]) (*gqlmodels.ResendOTPOutput, error) {
	user := baseInput.User
	input := baseInput.Input

	if !config.GetAppEnv().IsDevelopment() {
		replacer := strings.NewReplacer(" ", "", "\n", "", "\r", "")
		stdValue := strings.ToLower(replacer.Replace(input.ChannelValue))
		rateLimitKey := fmt.Sprintf("otp_rate_limit_%s_%s_%s", user.Id, input.ChannelType, stdValue)

		if cachedTime, found := s.rateLimitCache.Get(rateLimitKey); found {
			if lastSentTime, ok := cachedTime.(time.Time); ok {
				elapsed := time.Since(lastSentTime)
				remaining := 60*time.Second - elapsed
				if remaining > 0 {
					remainingSeconds := int32(remaining.Seconds())
					return &gqlmodels.ResendOTPOutput{
						Code:      1,
						ID:        input.ID,
						RateLimit: &remainingSeconds,
						Message:   utils.StringPtr(fmt.Sprintf("OTP can only be sent once per minute. Please try again in %d seconds", remainingSeconds)),
					}, nil
				}
			} else {
				return &gqlmodels.ResendOTPOutput{
					Code:      1,
					ID:        input.ID,
					RateLimit: utils.Int32Ptr(60),
					Message:   utils.StringPtr("OTP can only be sent once per minute. Please try again later"),
				}, nil
			}
		}

		s.rateLimitCache.Set(rateLimitKey, time.Now(), 1*time.Minute)
	}

	ensName, err := s.userService.GetENSName(user.Address)
	if err != nil {
		slog.Warn("Failed to get ENS name", "address", user.Address, "err", err)
		// ensName = utils.StringPtr("")
	}

	switch input.ChannelType {
	case dbmodels.NotificationChannelTypeEmail:
		otpCode, err := utils.NextOTPCode()
		if err != nil {
			return nil, fmt.Errorf("error generating OTP code: %w", err)
		}
		s.otpCache.Set(input.ID, otpCode, 3*time.Minute)

		templateOutput, err := s.templateService.GenerateTemplateOTP(types.GenerateTemplateOTPInput{
			DegovSiteConfig: config.GetDegovSiteConfig(),
			OTP:             otpCode,
			Expiration:      3,
			UserAddress:     user.Address,
			EnsName:         ensName,
		})
		if err != nil {
			return nil, fmt.Errorf("error generating email content: %w", err)
		}
		if err := s.notifierService.Notify(types.NotifyInput{
			Type:     dbmodels.NotificationChannelTypeEmail,
			To:       input.ChannelValue,
			Template: templateOutput,
		}); err != nil {
			slog.Warn("Failed to notify", "err", err)
		}

		return &gqlmodels.ResendOTPOutput{
			Code:       0,
			ID:         input.ID,
			Expiration: utils.Int32Ptr(3 * 60),
		}, nil

	case dbmodels.NotificationChannelTypeWebhook:
		return &gqlmodels.ResendOTPOutput{
			Code:    0,
			ID:      input.ID,
			Message: utils.StringPtr("this method do not need send OTP to verify"),
		}, nil

	default:
		return &gqlmodels.ResendOTPOutput{
			Code:    0,
			ID:      input.ID,
			Message: utils.StringPtr("do not support this notification channel"),
		}, nil
	}

}

func (s UserInteractionService) ListChannel(
	baseInput types.BasicInput[types.ListChannelInput],
) ([]dbmodels.NotificationChannel, error) {
	user := baseInput.User
	var results []dbmodels.NotificationChannel

	query := s.db.Where("user_id = ?", user.Id)

	if baseInput.Input.Verified != nil {
		verified := 0
		if *baseInput.Input.Verified {
			verified = 1
		}
		query = query.Where("verified = ?", verified)
	}

	if err := query.Find(&results).Error; err != nil {
		return nil, err
	}
	return results, nil
}
