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
		store:                store,
		lastDoorOpenReminder: make(map[string]time.Time),
		openReminderInterval: 5 * time.Minute,
	}
}

func (tc *TTSController) Init(context.Context, app.Handle) {

}

func (tc *TTSController) Deinit(context.Context) {

}

func (tc *TTSController) Reinitialize(ctx context.Context) {
	for _, token := range tc.notificationTokens {
		token.Unbind(ctx)
	}

	tc.notificationTokens = []data.NotificationToken{}

	tc.notificationTokens = append(tc.notificationTokens, tc.store.Notify(
		ctx,
		notification.NewConfig().
			SetEntityType("Root").
			SetFieldName("SchemaUpdateTrigger"),
		notification.NewCallback(tc.OnSchemaUpdated)))

	tc.notificationTokens = append(tc.notificationTokens, tc.store.Notify(
		ctx,
		notification.NewConfig().
			SetEntityType("GarageDoor").
			SetFieldName("IsClosed").
			SetNotifyOnChange(true),
		notification.NewCallback(tc.OnGarageDoorStatusChanged)))

	tc.notificationTokens = append(tc.notificationTokens, tc.store.Notify(
		ctx,
		notification.NewConfig().
			SetEntityType("GarageController").
			SetFieldName("OpenReminderInterval"),
		notification.NewCallback(tc.OnOpenReminderIntervalChanged)))

	garageControllers := query.New(tc.store).ForType("GarageController").Execute(ctx)

	for _, garageController := range garageControllers {
		tc.openReminderInterval = time.Duration(garageController.GetField("OpenReminderInterval").ReadInt(ctx)) * time.Minute
	}
}

func (tc *TTSController) OnSchemaUpdated(ctx context.Context, n data.Notification) {
	tc.Reinitialize(ctx)
}

func (tc *TTSController) OnBecameLeader(ctx context.Context) {
	tc.isLeader = true

	tc.Reinitialize(ctx)
}

func (tc *TTSController) OnLostLeadership(ctx context.Context) {
	tc.isLeader = false

	for _, token := range tc.notificationTokens {
		token.Unbind(ctx)
	}
}

func (tc *TTSController) DoWork(ctx context.Context) {
	if !tc.isLeader {
		return
	}

	for doorName, lastReminder := range tc.lastDoorOpenReminder {
		if time.Since(lastReminder) > tc.openReminderInterval {
			tc.DoTTS(ctx, doorName, OpenReminderTTS)
			tc.lastDoorOpenReminder[doorName] = time.Now()
		}
	}
}

func (tc *TTSController) OnGarageDoorStatusChanged(ctx context.Context, notification data.Notification) {
	isClosed := notification.GetCurrent().GetValue().GetBool()

	doorName := binding.NewEntity(ctx, tc.store, notification.GetCurrent().GetEntityId()).GetName()
	if !isClosed {
		tc.lastDoorOpenReminder[doorName] = time.Now()
		tc.DoTTS(ctx, doorName, OpenTTS)
	} else {
		delete(tc.lastDoorOpenReminder, doorName)
		tc.DoTTS(ctx, doorName, CloseTTS)
	}
}

func (tc *TTSController) OnOpenReminderIntervalChanged(ctx context.Context, n data.Notification) {
	interval := n.GetCurrent().GetValue().GetInt()

	if interval < 1 {
		interval = 1
	}

	tc.openReminderInterval = time.Duration(interval) * time.Minute
}

func (tc *TTSController) DoTTS(ctx context.Context, doorName string, ttsType TTSType) {
	garageControllers := query.New(tc.store).ForType("GarageController").Execute(ctx)

	for _, garageController := range garageControllers {
		tts := garageController.GetField(string(ttsType)).ReadString(ctx)

		if tts == "" {
			return
		}

		// Replace instances of {Door} with the door name
		tts = strings.ReplaceAll(tts, "{Door}", doorName)

		// Perform TTS
		multi := binding.NewMulti(tc.store)
		alertControllers := query.New(multi).ForType("AlertController").Execute(ctx)
		for _, alertController := range alertControllers {
			alertController.GetField("ApplicationName").WriteString(ctx, qdb.GetApplicationName())
			alertController.GetField("Description").WriteString(ctx, tts)
			alertController.GetField("TTSLanguage").WriteString(ctx, "en") // TODO: Get this from the store
			alertController.GetField("TTSAlert").WriteBool(ctx, strings.Contains(os.Getenv("ALERTS"), "TTS"))
			alertController.GetField("EmailAlert").WriteBool(ctx, strings.Contains(os.Getenv("ALERTS"), "EMAIL"))
			alertController.GetField("SendTrigger").WriteInt(ctx)
		}
	}
}
