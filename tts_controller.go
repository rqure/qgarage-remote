package main

import (
	"context"
	"os"
	"strings"
	"time"

	qdb "github.com/rqure/qdb/src"
	"github.com/rqure/qlib/pkg/app"
	"github.com/rqure/qlib/pkg/data"
	"github.com/rqure/qlib/pkg/data/binding"
	"github.com/rqure/qlib/pkg/data/notification"
	"github.com/rqure/qlib/pkg/data/query"
)

type TTSType string

const (
	OpenTTS         TTSType = "OpenTTS"
	CloseTTS        TTSType = "CloseTTS"
	OpenReminderTTS TTSType = "OpenReminderTTS"
)

type TTSController struct {
	store                data.Store
	isLeader             bool
	notificationTokens   []data.NotificationToken
	lastDoorOpenReminder map[string]time.Time
	openReminderInterval time.Duration
}

func NewTTSController(store data.Store) *TTSController {
	return &TTSController{
		db:                   db,
		lastDoorOpenReminder: make(map[string]time.Time),
		openReminderInterval: 5 * time.Minute,
	}
}

func (tc *TTSController) Init(context.Context, app.Handle) {

}

func (tc *TTSController) Deinit(context.Context) {

}

func (tc *TTSController) Reinitialize() {
	for _, token := range tc.notificationTokens {
		token.Unbind()
	}

	tc.notificationTokens = []data.NotificationToken{}

	tc.notificationTokens = append(tc.notificationTokens, tc.db.Notify(&qdb.DatabaseNotificationConfig{
		Type:           "GarageDoor",
		Field:          "IsClosed",
		NotifyOnChange: true,
	}, notification.NewCallback(tc.OnGarageDoorStatusChanged)))

	tc.notificationTokens = append(tc.notificationTokens, tc.db.Notify(
		ctx,
		notification.NewConfig().
			SetEntityType("GarageController").
			SetFieldName("OpenReminderInterval"), notification.NewCallback(tc.OnOpenReminderIntervalChanged)))

	garageControllers := query.New(tc.db).Find(qdb.SearchCriteria{
		EntityType: "GarageController",
		Conditions: []qdb.FieldConditionEval{},
	})

	for _, garageController := range garageControllers {
		tc.openReminderInterval = time.Duration(garageController.GetField("OpenReminderInterval").ReadInt(ctx)) * time.Minute
	}
}

func (tc *TTSController) OnSchemaUpdated() {
	tc.Reinitialize()
}

func (tc *TTSController) OnBecameLeader(context.Context) {
	tc.isLeader = true

	tc.Reinitialize()
}

func (tc *TTSController) OnLostLeadership(context.Context) {
	tc.isLeader = false

	for _, token := range tc.notificationTokens {
		token.Unbind()
	}
}

func (tc *TTSController) DoWork(context.Context) {
	if !tc.isLeader {
		return
	}

	for doorName, lastReminder := range tc.lastDoorOpenReminder {
		if time.Since(lastReminder) > tc.openReminderInterval {
			tc.DoTTS(doorName, OpenReminderTTS)
			tc.lastDoorOpenReminder[doorName] = time.Now()
		}
	}
}

func (tc *TTSController) OnGarageDoorStatusChanged(ctx context.Context, notification data.Notification) {
	isClosed := notification.GetCurrent().GetValue().GetBool()

	doorName := binding.NewEntity(ctx, tc.db, notification.GetCurrent().GetEntityId()).GetName()
	if !isClosed {
		tc.lastDoorOpenReminder[doorName] = time.Now()
		tc.DoTTS(doorName, OpenTTS)
	} else {
		delete(tc.lastDoorOpenReminder, doorName)
		tc.DoTTS(doorName, CloseTTS)
	}
}

func (tc *TTSController) OnOpenReminderIntervalChanged(ctx context.Context, notification data.Notification) {
	interval := qdb.ValueCast[*qdb.Int](notification.GetCurrent().GetValue())

	if interval.Raw < 1 {
		interval.Raw = 1
	}

	tc.openReminderInterval = time.Duration(interval.Raw) * time.Minute
}

func (tc *TTSController) DoTTS(doorName string, ttsType TTSType) {
	garageControllers := query.New(tc.db).Find(qdb.SearchCriteria{
		EntityType: "GarageController",
		Conditions: []qdb.FieldConditionEval{},
	})

	for _, garageController := range garageControllers {
		tts := garageController.GetField(string(ttsType)).ReadString(ctx)

		if tts == "" {
			return
		}

		// Replace instances of {Door} with the door name
		tts = strings.ReplaceAll(tts, "{Door}", doorName)

		// Perform TTS
		alertControllers := query.New(tc.db).Find(qdb.SearchCriteria{
			EntityType: "AlertController",
			Conditions: []qdb.FieldConditionEval{},
		})

		for _, alertController := range alertControllers {
			tc.db.Write([]*qdb.DatabaseRequest{
				{
					Id:    alertController.GetId(),
					Field: "ApplicationName",
					Value: qdb.NewStringValue(qdb.GetApplicationName()),
				},
				{
					Id:    alertController.GetId(),
					Field: "Description",
					Value: qdb.NewStringValue(tts),
				},
				{
					Id:    alertController.GetId(),
					Field: "TTSAlert",
					Value: qdb.NewBoolValue(strings.Contains(os.Getenv("ALERTS"), "TTS")),
				},
				{
					Id:    alertController.GetId(),
					Field: "EmailAlert",
					Value: qdb.NewBoolValue(strings.Contains(os.Getenv("ALERTS"), "EMAIL")),
				},
				{
					Id:    alertController.GetId(),
					Field: "SendTrigger",
					Value: qdb.NewIntValue(0),
				},
			})
		}
	}
}
