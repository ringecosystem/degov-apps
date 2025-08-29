package services

import (
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
	"gorm.io/gorm"

	"github.com/jinzhu/copier"
	"github.com/ringecosystem/degov-apps/database"
	dbmodels "github.com/ringecosystem/degov-apps/database/models"
	gqlmodels "github.com/ringecosystem/degov-apps/graph/models"
	"github.com/ringecosystem/degov-apps/internal/utils"
	"github.com/ringecosystem/degov-apps/types"
)

type DaoService struct {
	db *gorm.DB
}

func NewDaoService() *DaoService {
	return &DaoService{
		db: database.GetDB(),
	}
}

func (s *DaoService) convertToGqlDao(dbDao dbmodels.Dao) *gqlmodels.Dao {
	var tags []string
	if dbDao.Tags != "" {
		if err := json.Unmarshal([]byte(dbDao.Tags), &tags); err != nil {
			tags = []string{} // If JSON unmarshal fails, treat as empty array
		}
	}
	gqlDao := gqlmodels.Dao{}
	copier.Copy(&gqlDao, &dbDao)
	gqlDao.Tags = tags
	return &gqlDao
}

func (s *DaoService) Inspect(baseInput types.BasicInput[string]) (*gqlmodels.Dao, error) {
	code := baseInput.Input
	var dbDao dbmodels.Dao
	err := s.db.Where("code = ?", code).First(&dbDao).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("dao not found")
		}
		return nil, err
	}
	dao := s.convertToGqlDao(dbDao)

	// chips
	chips, err := s.SingleDaoChips(code)
	if err != nil {
		return nil, err
	}
	dao.Chips = chips

	return dao, nil
}

func (s *DaoService) SingleDaoChips(code string) ([]*gqlmodels.DaoChip, error) {
	return s.MultipleDaoChips([]string{code})
}

func (s *DaoService) MultipleDaoChips(codes []string) ([]*gqlmodels.DaoChip, error) {
	var dbChips []dbmodels.DgvDaoChip
	err := s.db.Where("dao_code IN ?", codes).Find(&dbChips).Error
	if err != nil {
		return nil, err
	}

	var chips []*gqlmodels.DaoChip
	for _, dbChip := range dbChips {
		chip := gqlmodels.DaoChip{}
		copier.Copy(&chip, &dbChip)
		chips = append(chips, &chip)
	}

	return chips, nil
}

