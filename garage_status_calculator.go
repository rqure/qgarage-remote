package main

import (
	"time"

	qdb "github.com/rqure/qdb/src"
)

type MovingGarageDoorContext struct {
	InitialPercentClosed int64
	PercentClosed        int64
	Closing              bool
	TotalTimeToOpen      int64
	TotalTimeToClose     int64
	ButtonPressTime      time.Time
}

type GarageStatusCalculator struct {
	db                      qdb.IDatabase
	isLeader                bool
	notificationTokens      []qdb.INotificationToken
	movingGarageDoorContext map[string]MovingGarageDoorContext
}

func NewGarageStatusCalculator(db qdb.IDatabase) *GarageStatusCalculator {
	return &GarageStatusCalculator{
		db: db,
	}
}

func (gsc *GarageStatusCalculator) Init() {

}

func (gsc *GarageStatusCalculator) Deinit() {

}

func (gsc *GarageStatusCalculator) Reinitialize() {
	for _, token := range gsc.notificationTokens {
		token.Unbind()
	}

	gsc.notificationTokens = []qdb.INotificationToken{}

	if !gsc.isLeader {
		return
	}

	gsc.notificationTokens = append(gsc.notificationTokens, gsc.db.Notify(&qdb.DatabaseNotificationConfig{
		Type:  "GarageDoor",
		Field: "OpenTrigger",
	}, qdb.NewNotificationCallback(gsc.OnOpenTrigger)))

	gsc.notificationTokens = append(gsc.notificationTokens, gsc.db.Notify(&qdb.DatabaseNotificationConfig{
		Type:  "GarageDoor",
		Field: "CloseTrigger",
	}, qdb.NewNotificationCallback(gsc.OnCloseTrigger)))

	gsc.notificationTokens = append(gsc.notificationTokens, gsc.db.Notify(&qdb.DatabaseNotificationConfig{
		Type:           "GarageDoor",
		Field:          "GarageDoorStatus",
		NotifyOnChange: true,
	}, qdb.NewNotificationCallback(gsc.OnGarageDoorStatusChanged)))

	gsc.notificationTokens = append(gsc.notificationTokens, gsc.db.Notify(&qdb.DatabaseNotificationConfig{
		Type:           "GarageDoor",
		Field:          "Moving",
		NotifyOnChange: true,
		ContextFields: []string{
			"Closing",
			"PercentClosed",
			"TimeToOpen",
			"TimeToClose",
		},
	}, qdb.NewNotificationCallback(gsc.OnGarageDoorMoving)))
}

func (gsc *GarageStatusCalculator) OnGarageDoorStatusChanged(notification *qdb.DatabaseNotification) {
	status := qdb.ValueCast[*qdb.GarageDoorState](notification.Current.Value)

	door := qdb.NewEntity(gsc.db, notification.Current.Id)

	if status.Raw == qdb.GarageDoorState_CLOSED {
		door.GetField("Moving").PushBool(false)
		door.GetField("PercentClosed").PushInt(100)
	} else if status.Raw == qdb.GarageDoorState_OPENED {
		door.GetField("Moving").PushBool(true)
	}
}

func (gsc *GarageStatusCalculator) OnGarageDoorMoving(notification *qdb.DatabaseNotification) {
	moving := qdb.ValueCast[*qdb.Bool](notification.Current.Value)
	closing := qdb.ValueCast[*qdb.Bool](notification.Context[0].Value)
	percentClosed := qdb.ValueCast[*qdb.Int](notification.Context[1].Value)
	timeToOpen := qdb.ValueCast[*qdb.Int](notification.Context[2].Value)
	timeToClose := qdb.ValueCast[*qdb.Int](notification.Context[3].Value)

	if moving.Raw {
		gsc.movingGarageDoorContext[notification.Current.Id] = MovingGarageDoorContext{
			InitialPercentClosed: percentClosed.Raw,
			PercentClosed:        percentClosed.Raw,
			Closing:              closing.Raw,
			TotalTimeToOpen:      timeToOpen.Raw,
			TotalTimeToClose:     timeToClose.Raw,
			ButtonPressTime:      notification.Current.WriteTime.AsTime(),
		}
	} else {
		delete(gsc.movingGarageDoorContext, notification.Current.Id)
	}
}

func (gsc *GarageStatusCalculator) OnOpenTrigger(notification *qdb.DatabaseNotification) {
	door := qdb.NewEntity(gsc.db, notification.Current.Id)
	door.GetField("Closing").PushBool(false)
	door.GetField("Moving").PushBool(true)
}

func (gsc *GarageStatusCalculator) OnCloseTrigger(notification *qdb.DatabaseNotification) {
	door := qdb.NewEntity(gsc.db, notification.Current.Id)
	door.GetField("Closing").PushBool(true)
	door.GetField("Moving").PushBool(true)
}

func (gsc *GarageStatusCalculator) OnSchemaUpdated() {
	gsc.Reinitialize()
}

func (gsc *GarageStatusCalculator) OnBecameLeader() {
	gsc.isLeader = true

	gsc.Reinitialize()
}

func (gsc *GarageStatusCalculator) OnLostLeadership() {
	gsc.isLeader = false

	for _, token := range gsc.notificationTokens {
		token.Unbind()
	}
}

func (gsc *GarageStatusCalculator) DoWork() {
	if !gsc.isLeader {
		return
	}

	for doorId, movingGarageDoor := range gsc.movingGarageDoorContext {
		if movingGarageDoor.Closing {
			// Remaining time to Close = Total Time to Close - ( Time elapsed before pause + Time elapsed after resume)
			timeElapsedBeforePause := (movingGarageDoor.InitialPercentClosed / 100) * movingGarageDoor.TotalTimeToOpen
			timeElapsedAfterResume := time.Since(movingGarageDoor.ButtonPressTime).Milliseconds()
			remainingTimeToClose := max(movingGarageDoor.TotalTimeToClose-(timeElapsedBeforePause+timeElapsedAfterResume), 0)
			movingGarageDoor.PercentClosed = max(movingGarageDoor.TotalTimeToClose-remainingTimeToClose, 0) / movingGarageDoor.TotalTimeToClose * 100
		} else {
			// Remaining time to Open = Total Time to Open - ( Time elapsed before pause + Time elapsed after resume)
			timeElapsedBeforePause := (movingGarageDoor.InitialPercentClosed / 100) * movingGarageDoor.TotalTimeToClose
			timeElapsedAfterResume := time.Since(movingGarageDoor.ButtonPressTime).Milliseconds()
			remainingTimeToOpen := max(movingGarageDoor.TotalTimeToOpen-(timeElapsedBeforePause+timeElapsedAfterResume), 0)
			movingGarageDoor.PercentClosed = remainingTimeToOpen / movingGarageDoor.TotalTimeToOpen * 100
		}

		door := qdb.NewEntity(gsc.db, doorId)
		door.GetField("PercentClosed").PushInt(movingGarageDoor.PercentClosed)

		if movingGarageDoor.PercentClosed == 0 {
			door.GetField("Moving").PushBool(false)
		}
	}
}
