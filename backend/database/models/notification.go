package dbmodels

import (
	"time"
)

type NotificationType string
type NotificationRecordState string
type UserChannelType string
type NotificationEventState string

const (
	NotificationTypeNewProposal     NotificationType = "NEW_PROPOSAL"
	NotificationTypeVote            NotificationType = "VOTE"
	NotificationTypeStateChanged    NotificationType = "STATE_CHANGED"
	NotificationTypeVoteEndReminder NotificationType = "VOTE_END_REMINDER"
)

const (
	NotificationRecordStateWait     NotificationRecordState = "WAIT"
	NotificationRecordStateSentOk   NotificationRecordState = "SENT_OK"
	NotificationRecordStateSentFail NotificationRecordState = "SENT_FAIL"
)

const (
	NotificationEventStatePending   NotificationEventState = "PENDING"
	NotificationEventStateProgress  NotificationEventState = "PROGRESS"
	NotificationEventStateCompleted NotificationEventState = "COMPLETED"
)

type NotificationRecord struct {
	ID          string                  `gorm:"column:id;type:varchar(50);primaryKey" json:"id"`
	ChainID     int                     `gorm:"column:chain_id;not null" json:"chain_id"`
	DaoCode     string                  `gorm:"column:dao_code;type:varchar(255);not null" json:"dao_code"`
	Type        NotificationType        `gorm:"column:type;type:varchar(50);not null" json:"type"`
	ProposalID  string                  `gorm:"column:proposal_id;type:varchar(255);not null" json:"proposal_id"`
	VoteID      *string                 `gorm:"column:vote_id;type:varchar(255)" json:"vote_id,omitempty"`
	UserID      string                  `gorm:"column:user_id;type:varchar(50);not null;uniqueIndex:uq_notification_record_event_id_user_id" json:"user_id"`
	UserAddress string                  `gorm:"column:user_address;type:varchar(255);not null" json:"user_address"`
	State       NotificationRecordState `gorm:"column:state;type:varchar(50);not null" json:"state"`
	Message     *string                 `gorm:"column:message;type:text" json:"message,omitempty"`
	Payload     *string                 `gorm:"column:payload;type:text" json:"payload,omitempty"`
	RetryTimes  int                     `gorm:"column:retry_times;not null;default:0" json:"retry_times"`
	CTime       time.Time               `gorm:"column:ctime;default:now()" json:"ctime"`
	EventID     string                  `gorm:"column:event_id;type:varchar(255);not null;uniqueIndex:uq_notification_record_event_id_user_id" json:"event_id"`
}

func (NotificationRecord) TableName() string {
	return "dgv_notification_record"
}

type NotificationEvent struct {
	ID         string                 `gorm:"column:id;type:varchar(50);primaryKey" json:"id"`
	ChainID    int                    `gorm:"column:chain_id;not null" json:"chain_id"`
	DaoCode    string                 `gorm:"column:dao_code;type:varchar(255);not null" json:"dao_code"`
	Type       NotificationType       `gorm:"column:type;type:varchar(50);not null" json:"type"`
	ProposalID string                 `gorm:"column:proposal_id;type:varchar(255);not null" json:"proposal_id"`
	VoteID     *string                `gorm:"column:vote_id;type:varchar(255)" json:"vote_id,omitempty"`
	Reached    int                    `gorm:"column:reached;not null;default:0" json:"reached"`
	State      NotificationEventState `gorm:"column:state;type:varchar(50);not null" json:"state"`
	Payload    string                 `gorm:"column:payload;type:text" json:"payload"`
	TimeEvent  time.Time              `gorm:"column:time_event" json:"time_event"`
	CTime      time.Time              `gorm:"column:ctime;default:now()" json:"ctime"`
}

func (NotificationEvent) TableName() string {
	return "dgv_notification_event"
}

const (
	ChannelTypeEmail   UserChannelType = "EMAIL"
	ChannelTypeWebhook UserChannelType = "WEBHOOK"
)

type UserChannel struct {
	ID           string          `gorm:"column:id;type:varchar(50);primaryKey" json:"id"`
	UserID       string          `gorm:"column:user_id;type:varchar(50);not null" json:"user_id"`
	Verified     int             `gorm:"column:verified;not null;default:0" json:"verified"`
	ChannelType  UserChannelType `gorm:"column:channel_type;type:varchar(50);not null" json:"channel_type"`
	ChannelValue string          `gorm:"column:channel_value;type:varchar(500);not null" json:"channel_value"`
	Payload      *string         `gorm:"column:payload;type:text" json:"payload,omitempty"`
	CTime        time.Time       `gorm:"column:ctime;default:now()" json:"ctime"`
}

func (UserChannel) TableName() string {
	return "dgv_user_channel"
}