func (s *DaoService) ListDaos(baseInput types.BasicInput[*types.ListDaosInput]) ([]*gqlmodels.Dao, error) {
	input := baseInput.Input
	var whereClauses []string
	var params []interface{}

	if input != nil {
		if input.Codes != nil && len(*input.Codes) > 0 {
			whereClauses = append(whereClauses, "d.code IN ?")
			params = append(params, *input.Codes)
		}
		if input.State != nil && len(*input.State) > 0 {
			whereClauses = append(whereClauses, "d.state IN ?")
			params = append(params, *input.State)
		}
	}

	if input == nil || input.State == nil || len(*input.State) == 0 {
		whereClauses = append(whereClauses, "d.state = ?")
		params = append(params, dbmodels.DaoStateActive)
	}

	whereSQL := ""
	if len(whereClauses) > 0 {
		whereSQL = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	var allParams []interface{}
	user := baseInput.User
	var selectBuilder, joinBuilder, orderBuilder strings.Builder

	// Base SELECT columns, with corrected aliases
	selectBuilder.WriteString(`
		d.*,
		lp.id as lp_id,
		lp.dao_code as lp_dao_code,
		lp.chain_id as lp_chain_id,
		lp.proposal_link as lp_proposal_link,
		lp.proposal_id as lp_proposal_id,
		lp.state as lp_state,
		lp.proposal_at_block as lp_proposal_at_block,
		lp.proposal_created_at as lp_proposal_created_at,
		lp.times_track as lp_times_track,
		lp.time_next_track as lp_time_next_track,
		lp.message as lp_message,
		lp.ctime as lp_ctime,
		lp.utime as lp_utime
	`)

	if user == nil {
		allParams = params
		orderBuilder.WriteString("ORDER BY lp.proposal_created_at DESC NULLS LAST")
	} else {
		allParams = append([]interface{}{user.Id}, params...)
		selectBuilder.WriteString(", CASE WHEN uld.dao_code IS NOT NULL THEN 1 ELSE 0 END AS liked")
		joinBuilder.WriteString("LEFT JOIN dgv_user_liked_dao uld ON d.code = uld.dao_code AND uld.user_id = ?")
		orderBuilder.WriteString("ORDER BY liked DESC, lp.proposal_created_at DESC NULLS LAST")
	}

	sql := `
		WITH LatestProposals AS (
			SELECT * FROM (
				SELECT *, ROW_NUMBER() OVER(PARTITION BY dao_code ORDER BY proposal_created_at DESC) as rn
				FROM dgv_proposal_tracking
			) RankedProposals
			WHERE rn = 1
		)
		SELECT ` + selectBuilder.String() + `
		FROM dgv_dao d
		LEFT JOIN LatestProposals lp ON d.code = lp.dao_code ` +
		joinBuilder.String() + ` ` +
		whereSQL + ` ` +
		orderBuilder.String()

	// --- daoRow struct modified to use native pointer types ---
	type daoRow struct {
		dbmodels.Dao
		Liked               *int    // Pointer to int
		LpID                *string // Pointer to string
		LpDaoCode           *string
		LpChainID           *int64
		LpProposalLink      *string
		LpProposalID        *string
		LpState             *string
		LpProposalAtBlock   *int64
		LpProposalCreatedAt *time.Time // Already a pointer, no change
		LpTimesTrack        *int64
		LpTimeNextTrack     *time.Time // Already a pointer, no change
		LpMessage           *string
		LpCTime             *time.Time
		LpUTime             *time.Time // Already a pointer, no change
	}

	var rows []daoRow
	if err := s.db.Raw(sql, allParams...).Scan(&rows).Error; err != nil {
		return nil, err
	}

	// Extract dao codes from dbDaos
	var daoCodes []string
	for _, row := range rows {
		daoCodes = append(daoCodes, row.Code)
	}

	// Batch query chips for all DAOs
	daosChips, err := s.MultipleDaoChips(daoCodes)
	if err != nil {
		return nil, err
	}

	daos := make([]*gqlmodels.Dao, len(rows))
	for i, row := range rows {
		dao := s.convertToGqlDao(row.Dao)

		// Check pointer for nil and then check its value
		liked := row.Liked != nil && *row.Liked == 1
		dao.Liked = &liked

		// Find chips for this DAO
		var chips []*gqlmodels.DaoChip
		for _, daoChip := range daosChips {
			if daoChip.DaoCode == dao.Code {
				chips = append(chips, daoChip)
			}
		}
		dao.Chips = chips

		// --- Mapping logic adjusted for pointers ---
		// Check if the pointer is not nil to determine if a proposal exists
		if row.LpID != nil {
			proposal := &gqlmodels.Proposal{
				ID: *row.LpID, // Dereference safely after nil check
				// Directly assign pointers where the target type is also a pointer
				TimeNextTrack: row.LpTimeNextTrack,
				Utime:         row.LpUTime,
			}

			// Safely handle other nullable fields by checking for nil before dereferencing
			if row.LpDaoCode != nil {
				proposal.DaoCode = *row.LpDaoCode
			}
			if row.LpChainID != nil {
				proposal.ChainID = int32(*row.LpChainID)
			}
			if row.LpProposalLink != nil {
				proposal.ProposalLink = *row.LpProposalLink
			}
			if row.LpProposalID != nil {
				proposal.ProposalID = *row.LpProposalID
			}
			if row.LpState != nil {
				proposal.State = gqlmodels.ProposalState(*row.LpState)
			}
			if row.LpProposalAtBlock != nil {
				proposal.ProposalAtBlock = int32(*row.LpProposalAtBlock)
			}
			if row.LpProposalCreatedAt != nil {
				proposal.ProposalCreatedAt = *row.LpProposalCreatedAt
			}
			if row.LpTimesTrack != nil {
				proposal.TimesTrack = int32(*row.LpTimesTrack)
			}
			if row.LpCTime != nil {
				proposal.Ctime = *row.LpCTime
			}
			if row.LpMessage != nil {
				proposal.Message = row.LpMessage // Assign pointer directly
			}

			dao.LastProposal = proposal
		}
		daos[i] = dao
	}
	return daos, nil
}

func (s *DaoService) RefreshDaoAndConfig(input types.RefreshDaoAndConfigInput) error {
	var existingDao dbmodels.Dao
	result := s.db.Where("code = ?", input.Code).First(&existingDao)

	tagsJson := utils.ToJSON(input.Tags)
	if result.Error == gorm.ErrRecordNotFound {
		// Insert new DAO
		dao := &dbmodels.Dao{
			ID:                  utils.NextIDString(),
			ChainID:             input.Config.Chain.ID,
			ChainName:           input.Config.Chain.Name,
			ChainLogo:           input.Config.Chain.Logo,
			Name:                input.Config.Name,
			Code:                input.Code,
			Logo:                input.Config.Logo,
			Endpoint:            input.Config.SiteURL,
			State:               input.State,
			Tags:                tagsJson,
			ConfigLink:          input.ConfigLink,
			TimeSyncd:           utils.TimePtrNow(),
			OffsetTrackingBlock: 0, // Default to 0 for new DAOs
		}

		// Set metrics fields if they are provided (not nil)
		if input.MetricsCountProposals != nil {
			dao.MetricsCountProposals = *input.MetricsCountProposals
		}
		if input.MetricsCountMembers != nil {
			dao.MetricsCountMembers = *input.MetricsCountMembers
		}
		if input.MetricsSumPower != nil {
			dao.MetricsSumPower = *input.MetricsSumPower
		}
		if input.MetricsCountVote != nil {
			dao.MetricsCountVote = *input.MetricsCountVote
		}
		if err := s.db.Create(dao).Error; err != nil {
			return err
		}
	} else {
		// Update existing DAO
		existingDao.ChainID = input.Config.Chain.ID
		existingDao.ChainName = input.Config.Chain.Name
		existingDao.ChainLogo = input.Config.Chain.Logo
		existingDao.Name = input.Config.Name
		existingDao.Logo = input.Config.Logo
		existingDao.Endpoint = input.Config.SiteURL
		existingDao.State = input.State
		existingDao.Tags = tagsJson
		existingDao.ConfigLink = input.ConfigLink
		existingDao.UTime = utils.TimePtrNow()
		existingDao.TimeSyncd = utils.TimePtrNow()

		// Only update metrics if they are provided (not nil)
		if input.MetricsCountProposals != nil {
			existingDao.MetricsCountProposals = *input.MetricsCountProposals
		}
		if input.MetricsCountMembers != nil {
			existingDao.MetricsCountMembers = *input.MetricsCountMembers
		}
		if input.MetricsSumPower != nil {
			existingDao.MetricsSumPower = *input.MetricsSumPower
		}
		if input.MetricsCountVote != nil {
			existingDao.MetricsCountVote = *input.MetricsCountVote
		}
		if err := s.db.Save(&existingDao).Error; err != nil {
			return err
		}
	}

	var existingConfig dbmodels.DgvDaoConfig
	r2 := s.db.Where("dao_code = ?", input.Code).First(&existingConfig)

	if r2.Error == gorm.ErrRecordNotFound {
		// Insert new DAO config
		config := &dbmodels.DgvDaoConfig{
			ID:      utils.NextIDString(),
			DaoCode: input.Code,
			Config:  input.Raw,
			CTime:   time.Now(),
		}
		return s.db.Create(config).Error
	} else if r2.Error != nil {
		return r2.Error
	}

	// Update existing DAO config
	existingConfig.Config = input.Raw
	existingConfig.UTime = utils.TimePtrNow()

	return s.db.Save(&existingConfig).Error
}

// MarkInactiveDAOs marks DAOs as inactive if they're not in the active list
func (s *DaoService) MarkInactiveDAOs(activeCodes map[string]bool) error {
	// Use a more efficient query to find and update inactive DAOs in one go
	result := s.db.Model(&dbmodels.Dao{}).
		Where("code NOT IN ? AND state != ?", getMapKeys(activeCodes), dbmodels.DaoStateInactive).
		Updates(map[string]interface{}{
			"state": dbmodels.DaoStateInactive,
			"utime": utils.TimePtrNow(),
		})

	if result.Error != nil {
		return result.Error
	}

	return nil
}

// UpdateDaoLastTrackingBlock updates the last tracking block for a DAO
func (s *DaoService) UpdateDaoOffsetTrackingProposal(daoCode string, offset int) error {
	return s.db.Model(&dbmodels.Dao{}).
		Where("code = ?", daoCode).
		Update("offset_tracking_proposal", offset).Error
}

// getMapKeys extracts keys from a map[string]bool
func getMapKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

type DaoConfigService struct {
	db *gorm.DB
}

func NewDaoConfigService() *DaoConfigService {
	return &DaoConfigService{
		db: database.GetDB(),
	}
}

func (s *DaoConfigService) Inspect(daoCode string) (*dbmodels.DgvDaoConfig, error) {
	var config dbmodels.DgvDaoConfig
	err := s.db.Where("dao_code = ?", daoCode).First(&config).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("dao config not found")
		}
		return nil, err
	}
	return &config, nil
}

