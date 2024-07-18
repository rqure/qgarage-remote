package main

import (
	"strings"
	"time"

	qdb "github.com/rqure/qdb/src"
)

type TTSType string

const (
	OpenTTS         TTSType = "OpenTTS"
	CloseTTS        TTSType = "CloseTTS"
	OpenReminderTTS TTSType = "OpenReminderTTS"
)

type TTSController struct {
	db                   qdb.IDatabase
	isLeader             bool
	notificationTokens   []qdb.INotificationToken
	lastDoorOpenReminder map[string]time.Time
	openReminderInterval time.Duration
}

func NewTTSController(db qdb.IDatabase) *TTSController {
	return &TTSController{
		db:                   db,
		lastDoorOpenReminder: make(map[string]time.Time),
		openReminderInterval: 5 * time.Minute,
	}
}

func (tc *TTSController) Init() {

}

func (tc *TTSController) Deinit() {

}

func (tc *TTSController) Reinitialize() {
	for _, token := range tc.notificationTokens {
		token.Unbind()
	}

	tc.notificationTokens = []qdb.INotificationToken{}

	tc.notificationTokens = append(tc.notificationTokens, tc.db.Notify(&qdb.DatabaseNotificationConfig{
		Type:           "GarageDoor",
		Field:          "GarageDoorStatus",
		NotifyOnChange: true,
	}, qdb.NewNotificationCallback(tc.OnGarageDoorStatusChanged)))

	tc.notificationTokens = append(tc.notificationTokens, tc.db.Notify(&qdb.DatabaseNotificationConfig{
		Type:  "GarageController",
		Field: "OpenReminderInterval",
	}, qdb.NewNotificationCallback(tc.OnOpenReminderIntervalChanged)))

	garageControllers := qdb.NewEntityFinder(tc.db).Find(qdb.SearchCriteria{
		EntityType: "GarageController",
		Conditions: []qdb.FieldConditionEval{},
	})

	for _, garageController := range garageControllers {
		tc.openReminderInterval = time.Duration(garageController.GetField("OpenReminderInterval").PullInt()) * time.Minute
	}
}

func (tc *TTSController) OnSchemaUpdated() {
	tc.Reinitialize()
}

func (tc *TTSController) OnBecameLeader() {
	tc.isLeader = true

	tc.Reinitialize()
}

func (tc *TTSController) OnLostLeadership() {
	tc.isLeader = false

	for _, token := range tc.notificationTokens {
		token.Unbind()
	}
}

func (tc *TTSController) DoWork() {
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

func (tc *TTSController) OnGarageDoorStatusChanged(notification *qdb.DatabaseNotification) {
	status := qdb.ValueCast[*qdb.GarageDoorState](notification.Current.Value)

	doorName := qdb.NewEntity(tc.db, notification.Current.Id).GetName()
	if status.Raw == qdb.GarageDoorState_OPENED {
		tc.lastDoorOpenReminder[doorName] = time.Now()
		tc.DoTTS(doorName, OpenTTS)
	} else {
		delete(tc.lastDoorOpenReminder, doorName)
		tc.DoTTS(doorName, CloseTTS)
	}
}

func (tc *TTSController) OnOpenReminderIntervalChanged(notification *qdb.DatabaseNotification) {
	interval := qdb.ValueCast[*qdb.Int](notification.Current.Value)

	if interval.Raw < 1 {
		interval.Raw = 1
	}

	tc.openReminderInterval = time.Duration(interval.Raw) * time.Minute
}

func (tc *TTSController) DoTTS(doorName string, ttsType TTSType) {
	garageControllers := qdb.NewEntityFinder(tc.db).Find(qdb.SearchCriteria{
		EntityType: "GarageController",
		Conditions: []qdb.FieldConditionEval{},
	})

	for _, garageController := range garageControllers {
		tts := garageController.GetField(string(ttsType)).PullString()

		if tts == "" {
			return
		}

		// Replace instances of {Door} with the door name
		tts = strings.ReplaceAll(tts, "{Door}", doorName)

		// Perform TTS
		audioControllers := qdb.NewEntityFinder(tc.db).Find(qdb.SearchCriteria{
			EntityType: "AudioController",
			Conditions: []qdb.FieldConditionEval{},
		})

		for _, audioController := range audioControllers {
			audioController.GetField("TextToSpeech").PushString(tts)
		}
	}
}
