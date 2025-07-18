package models

import (
	"time"
)

type UserLikedDao struct {
	ID      string    `gorm:"column:id;type:varchar(50);primaryKey" json:"id"`
	DaoCode string    `gorm:"column:dao_code;type:varchar(50);not null;uniqueIndex:uq_dgv_like_dao,priority:1" json:"dao_code"`
	UserID  string    `gorm:"column:user_id;type:varchar(50);not null;uniqueIndex:uq_dgv_like_dao,priority:2" json:"user_id"`
	CTime   time.Time `gorm:"column:ctime;default:now()" json:"ctime"`
}

func (UserLikedDao) TableName() string {
	return "dgv_user_liked_dao"
}

type UserFollowedDao struct {
	ID                      string    `gorm:"column:id;type:varchar(50);primaryKey" json:"id"`
	ChainID                 int       `gorm:"column:chain_id;not null" json:"chain_id"`
	DaoCode                 string    `gorm:"column:dao_code;type:varchar(50);not null" json:"dao_code"`
	UserID                  string    `gorm:"column:user_id;type:varchar(50);not null;uniqueIndex:uq_dgv_notification,priority:1" json:"user_id"`
	EnableNewProposal       int       `gorm:"column:enable_new_proposal;not null;default:1" json:"enable_new_proposal"`
	EnableVotingEndReminder int       `gorm:"column:enable_voting_end_reminder;not null;default:0" json:"enable_voting_end_reminder"`
	CTime                   time.Time `gorm:"column:ctime;default:now()" json:"ctime"`
}

func (UserFollowedDao) TableName() string {
	return "dgv_user_followed_dao"
}
