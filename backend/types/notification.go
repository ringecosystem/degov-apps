package types

import (
	dbmodels "github.com/ringecosystem/degov-apps/database/models"
	gqlmodels "github.com/ringecosystem/degov-apps/graph/models"
)

type InspectNotificationEventInput struct {
	DaoCode    string
	ProposalID string
	VoteID     *string
	Type       dbmodels.SubscribeFeatureName
	States     *[]dbmodels.NotificationEventState
}

type ListLimitEventsInput struct {
	Limit  int
	States *[]dbmodels.NotificationEventState
}

type ListLimitRecordsInput struct {
	Limit  int
	States *[]dbmodels.NotificationRecordState
}

type UpdateEventStateInput struct {
	ID    string
	State dbmodels.NotificationEventState
}

type UpdateEventRetryTimes struct {
	ID         string
	TimesRetry int
	Message    string
}

type UpdateRecordStateInput struct {
	ID    string
	State dbmodels.NotificationRecordState
}

type UpdateRecordRetryTimes struct {
	ID         string
	TimesRetry int
	Message    string
}

type NotifyProposalTemplateInput struct {
	Dao      gqlmodels.Dao
	Proposal dbmodels.ProposalTracking
}
