package services

import (
	"gorm.io/gorm"

	"github.com/ringecosystem/degov-apps/graph/model"
	"github.com/ringecosystem/degov-apps/internal/database"
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