func (s *DaoConfigService) StandardConfig(daoCode string) (*types.DaoConfig, error) {
	rawDaoConfig, err := s.Inspect(daoCode)
	if err != nil {
		return nil, err
	}

	var daoConfig types.DaoConfig
	if err := yaml.Unmarshal([]byte(rawDaoConfig.Config), &daoConfig); err != nil {
		slog.Error("failed to parse daoconfig", "err", err)
		return nil, err
	}
	return &daoConfig, nil
}

func (s *DaoConfigService) RawConfig(input gqlmodels.GetDaoConfigInput) (string, error) {
	daoConfig, err := s.Inspect(input.DaoCode)
	if err != nil {
		return "", err
	}

	format := gqlmodels.ConfigFormatYaml
	if input.Format != nil {
		format = *input.Format
	}

	if format == gqlmodels.ConfigFormatJSON {
		// Convert YAML to JSON
		var yamlData interface{}
		err := yaml.Unmarshal([]byte(daoConfig.Config), &yamlData)
		if err != nil {
			return "", errors.New("failed to convert YAML to JSON")
		}

		jsonData, err := json.MarshalIndent(yamlData, "", "  ")
		if err != nil {
			return "", errors.New("failed to convert YAML to JSON")
		}

		return string(jsonData), nil
	} else {
		// Default to YAML format
		return daoConfig.Config, nil
	}
}

