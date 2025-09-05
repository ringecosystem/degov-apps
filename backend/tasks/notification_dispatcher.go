package tasks

import (
	"fmt"
	"log/slog"

	dbmodels "github.com/ringecosystem/degov-apps/database/models"
	"github.com/ringecosystem/degov-apps/services"
	"github.com/ringecosystem/degov-apps/types"
)

type NotificationDispatcherTask struct {
	notificationService    *services.NotificationService
	templateService        *services.TemplateService
	notifierService        *services.NotifierService
	userInteractionService *services.UserInteractionService
}

func NewNotificationDispatcherTask() *NotificationDispatcherTask {
	return &NotificationDispatcherTask{
		notificationService:    services.NewNotificationService(),
		templateService:        services.NewTemplateService(),
		notifierService:        services.NewNotifierService(),
		userInteractionService: services.NewUserInteractionService(),
	}
}

func (t *NotificationDispatcherTask) Name() string {
	return "notification-dispatcher"
}

func (t *NotificationDispatcherTask) Execute() error {
	return t.dispatcherNotificationRecord()
}

func (t *NotificationDispatcherTask) dispatcherNotificationRecord() error {
	states := []dbmodels.NotificationRecordState{
		dbmodels.NotificationRecordStatePending,
	}
	records, err := t.notificationService.ListLimitRecords(types.ListLimitRecordsInput{
		Limit:  100,
		States: &states,
	})
	if err != nil {
		return err
	}
	for _, record := range records {
		verified := true
		channels, err := t.userInteractionService.ListChannel(types.BasicInput[types.ListChannelInput]{
			User: &types.UserSessInfo{
				Id: record.UserID,
			},
			Input: types.ListChannelInput{
				Verified: &verified,
			},
		})
		timesRetry := record.TimesRetry + 1

		if err != nil {
			slog.Error("Failed to list user channels", "user_id", record.UserID, "error", err)

			var message string
			if record.Message != nil {
				message = fmt.Sprintf("%s\n\n-------\n[%d] Failed to list user channels: %s", *record.Message, timesRetry, err.Error())
			} else {
				message = fmt.Sprintf("[%d] Failed to list user channels: %s", timesRetry, err.Error())
			}

			t.notificationService.UpdateRecordRetryTimes(types.UpdateRecordRetryTimes{
				ID:         record.ID,
				TimesRetry: timesRetry,
				Message:    message,
			})
			continue
		}

		if err := t.dispatchNotificationRecordByRecord(&record, channels); err != nil {
			slog.Error("Failed to dispatch notification record", "record_id", record.ID, "error", err)

			var message string
			if record.Message != nil {
				message = fmt.Sprintf("%s\n\n-------\n[%d] Failed to build notification record: %s", *record.Message, timesRetry, err.Error())
			} else {
				message = fmt.Sprintf("[%d] Failed to build notification record: %s", timesRetry, err.Error())
			}

			if err := t.notificationService.UpdateRecordRetryTimes(types.UpdateRecordRetryTimes{
				ID:         record.ID,
				TimesRetry: timesRetry,
				Message:    message,
			}); err != nil {
				slog.Error("Failed to update record retry times", "record_id", record.ID, "error", err)
			}

			if timesRetry > 4 {
				if err := t.notificationService.UpdateRecordState(types.UpdateRecordStateInput{
					ID:    record.ID,
					State: dbmodels.NotificationRecordStateSentFail,
				}); err != nil {
					slog.Error("Failed to update record state to failed", "record_id", record.ID, "error", err)
				}
			}
			continue
		}

		if err := t.notificationService.UpdateRecordState(types.UpdateRecordStateInput{
			ID:    record.ID,
			State: dbmodels.NotificationRecordStateSentOk,
		}); err != nil {
			slog.Error("Failed to update record state to send_ok", "record_id", record.ID, "error", err)
			continue
		}
	}
	return nil
}

func (t *NotificationDispatcherTask) dispatchNotificationRecordByRecord(record *dbmodels.NotificationRecord, channels []dbmodels.NotificationChannel) error {
	templateOutput, err := t.templateService.GenerateTemplateByNotificationRecord(record)
	if err != nil {
		return err
	}
	slog.Debug("Dispatch notification record", "record_id", record.ID, "template", templateOutput)

	for _, channel := range channels {
		if err := t.notifierService.Notify(types.NotifyInput{
			Type:     channel.ChannelType,
			To:       channel.ChannelValue,
			Template: templateOutput,
		}); err != nil {
			// todo: The best practice is to record the failure of a channel to avoid repeated pushes, but there will not be multiple channels for the time being
			slog.Warn(
				"Failed to notify",
				"channel_type", channel.ChannelType,
				"channel_to", channel.ChannelValue,
				"error", err,
			)
			return err
		}
	}
	return nil
}
