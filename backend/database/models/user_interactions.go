package dbmodels

import (
	"time"
)

type UserLikedDao struct {
	ID          string    `gorm:"column:id;type:varchar(50);primaryKey" json:"id"`
	DaoCode     string    `gorm:"column:dao_code;type:varchar(50);not null" json:"dao_code"`
	UserID      string    `gorm:"column:user_id;type:varchar(50);not null;uniqueIndex:uq_dgv_user_liked_dao_code_uid,priority:1" json:"user_id"`
	UserAddress string    `gorm:"column:user_address;type:varchar(255);not null;uniqueIndex:uq_dgv_user_liked_dao_code_address,priority:2" json:"user_address"`
	CTime       time.Time `gorm:"column:ctime;default:now()" json:"ctime"`
}

func (UserLikedDao) TableName() string {
	return "dgv_user_liked_dao"
}