type UserLikedDaoService struct {
	db         *gorm.DB
	daoService *DaoService
}

func NewUserLikedDaoService() *UserLikedDaoService {
	return &UserLikedDaoService{
		db:         database.GetDB(),
		daoService: NewDaoService(),
	}
}

// liked daos
func (s *UserLikedDaoService) LikedDaos(baseInput types.BasicInput[*string]) ([]*gqlmodels.Dao, error) {
	if baseInput.User == nil {
		return []*gqlmodels.Dao{}, nil
	}

	// First get liked dao codes by joining with dgv_dao to ensure only ACTIVE DAOs
	var daoCodes []string
	err := s.db.Table("dgv_user_liked_dao").
		Select("dgv_user_liked_dao.dao_code").
		Joins("INNER JOIN dgv_dao ON dgv_user_liked_dao.dao_code = dgv_dao.code").
		Where("dgv_user_liked_dao.user_id = ? AND dgv_dao.state = ?", baseInput.User.Id, "ACTIVE").
		Pluck("dao_code", &daoCodes).Error

	if err != nil {
		return nil, err
	}

	// If no liked DAOs found, return empty array
	if len(daoCodes) == 0 {
		return []*gqlmodels.Dao{}, nil
	}

	// Use existing ListDaos method to get the DAOs with all their data
	return s.daoService.ListDaos(types.BasicInput[*types.ListDaosInput]{
		Input: &types.ListDaosInput{
			Codes: &daoCodes,
		},
		User: baseInput.User,
	})
}

type UserSubscribedDaoService struct {
	db         *gorm.DB
	daoService *DaoService
}

func NewUserSubscribedDaoService() *UserSubscribedDaoService {
	return &UserSubscribedDaoService{
		db:         database.GetDB(),
		daoService: NewDaoService(),
	}
}

