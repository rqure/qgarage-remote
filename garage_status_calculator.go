package main

import (
	"time"

	qdb "github.com/rqure/qdb/src"
	"google.golang.org/protobuf/types/known/anypb"
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
		db:                      db,
		movingGarageDoorContext: make(map[string]MovingGarageDoorContext),
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
		ContextFields: []string{
			"Moving",
			"PercentClosed",
		},
	}, qdb.NewNotificationCallback(gsc.OnOpenTrigger)))

	gsc.notificationTokens = append(gsc.notificationTokens, gsc.db.Notify(&qdb.DatabaseNotificationConfig{
		Type:  "GarageDoor",
		Field: "CloseTrigger",
		ContextFields: []string{
			"Moving",
			"PercentClosed",
		},
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

	if timeToOpen.Raw == 0 || timeToClose.Raw == 0 {
		qdb.Warn("[GarageStatusCalculator::OnGarageDoorMoving] TimeToOpen and/or TimeToClose is 0 for door %s", notification.Current.Id)
		return
	}

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
	moving := qdb.ValueCast[*qdb.Bool](notification.Context[0].Value)
	percentClosed := qdb.ValueCast[*qdb.Int](notification.Context[1].Value)

	door := qdb.NewEntity(gsc.db, notification.Current.Id)
	if moving.Raw {
		moving.Raw = false
	} else {
		moving.Raw = true
		door.GetField("Closing").PushBool(false)
	}

	if percentClosed.Raw == 100 || percentClosed.Raw == 0 {
		door.GetField("Moving").PushBool(moving.Raw)
	} else {
		// Set Moving without changing the writetime
		// This is important because writetime signifies when the button
		// was originally pressed, and we want to keep that information
		// to calculate the PercentClosed correctly
		value, err := anypb.New(moving)
		if err != nil {
			qdb.Error("[GarageStatusCalculator::OnCloseTrigger] Failed to create Any from bool: %v", err)
			return
		}

		gsc.db.Write([]*qdb.DatabaseRequest{
			{
				Id:        door.GetId(),
				Field:     "Moving",
				Value:     value,
				WriteTime: &qdb.Timestamp{Raw: notification.Context[0].WriteTime},
			},
		})
	}
}

func (gsc *GarageStatusCalculator) OnCloseTrigger(notification *qdb.DatabaseNotification) {
	moving := qdb.ValueCast[*qdb.Bool](notification.Context[0].Value)
	percentClosed := qdb.ValueCast[*qdb.Int](notification.Context[1].Value)

	door := qdb.NewEntity(gsc.db, notification.Current.Id)
	if moving.Raw {
		moving.Raw = false
	} else {
		moving.Raw = true
		door.GetField("Closing").PushBool(true)
	}

	if percentClosed.Raw == 100 || percentClosed.Raw == 0 {
		door.GetField("Moving").PushBool(moving.Raw)
	} else {
		// Set Moving without changing the writetime
		// This is important because writetime signifies when the button
		// was originally pressed, and we want to keep that information
		// to calculate the PercentClosed correctly
		value, err := anypb.New(moving)
		if err != nil {
			qdb.Error("[GarageStatusCalculator::OnCloseTrigger] Failed to create Any from bool: %v", err)
			return
		}

		gsc.db.Write([]*qdb.DatabaseRequest{
			{
				Id:        door.GetId(),
				Field:     "Moving",
				Value:     value,
				WriteTime: &qdb.Timestamp{Raw: notification.Context[0].WriteTime},
			},
		})
	}
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
		var percentClosed float64

		if movingGarageDoor.Closing {
			// Remaining time to Close = Total Time to Close - ( Time elapsed before pause + Time elapsed after resume)
			timeElapsedBeforePause := int64(0)
			if movingGarageDoor.InitialPercentClosed > 0 {
				timeElapsedBeforePause = (movingGarageDoor.InitialPercentClosed / 100) * movingGarageDoor.TotalTimeToOpen
			}
			timeElapsedAfterResume := time.Since(movingGarageDoor.ButtonPressTime).Milliseconds()
			remainingTimeToClose := max(movingGarageDoor.TotalTimeToClose-(timeElapsedBeforePause+timeElapsedAfterResume), 0)
			percentClosed = float64(max(movingGarageDoor.TotalTimeToClose-remainingTimeToClose, 0)) / float64(movingGarageDoor.TotalTimeToClose) * float64(100)
		} else {
			// Remaining time to Open = Total Time to Open - ( Time elapsed before pause + Time elapsed after resume)
			timeElapsedBeforePause := int64(0)
			if movingGarageDoor.InitialPercentClosed < 100 {
				timeElapsedBeforePause = (movingGarageDoor.InitialPercentClosed / 100) * movingGarageDoor.TotalTimeToClose
			}
			timeElapsedAfterResume := time.Since(movingGarageDoor.ButtonPressTime).Milliseconds()
			remainingTimeToOpen := max(movingGarageDoor.TotalTimeToOpen-(timeElapsedBeforePause+timeElapsedAfterResume), 0)
			percentClosed = float64(remainingTimeToOpen) / float64(movingGarageDoor.TotalTimeToOpen) * float64(100)
		}

		door := qdb.NewEntity(gsc.db, doorId)
		movingGarageDoor.PercentClosed = int64(percentClosed)
		door.GetField("PercentClosed").PushInt(movingGarageDoor.PercentClosed)

		if percentClosed == 0 || percentClosed == 100 {
			door.GetField("Moving").PushBool(false)
		}
	}
}
