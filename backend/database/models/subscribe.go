package dbmodels

import "time"

type SubscribeFeatureName string

const (
	SubscribeFeatureEnableProposal          SubscribeFeatureName = "ENABLE_PROPOSAL"
	SubscribeFeatureEnableVotingEndReminder SubscribeFeatureName = "ENABLE_VOTING_END_REMINDER"
	SubscribeFeatureEnableVoted             SubscribeFeatureName = "ENABLE_VOTED"
	SubscribeFeatureEnableStateChanged      SubscribeFeatureName = "ENABLE_STATE_CHANGED"
)

type SubscribeState string

const (
	SubscribeStateActive   SubscribeState = "ACTIVE"
	SubscribeStateInactive SubscribeState = "INACTIVE"
)

type UserSubscribedDao struct {
	ID          string         `gorm:"column:id;type:varchar(50);primaryKey" json:"id"`
	ChainID     int            `gorm:"column:chain_id;not null" json:"chain_id"`
	DaoCode     string         `gorm:"column:dao_code;type:varchar(255);not null" json:"dao_code"`
	UserID      string         `gorm:"column:user_id;type:varchar(50);not null;" json:"user_id"`
	UserAddress string         `gorm:"column:user_address;type:varchar(255);not null;" json:"user_address"`
	State       SubscribeState `gorm:"column:state;type:varchar(50);not null" json:"state"` // SUBSCRIBED, UNSUBSCRIBED
	CTime       time.Time      `gorm:"column:ctime;default:now()" json:"ctime"`
	UTime       time.Time      `gorm:"column:utime;autoUpdateTime" json:"utime"`
}

func (UserSubscribedDao) TableName() string {
	return "dgv_user_subscribed_dao"
}

type UserSubscribedProposal struct {
	ID          string         `gorm:"column:id;type:varchar(50);primaryKey" json:"id"`
	ChainID     int            `gorm:"column:chain_id;not null" json:"chain_id"`
	DaoCode     string         `gorm:"column:dao_code;type:varchar(255);not null" json:"dao_code"`
	UserID      string         `gorm:"column:user_id;type:varchar(50);not null" json:"user_id"`
	UserAddress string         `gorm:"column:user_address;type:varchar(255);not null" json:"user_address"`
	State       SubscribeState `gorm:"column:state;type:varchar(50);not null" json:"state"` // { ACTIVE, INACTIVE }
	ProposalID  string         `gorm:"column:proposal_id;type:varchar(255);not null" json:"proposal_id"`
	CTime       time.Time      `gorm:"column:ctime;default:now()" json:"ctime"`
	UTime       time.Time      `gorm:"column:utime;autoUpdateTime" json:"utime"`
}

func (UserSubscribedProposal) TableName() string {
	return "dgv_user_subscribed_proposal"
}

type SubscribeFeature struct {
	ID          string               `gorm:"column:id;primaryKey" json:"id"`
	ChainID     int                  `gorm:"column:chain_id" json:"chain_id"`
	DaoCode     string               `gorm:"column:dao_code" json:"dao_code"`
	UserID      string               `gorm:"column:user_id" json:"user_id"`
	UserAddress string               `gorm:"column:user_address" json:"user_address"`
	Feature     SubscribeFeatureName `gorm:"column:feature" json:"feature"`         // subscribe feature
	Strategy    string               `gorm:"column:strategy" json:"strategy"`       // subscribe strategy
	ProposalID  *string              `gorm:"column:proposal_id" json:"proposal_id"` // nullable
	Ctime       time.Time            `gorm:"column:ctime;autoCreateTime" json:"ctime"`
}

// TableName overrides the default table name
func (SubscribeFeature) TableName() string {
	return "dgv_subscribed_feature"
}
