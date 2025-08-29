package services

import (
	"fmt"
	"time"

	"github.com/ringecosystem/degov-apps/database"
	dbmodels "github.com/ringecosystem/degov-apps/database/models"
	gqlmodels "github.com/ringecosystem/degov-apps/graph/models"
	"github.com/ringecosystem/degov-apps/internal/utils"
	"github.com/ringecosystem/degov-apps/types"
	"gorm.io/gorm"
)

type SubscribeService struct {
	db         *gorm.DB
	daoService *DaoService
}

func NewSubscribeService() *SubscribeService {
	return &SubscribeService{
		db:         database.GetDB(),
		daoService: NewDaoService(),
	}
}

type buildDaoFeatureInput struct {
	ChainID      int
	DaoCode      string
	UserID       string
	UserAddress  string
	FeatureInput *gqlmodels.FeatureSettingsDaoInput
}

type resetDaoFeaturesInput struct {
	DaoCode  string
	UserID   string
	Features []dbmodels.SubscribeFeature
}

type buildProposalFeatureInput struct {
	ChainID      int
	DaoCode      string
	ProposalID   string
	UserID       string
	UserAddress  string
	FeatureInput *gqlmodels.FeatureSettingsProposalInput
}

type resetProposalFeaturesInput struct {
	DaoCode    string
	UserID     string
	ProposalID string
	Features   []dbmodels.SubscribeFeature
}

type InspectSubscribeProposalInput struct {
	DaoCode    string
	ProposalID string
}

func (s *SubscribeService) buildDaoFeatures(
	input *buildDaoFeatureInput,
) []dbmodels.SubscribeFeature {
	featureInput := input.FeatureInput
	var features []dbmodels.SubscribeFeature

	if featureInput == nil {
		return features
	}

	if featureInput.EnableProposal != nil {
		features = append(features, dbmodels.SubscribeFeature{
			ID:       utils.NextIDString(),
			ChainID:  input.ChainID,
			DaoCode:  input.DaoCode,
			UserID:   input.UserID,
			Feature:  dbmodels.SubscribeFeatureEnableProposal,
			Strategy: utils.SafeBoolString(featureInput.EnableProposal),
		})
	}

	if featureInput.EnableVotingEndReminder != nil {
		features = append(features, dbmodels.SubscribeFeature{
			ID:       utils.NextIDString(),
			ChainID:  input.ChainID,
			DaoCode:  input.DaoCode,
			UserID:   input.UserID,
			Feature:  dbmodels.SubscribeFeatureEnableVotingEndReminder,
			Strategy: utils.SafeBoolString(featureInput.EnableVotingEndReminder),
		})
	}

	return features
}

// buildProposalFeatures builds SubscribeFeature records for a proposal based on the
// provided FeatureSettingsProposalInput. It defaults missing booleans to false and
// only creates feature rows for fields that were explicitly provided (non-nil).
func (s *SubscribeService) buildProposalFeatures(input *buildProposalFeatureInput) []dbmodels.SubscribeFeature {
	featureInput := input.FeatureInput
	var features []dbmodels.SubscribeFeature

	if featureInput == nil {
		return features
	}

	pid := input.ProposalID
	if featureInput.EnableVotingEndReminder != nil {
		features = append(features, dbmodels.SubscribeFeature{
			ID:          utils.NextIDString(),
			ChainID:     input.ChainID,
			DaoCode:     input.DaoCode,
			UserID:      input.UserID,
			UserAddress: input.UserAddress,
			Feature:     dbmodels.SubscribeFeatureEnableVotingEndReminder,
			Strategy:    utils.SafeBoolString(featureInput.EnableVotingEndReminder),
			ProposalID:  &pid,
		})
	}

	if featureInput.EnableVoted != nil {
		features = append(features, dbmodels.SubscribeFeature{
			ID:          utils.NextIDString(),
			ChainID:     input.ChainID,
			DaoCode:     input.DaoCode,
			UserID:      input.UserID,
			UserAddress: input.UserAddress,
			Feature:     dbmodels.SubscribeFeatureEnableVoted,
			Strategy:    utils.SafeBoolString(featureInput.EnableVoted),
			ProposalID:  &pid,
		})
	}

	if featureInput.EnableStateChanged != nil {
		features = append(features, dbmodels.SubscribeFeature{
			ID:          utils.NextIDString(),
			ChainID:     input.ChainID,
			DaoCode:     input.DaoCode,
			UserID:      input.UserID,
			UserAddress: input.UserAddress,
			Feature:     dbmodels.SubscribeFeatureEnableStateChanged,
			Strategy:    utils.SafeBoolString(featureInput.EnableStateChanged),
			ProposalID:  &pid,
		})
	}

	return features
}

