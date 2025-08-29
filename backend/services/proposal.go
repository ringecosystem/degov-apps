package services

import (
	"errors"
	"log/slog"
	"time"

	"gorm.io/gorm"

	"github.com/ringecosystem/degov-apps/database"
	dbmodels "github.com/ringecosystem/degov-apps/database/models"
	"github.com/ringecosystem/degov-apps/internal/utils"
	"github.com/ringecosystem/degov-apps/types"
)

type ProposalService struct {
	db                  *gorm.DB
	notificationService *NotificationService
}

func NewProposalService() *ProposalService {
	return &ProposalService{
		db:                  database.GetDB(),
		notificationService: NewNotificationService(),
	}
}

// StoreProposalTracking checks if proposal exists and creates it if not
// Returns true if proposal was newly created, false if it already existed
func (s *ProposalService) StoreProposalTracking(input types.ProposalTrackingInput) (bool, error) {
	// Check if proposal already exists
	var existingProposal dbmodels.ProposalTracking
	err := s.db.
		Where("proposal_id = ? AND dao_code = ?", input.ProposalID, input.DaoCode).
		First(&existingProposal).
		Error

	if err == nil {
		// Proposal already exists
		return false, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		// Database error
		return false, err
	}

	// Proposal doesn't exist, create it
	newProposal := &dbmodels.ProposalTracking{
		ID:                utils.NextIDString(),
		DaoCode:           input.DaoCode,
		ChainId:           input.ChainId,
		ProposalLink:      input.ProposalLink,
		ProposalId:        input.ProposalID,
		State:             dbmodels.ProposalStateUnknown, // Default state
		ProposalCreatedAt: input.ProposalCreatedAt,
		ProposalAtBlock:   input.ProposalAtBlock,
		CTime:             time.Now(),
	}

	if err := s.db.Create(newProposal).Error; err != nil {
		return false, err
	}

	if err := s.notificationService.SaveEvent(dbmodels.NotificationEvent{
		ChainID:    newProposal.ChainId,
		DaoCode:    newProposal.DaoCode,
		Type:       dbmodels.NotificationTypeNewProposal,
		ProposalID: newProposal.ProposalId,
		TimeEvent:  newProposal.CTime,
	}); err != nil {
		slog.Warn("failed to save notification event for new proposal", "error", err, "proposal_id", newProposal.ProposalId, "dao_code", newProposal.DaoCode)
	}

	return true, nil
}

func (s *ProposalService) TrackingStateProposals(input types.TrackingStateProposalsInput) ([]*dbmodels.ProposalTracking, error) {
	var proposals []*dbmodels.ProposalTracking

	timesTrack := 10
	if input.TimesTrack != nil {
		timesTrack = *input.TimesTrack
	}

	// Query proposals with specific states, tracking limits, and time conditions
	err := s.db.Where(`dao_code = ?
		AND state IN ?
		AND times_track < ?
		AND (time_next_track IS NULL OR time_next_track <= ?)`,
		input.DaoCode,
		input.States,
		timesTrack,
		time.Now(),
	).
		Order("proposal_created_at asc").
		Find(&proposals).Error

	if err != nil {
		return nil, err
	}
	return proposals, nil
}

// UpdateProposalState updates the state of a proposal
func (s *ProposalService) UpdateProposalState(proposalID, daoCode string, newState dbmodels.ProposalState) error {
	return s.db.Model(&dbmodels.ProposalTracking{}).
		Where("proposal_id = ? AND dao_code = ?", proposalID, daoCode).
		Updates(map[string]interface{}{
			"state": newState,
			"utime": time.Now(),
		}).Error
}

// UpdateProposalTrackingError updates tracking info when processing fails
func (s *ProposalService) UpdateProposalTrackingError(proposalID, daoCode string, errorMessage string) error {
	// Get current proposal to build new message
	var proposal dbmodels.ProposalTracking
	err := s.db.Where("proposal_id = ? AND dao_code = ?", proposalID, daoCode).First(&proposal).Error
	if err != nil {
		return err
	}

	// Calculate next tracking time (exponential backoff based on attempts)
	nextTrackTime := time.Now()
	attempts := proposal.TimesTrack + 1
	if attempts <= 3 {
		nextTrackTime = nextTrackTime.Add(time.Hour) // 1 hour for first 3 attempts
	} else if attempts <= 6 {
		nextTrackTime = nextTrackTime.Add(2 * time.Hour) // 2 hours for attempts 4-6
	} else {
		nextTrackTime = nextTrackTime.Add(5 * time.Hour) // 5 hours for more attempts
	}

	// Build new message in format: [${time_next_track}] ${message}\n----\nOLD_MESSAGE
	newMessage := "[" + nextTrackTime.Format("2006-01-02 15:04:05") + "] " + errorMessage
	if proposal.Message != "" {
		newMessage += "\n----\n" + proposal.Message
	}

	return s.db.Model(&dbmodels.ProposalTracking{}).
		Where("proposal_id = ? AND dao_code = ?", proposalID, daoCode).
		Updates(map[string]interface{}{
			"times_track":     gorm.Expr("times_track + 1"),
			"time_next_track": nextTrackTime,
			"message":         newMessage,
			"utime":           time.Now(),
		}).Error
}

func (s *ProposalService) UpdateOffsetTrackingVote(proposalID, daoCode string, offset int) error {
	return s.db.Model(&dbmodels.ProposalTracking{}).
		Where("proposal_id = ? AND dao_code = ?", proposalID, daoCode).
		Updates(map[string]interface{}{
			"offset_tracking_vote": offset,
			"utime":                time.Now(),
		}).Error
}

// ProposalStateCount returns count of proposals by DAO and state for active DAOs
func (s *ProposalService) ProposalStateCount() ([]types.ProposalStateCountResult, error) {
	var results []types.ProposalStateCountResult

	err := s.db.Table("dgv_proposal_tracking as t").
		Select("t.dao_code, t.state, count(1) as total").
		Joins("left join dgv_dao as d on t.dao_code = d.code").
		Where("d.state = ?", "ACTIVE").
		Group("t.dao_code, t.state").
		Order("t.dao_code, t.state").
		Find(&results).Error

	if err != nil {
		return nil, err
	}

	return results, nil
}
