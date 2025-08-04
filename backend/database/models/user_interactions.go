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

type UserSubscribedDao struct {
	ID                      string    `gorm:"column:id;type:varchar(50);primaryKey" json:"id"`
	ChainID                 int       `gorm:"column:chain_id;not null" json:"chain_id"`
	DaoCode                 string    `gorm:"column:dao_code;type:varchar(255);not null" json:"dao_code"`
	UserID                  string    `gorm:"column:user_id;type:varchar(50);not null;uniqueIndex:uq_dgv_user_subscribe_uid_code,priority:1" json:"user_id"`
	UserAddress             string    `gorm:"column:user_address;type:varchar(255);not null;uniqueIndex:uq_dgv_user_subscribe_address_code,priority:1" json:"user_address"`
	State                   string    `gorm:"column:state;type:varchar(50);not null" json:"state"` // SUBSCRIBED, UNSUBSCRIBED
	EnableNewProposal       int       `gorm:"column:enable_new_proposal;not null;default:1" json:"enable_new_proposal"`
	EnableVotingEndReminder int       `gorm:"column:enable_voting_end_reminder;not null;default:0" json:"enable_voting_end_reminder"`
	CTime                   time.Time `gorm:"column:ctime;default:now()" json:"ctime"`
}

func (UserSubscribedDao) TableName() string {
	return "dgv_user_subscribed_dao"
}
