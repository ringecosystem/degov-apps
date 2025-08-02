package models

import (
	"time"
)

const (
	NotificationTypeNewProposal     = "NEW_PROPOSAL"
	NotificationTypeVote            = "VOTE"
	NotificationTypeStatus          = "STATUS"
	NotificationTypeVoteEndReminder = "VOTE_END_REMINDER"
)

const (
	NotificationStatusSentOk   = "SENT_OK"
	NotificationStatusSentFail = "SENT_FAIL"
)

type NotificationRecord struct {
	ID         string    `gorm:"column:id;type:varchar(50);primaryKey" json:"id"`
	ChainID    int       `gorm:"column:chain_id;not null" json:"chain_id"`
	ChainName  string    `gorm:"column:chain_name;type:varchar(255);not null" json:"chain_name"`
	DaoName    string    `gorm:"column:dao_name;type:varchar(255);not null" json:"dao_name"`
	DaoCode    string    `gorm:"column:dao_code;type:varchar(50);not null" json:"dao_code"`
	Type       string    `gorm:"column:type;type:varchar(50);not null" json:"type"`
	TargetID   *string   `gorm:"column:target_id;type:varchar(255)" json:"target_id,omitempty"`
	UserID     string    `gorm:"column:user_id;type:varchar(50);not null" json:"user_id"`
	Status     string    `gorm:"column:status;type:varchar(50);not null" json:"status"`
	Message    *string   `gorm:"column:message;type:text" json:"message,omitempty"`
	RetryTimes int       `gorm:"column:retry_times;not null;default:0" json:"retry_times"`
	CTime      time.Time `gorm:"column:ctime;default:now()" json:"ctime"`
}

func (NotificationRecord) TableName() string {
	return "dgv_notification_record"
}

const (
	ChannelTypeEmail = "EMAIL"
	ChannelTypeSMS   = "SMS"
	ChannelTypePush  = "PUSH"
)

type UserChannel struct {
	ID           string    `gorm:"column:id;type:varchar(50);primaryKey" json:"id"`
	UserID       string    `gorm:"column:user_id;type:varchar(50);not null" json:"user_id"`
	Verified     int       `gorm:"column:verified;not null;default:0" json:"verified"`
	ChannelType  string    `gorm:"column:channel_type;type:varchar(50);not null" json:"channel_type"`
	ChannelValue string    `gorm:"column:channel_value;type:varchar(500);not null" json:"channel_value"`
	Payload      *string   `gorm:"column:payload;type:text" json:"payload,omitempty"`
	CTime        time.Time `gorm:"column:ctime;default:now()" json:"ctime"`
}

func (UserChannel) TableName() string {
	return "dgv_user_channel"
}
