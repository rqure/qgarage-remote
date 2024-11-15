package main

import (
	"math"
	"time"

	qdb "github.com/rqure/qdb/src"
)

func isApproximatelyEqual(a, b, tolerance float64) bool {
	return math.Abs(a-b) < tolerance
}

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
		Type:           "GarageDoor",
		Field:          "IsClosed",
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
	isClosed := qdb.ValueCast[*qdb.Bool](notification.Current.Value).Raw
	door := qdb.NewEntity(gsc.db, notification.Current.Id)

	if isClosed {
		door.GetField("Closing").PushBool(false)
		door.GetField("Moving").PushBool(false)
		door.GetField("PercentClosed").PushInt(100)
	} else {
		door.GetField("Moving").PushBool(true)
	}
}

func (gsc *GarageStatusCalculator) OnGarageDoorMoving(notification *qdb.DatabaseNotification) {
	moving := qdb.ValueCast[*qdb.Bool](notification.Current.Value).Raw
	closing := qdb.ValueCast[*qdb.Bool](notification.Context[0].Value).Raw
	percentClosed := qdb.ValueCast[*qdb.Int](notification.Context[1].Value).Raw
	timeToOpen := qdb.ValueCast[*qdb.Int](notification.Context[2].Value).Raw
	timeToClose := qdb.ValueCast[*qdb.Int](notification.Context[3].Value).Raw

	if timeToOpen == 0 || timeToClose == 0 {
		qdb.Warn("[GarageStatusCalculator::OnGarageDoorMoving] TimeToOpen and/or TimeToClose is 0 for door %s", notification.Current.Id)
		return
	}

	if moving {
		gsc.movingGarageDoorContext[notification.Current.Id] = MovingGarageDoorContext{
			InitialPercentClosed: percentClosed,
			PercentClosed:        percentClosed,
			Closing:              closing,
			TotalTimeToOpen:      timeToOpen,
			TotalTimeToClose:     timeToClose,
			ButtonPressTime:      notification.Current.WriteTime.AsTime(),
		}
	} else {
		delete(gsc.movingGarageDoorContext, notification.Current.Id)
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
			timeElapsedBeforePause := float64(0)
			if movingGarageDoor.InitialPercentClosed > 0 && movingGarageDoor.InitialPercentClosed < 100 {
				timeElapsedBeforePause = (float64(movingGarageDoor.InitialPercentClosed) / 100) * float64(movingGarageDoor.TotalTimeToOpen)
			}
			timeElapsedAfterResume := float64(time.Since(movingGarageDoor.ButtonPressTime).Milliseconds())
			remainingTimeToClose := max(float64(movingGarageDoor.TotalTimeToClose)-(timeElapsedBeforePause+timeElapsedAfterResume), 0)
			percentClosed = max(float64(movingGarageDoor.TotalTimeToClose)-remainingTimeToClose, 0) / float64(movingGarageDoor.TotalTimeToClose) * float64(100)
		} else {
			// Remaining time to Open = Total Time to Open - ( Time elapsed before pause + Time elapsed after resume)
			timeElapsedBeforePause := float64(0)
			if movingGarageDoor.InitialPercentClosed < 100 && movingGarageDoor.InitialPercentClosed > 0 {
				timeElapsedBeforePause = (float64(movingGarageDoor.InitialPercentClosed) / 100) * float64(movingGarageDoor.TotalTimeToClose)
			}
			timeElapsedAfterResume := float64(time.Since(movingGarageDoor.ButtonPressTime).Milliseconds())
			remainingTimeToOpen := max(float64(movingGarageDoor.TotalTimeToOpen)-(timeElapsedBeforePause+timeElapsedAfterResume), 0)
			percentClosed = float64(remainingTimeToOpen) / float64(movingGarageDoor.TotalTimeToOpen) * float64(100)
		}

		door := qdb.NewEntity(gsc.db, doorId)
		movingGarageDoor.PercentClosed = int64(percentClosed)
		door.GetField("PercentClosed").PushInt(movingGarageDoor.PercentClosed)

		if isApproximatelyEqual(percentClosed, 0.0, 1.0/float64(movingGarageDoor.TotalTimeToOpen)) ||
			isApproximatelyEqual(percentClosed, 100.0, 1.0/float64(movingGarageDoor.TotalTimeToClose)) {
			door.GetField("Moving").PushBool(false)
		}
	}
}