func (s *UserSubscribedDaoService) SubscribedDaos(baseInput types.BasicInput[*string]) ([]*gqlmodels.Dao, error) {
	if baseInput.User == nil {
		return []*gqlmodels.Dao{}, nil
	}

	// First get subscribed dao codes by joining with dgv_dao to ensure only ACTIVE DAOs
	var daoCodes []string
	err := s.db.Table("dgv_user_subscribed_dao").
		Select("dgv_user_subscribed_dao.dao_code").
		Joins("INNER JOIN dgv_dao ON dgv_user_subscribed_dao.dao_code = dgv_dao.code").
		Where("dgv_user_subscribed_dao.user_id = ? AND dgv_user_subscribed_dao.state = ? AND dgv_dao.state = ?",
			baseInput.User.Id, "SUBSCRIBED", "ACTIVE").
		Pluck("dao_code", &daoCodes).Error

	if err != nil {
		return nil, err
	}

	// If no subscribed DAOs found, return empty array
	if len(daoCodes) == 0 {
		return []*gqlmodels.Dao{}, nil
	}

	// Use existing ListDaos method to get the DAOs with all their data
	return s.daoService.ListDaos(types.BasicInput[*types.ListDaosInput]{
		Input: &types.ListDaosInput{
			Codes: &daoCodes,
		},
		User: baseInput.User,
	})
}

type DaoChipService struct {
	db *gorm.DB
}

func NewDaoChipService() *DaoChipService {
	return &DaoChipService{
		db: database.GetDB(),
	}
}

func (s *DaoChipService) SyncAgentChips(agentDaoConfigs []types.AgentDaoConfig) error {
	chipCode := dbmodels.ChipCodeAgent

	// Build map for quick lookup and prepare new chips
	agentDaoMap := make(map[string]types.AgentDaoConfig, len(agentDaoConfigs))
	newChips := make([]dbmodels.DgvDaoChip, 0, len(agentDaoConfigs))

	for _, agentDao := range agentDaoConfigs {
		agentDaoMap[agentDao.Code] = agentDao
		newChips = append(newChips, dbmodels.DgvDaoChip{
			ID:         utils.NextIDString(),
			DaoCode:    agentDao.Code,
			ChipCode:   chipCode,
			Flag:       "ENABLED",
			Additional: utils.ToJSON(agentDao),
			CTime:      time.Now(),
		})
	}

	// Use transaction for atomicity
	return s.db.Transaction(func(tx *gorm.DB) error {
		// First, delete all existing agent chips
		if err := tx.Where("chip_code = ?", chipCode).Delete(&dbmodels.DgvDaoChip{}).Error; err != nil {
			return err
		}

		// Then, batch insert new chips if any
		if len(newChips) > 0 {
			if err := tx.Create(&newChips).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (s *DaoChipService) StoreChipMetricsState(input types.StoreDaoChipMetricsStateInput) error {
	chipCode := dbmodels.ChipCodeMetricsState

	// Extract unique dao_codes from input.MetricsStates
	daoCodesMap := make(map[string]struct{})
	for _, state := range input.MetricsStates {
		daoCodesMap[state.DaoCode] = struct{}{}
	}

	// Convert to slice
	daoCodes := make([]string, 0, len(daoCodesMap))
	for daoCode := range daoCodesMap {
		daoCodes = append(daoCodes, daoCode)
	}

	// Delete existing records for these dao_codes with chip_code = METRICS_STATE
	if len(daoCodes) > 0 {
		if err := s.db.Where("dao_code IN ? AND chip_code = ?", daoCodes, chipCode).Delete(&dbmodels.DgvDaoChip{}).Error; err != nil {
			return err
		}
	}

	// Batch insert new records
	chips := make([]dbmodels.DgvDaoChip, 0, len(input.MetricsStates))
	for _, state := range input.MetricsStates {
		chip := dbmodels.DgvDaoChip{
			ID:         utils.NextIDString(),
			DaoCode:    state.DaoCode,
			ChipCode:   chipCode,
			Flag:       string(state.State), // MetricsState.State as flag
			Additional: utils.ToJSON(state), // Single MetricsState as JSON
			CTime:      time.Now(),
			UTime:      utils.TimePtrNow(),
		}
		chips = append(chips, chip)
	}

	// Batch insert
	if len(chips) > 0 {
		if err := s.db.Create(&chips).Error; err != nil {
			return err
		}
	}

	return nil
}