func (s *SubscribeService) SubscribeDao(baseInput types.BasicInput[gqlmodels.SubscribeDaoInput]) (*gqlmodels.SubscribedDaoOutput, error) {
	user := baseInput.User
	sdInput := baseInput.Input
	featureInput := sdInput.Feature

	existingDao, err := s.daoService.Inspect(types.BasicInput[string]{
		User:  user,
		Input: sdInput.DaoCode,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to inspect existing DAO: %w", err)
	}

	existingSubscribedDao, err1 := s.InspectSubscribeDao(types.BasicInput[string]{
		User:  user,
		Input: sdInput.DaoCode,
	})
	chainId := int(existingDao.ChainID)

	if err1 == nil {
		existingSubscribedDao.UTime = time.Now()
		// You may want to save the updated time to DB here, e.g.:
		if err := s.db.Save(existingSubscribedDao).Error; err != nil {
			return nil, err
		}
	} else {
		newSubscribedDao := dbmodels.UserSubscribedDao{
			ID:          utils.NextIDString(),
			ChainID:     chainId,
			DaoCode:     sdInput.DaoCode,
			UserID:      user.Id,
			UserAddress: user.Address,
			State:       "SUBSCRIBED",
		}
		if err := s.db.Create(&newSubscribedDao).Error; err != nil {
			return nil, err
		}
	}

	features := s.buildDaoFeatures(&buildDaoFeatureInput{
		ChainID:      chainId,
		DaoCode:      sdInput.DaoCode,
		UserID:       user.Id,
		UserAddress:  user.Address,
		FeatureInput: featureInput,
	})

	if err := s.resetDaoFeatures(resetDaoFeaturesInput{
		UserID:   user.Id,
		DaoCode:  sdInput.DaoCode,
		Features: features,
	}); err != nil {
		return nil, err
	}

	output := &gqlmodels.SubscribedDaoOutput{
		DaoCode: sdInput.DaoCode,
	}
	return output, nil
}

func (s *SubscribeService) InspectSubscribeDao(baseInput types.BasicInput[string]) (*dbmodels.UserSubscribedDao, error) {
	user := baseInput.User
	daoCode := baseInput.Input

	var subscribedDao dbmodels.UserSubscribedDao
	err := s.db.
		Where(
			"(user_id = ? or user_address=?) AND dao_code = ?",
			user.Id,
			user.Address,
			daoCode,
		).
		First(&subscribedDao).Error
	if err != nil {
		return nil, err
	}
	return &subscribedDao, nil
}

func (s *SubscribeService) InspectSubscribeProposal(baseInput types.BasicInput[InspectSubscribeProposalInput]) (*dbmodels.UserSubscribedProposal, error) {
	user := baseInput.User
	input := baseInput.Input

	var subscribedProposal dbmodels.UserSubscribedProposal
	err := s.db.
		Where(
			"(user_id = ? or user_address= ?) AND dao_code = ? AND proposal_id = ?",
			user.Id,
			user.Address,
			input.DaoCode,
			input.ProposalID,
		).
		First(&subscribedProposal).Error
	if err != nil {
		return nil, err
	}
	return &subscribedProposal, nil
}

func (s *SubscribeService) SubscribeProposal(baseInput types.BasicInput[gqlmodels.SubscribeProposalInput]) (*gqlmodels.SubscribedProposalOutput, error) {
	user := baseInput.User
	spInput := baseInput.Input
	featureInput := spInput.Feature

	existingDao, err := s.daoService.Inspect(types.BasicInput[string]{
		User:  user,
		Input: spInput.DaoCode,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to inspect existing DAO: %w", err)
	}
	chainId := int(existingDao.ChainID)

	// check existing subscribed proposal
	existingSubscribedProposal, err1 := s.InspectSubscribeProposal(types.BasicInput[InspectSubscribeProposalInput]{
		User: user,
		Input: InspectSubscribeProposalInput{
			DaoCode:    spInput.DaoCode,
			ProposalID: spInput.ProposalID,
		},
	})
	if err1 == nil {
		existingSubscribedProposal.UTime = time.Now()
		if err := s.db.Save(existingSubscribedProposal).Error; err != nil {
			return nil, err
		}
	} else {
		newSubscribedProposal := dbmodels.UserSubscribedProposal{
			ID:          utils.NextIDString(),
			ChainID:     chainId,
			DaoCode:     spInput.DaoCode,
			UserID:      user.Id,
			UserAddress: user.Address,
			State:       dbmodels.SubscribeStateActive,
			ProposalID:  spInput.ProposalID,
		}
		if err := s.db.Create(&newSubscribedProposal).Error; err != nil {
			return nil, err
		}
	}

	features := s.buildProposalFeatures(&buildProposalFeatureInput{
		ChainID:      chainId,
		DaoCode:      spInput.DaoCode,
		ProposalID:   spInput.ProposalID,
		UserID:       user.Id,
		UserAddress:  user.Address,
		FeatureInput: featureInput,
	})

	if err := s.resetProposalFeatures(resetProposalFeaturesInput{
		UserID:     user.Id,
		DaoCode:    spInput.DaoCode,
		ProposalID: spInput.ProposalID,
		Features:   features,
	}); err != nil {
		return nil, err
	}

	output := &gqlmodels.SubscribedProposalOutput{
		DaoCode:    spInput.DaoCode,
		ProposalID: spInput.ProposalID,
	}
	return output, nil
}

func (s *SubscribeService) ListSubscribedUser(input types.ListSubscribeUserInput) ([]types.ListSubscribedUserOutput, error) {
	strategies := input.Strategies
	if len(strategies) == 0 {
		return nil, fmt.Errorf("no strategies provided for feature %s", input.Feature)
	}

	sql := "SELECT DISTINCT f.user_id, f.user_address, f.chain_id, f.dao_code, LEAST(d.ctime, p.ctime) AS ctime FROM dgv_subscribed_feature AS f " +
		"LEFT JOIN dgv_user_subscribed_dao AS d ON f.user_id = d.user_id AND f.dao_code = d.dao_code " +
		"LEFT JOIN dgv_user_subscribed_proposal AS p ON f.user_id = p.user_id AND f.proposal_id = p.proposal_id " +
		"WHERE f.feature = ? AND f.strategy IN ? AND f.dao_code = ? "

	queryParams := make([]interface{}, 0)
	queryParams = append(queryParams, input.Feature, strategies, input.DaoCode)

	// proposal id handling
	if input.ProposalId != nil {
		sql += "AND (f.proposal_id = ? OR f.proposal_id IS NULL) "
		queryParams = append(queryParams, *input.ProposalId)
	} else {
		sql += "AND f.proposal_id IS NULL "
	}

	if input.EventTime != nil {
		sql += "AND ((d.state = 'ACTIVE' AND d.ctime <= ?) OR (p.state = 'ACTIVE' AND p.ctime <= ?)) "
		queryParams = append(queryParams, *input.EventTime, *input.EventTime)
	} else {
		sql += "AND (d.state = 'ACTIVE' OR p.state = 'ACTIVE') "
	}

	// ordering and pagination
	if input.Limit <= 0 {
		input.Limit = 100
	}
	sql += "ORDER BY f.ctime ASC, f.user_id ASC LIMIT ? OFFSET ?"
	queryParams = append(queryParams, input.Limit, input.Offset)

	var outputs []types.ListSubscribedUserOutput
	if err := s.db.Raw(sql, queryParams...).Scan(&outputs).Error; err != nil {
		return nil, err
	}

	return outputs, nil
}

func (s *SubscribeService) resetDaoFeatures(input resetDaoFeaturesInput) error {
	if err := s.db.Where(
		"dao_code = ? and (user_id =? or user_address = ?)",
		input.DaoCode,
		input.UserID,
		input.UserID,
	).
		Delete(&dbmodels.SubscribeFeature{}).Error; err != nil {
		return err
	}

	if len(input.Features) > 0 {
		return s.db.Create(input.Features).Error
	}
	return nil
}

// resetProposalFeatures deletes existing features scoped to the dao + proposal + user
// and inserts the provided features (if any).
func (s *SubscribeService) resetProposalFeatures(input resetProposalFeaturesInput) error {
	if err := s.db.Where(
		"dao_code = ? and proposal_id = ? and (user_id = ? or user_address = ?)",
		input.DaoCode,
		input.ProposalID,
		input.UserID,
		input.UserID,
	).
		Delete(&dbmodels.SubscribeFeature{}).Error; err != nil {
		return err
	}

	if len(input.Features) > 0 {
		return s.db.Create(input.Features).Error
	}
	return nil
}
