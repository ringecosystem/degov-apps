package services

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"

	"github.com/ringecosystem/degov-apps/graph/model"
	"github.com/ringecosystem/degov-apps/internal"
	"github.com/ringecosystem/degov-apps/internal/database"
	"github.com/ringecosystem/degov-apps/internal/utils"
	"github.com/ringecosystem/degov-apps/models"
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

func (s *DaoService) GetDaos() ([]*model.Dao, error) {
	var dbDaos []model.Dao
	if err := s.db.Table("dgv_dao").Find(&dbDaos).Error; err != nil {
		return nil, err
	}

	var daos []*model.Dao
	for _, dbDao := range dbDaos {
		liked := false
		subscribed := false
		dao := &model.Dao{
			ID:         dbDao.ID,
			ChainID:    dbDao.ChainID,
			ChainName:  dbDao.ChainName,
			Name:       dbDao.Name,
			Code:       dbDao.Code,
			State:      dbDao.State,
			ConfigLink: dbDao.ConfigLink,
			TimeSyncd:  dbDao.TimeSyncd,
			Ctime:      dbDao.Ctime,
			Utime:      dbDao.Utime,
			Liked:      &liked,
			Subscribed: &subscribed,
		}
		daos = append(daos, dao)
	}

	return daos, nil
}

func (s *DaoService) RefreshDaoAndConfig(input types.RefreshDaoAndConfigInput) error {
	return nil
}

// UpsertDao inserts or updates a DAO in the database
func (s *DaoService) UpsertDao(dao *models.Dao) error {
	var existingDao models.Dao
	result := s.db.Where("code = ?", dao.Code).First(&existingDao)

	if result.Error == gorm.ErrRecordNotFound {
		// Insert new DAO
		dao.CTime = time.Now()
		return s.db.Create(dao).Error
	} else if result.Error != nil {
		return result.Error
	}

	// Update existing DAO
	dao.ID = existingDao.ID
	dao.CTime = existingDao.CTime
	dao.UTime = utils.TimePtrNow()

	return s.db.Save(dao).Error
}

// UpsertDaoConfig inserts or updates a DAO config in the database
func (s *DaoService) UpsertDaoConfig(daoCode string, rawConfig string) error {
	var existingConfig models.DgvDaoConfig
	result := s.db.Where("code = ?", daoCode).First(&existingConfig)

	if result.Error == gorm.ErrRecordNotFound {
		// Insert new DAO config
		config := &models.DgvDaoConfig{
			ID:     internal.NextIDString(),
			Code:   daoCode,
			Config: rawConfig,
			CTime:  time.Now(),
		}
		return s.db.Create(config).Error
	} else if result.Error != nil {
		return result.Error
	}

	// Update existing DAO config
	existingConfig.Config = rawConfig
	existingConfig.UTime = utils.TimePtrNow()

	return s.db.Save(&existingConfig).Error
}

// MarkInactiveDAOs marks DAOs as inactive if they're not in the active list
func (s *DaoService) MarkInactiveDAOs(activeCodes map[string]bool) error {
	// Use a more efficient query to find and update inactive DAOs in one go
	result := s.db.Model(&models.Dao{}).
		Where("code NOT IN ? AND state != ?", getMapKeys(activeCodes), "INACTIVE").
		Updates(map[string]interface{}{
			"state": "INACTIVE",
			"utime": utils.TimePtrNow(),
		})

	if result.Error != nil {
		return result.Error
	}

	return nil
}

// CreateDaoFromConfig creates a DAO model from config information
func (s *DaoService) CreateDaoFromConfig(daoCode string, daoConfig *types.DaoConfig, tags []string, configURL string) *models.Dao {
	return &models.Dao{
		ID:        internal.NextIDString(),
		ChainID:   daoConfig.Chain.ID,
		ChainName: daoConfig.Chain.Name,
		Name:      daoConfig.Name,
		Code:      daoCode,
		Seq:       0,
		State:     "ACTIVE",
		Tags: func() string {
			if tagsJSON, err := json.Marshal(tags); err == nil {
				return string(tagsJSON)
			}
			return ""
		}(),
		ConfigLink: configURL,
		TimeSyncd:  utils.TimePtrNow(),
	}
}

// getMapKeys extracts keys from a map[string]bool
func getMapKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
