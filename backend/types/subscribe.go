package types

import (
	"time"

	dbmodels "github.com/ringecosystem/degov-apps/database/models"
)

type ListSubscribeUserInput struct {
	Feature    dbmodels.SubscribeFeatureName
	Strategies []string
	DaoCode    string
	ProposalId *string
	// EventTime is the timestamp of the event; only users who subscribed
	// before or at this time should be returned.
	EventTime *time.Time
	Limit     int
	Offset    int
}

type ListSubscribedUserOutput struct {
	UserID      string
	UserAddress string
	ChainID     int
	DaoCode     string
	CTime       time.Time `gorm:"column:ctime"`
}
