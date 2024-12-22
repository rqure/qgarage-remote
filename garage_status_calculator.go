package main

import (
	"context"
	"math"
	"time"

	"github.com/rqure/qlib/pkg/app"
	"github.com/rqure/qlib/pkg/data"
	"github.com/rqure/qlib/pkg/data/binding"
	"github.com/rqure/qlib/pkg/data/notification"
	"github.com/rqure/qlib/pkg/log"
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
	store                   data.Store
	isLeader                bool
	notificationTokens      []data.NotificationToken
	movingGarageDoorContext map[string]MovingGarageDoorContext
}

func NewGarageStatusCalculator(store data.Store) *GarageStatusCalculator {
	return &GarageStatusCalculator{
		store:                   store,
		movingGarageDoorContext: make(map[string]MovingGarageDoorContext),
	}
}

func (gsc *GarageStatusCalculator) Init(context.Context, app.Handle) {

}

func (gsc *GarageStatusCalculator) Deinit(context.Context) {

}

func (gsc *GarageStatusCalculator) Reinitialize(ctx context.Context) {
	for _, token := range gsc.notificationTokens {
		token.Unbind(ctx)
	}

	gsc.notificationTokens = []data.NotificationToken{}

	if !gsc.isLeader {
		return
	}

	gsc.notificationTokens = append(gsc.notificationTokens, gsc.store.Notify(
		ctx,
		notification.NewConfig().
			SetEntityType("Root").
			SetFieldName("SchemaUpdateTrigger"),
		notification.NewCallback(gsc.OnSchemaUpdated)))

	gsc.notificationTokens = append(gsc.notificationTokens, gsc.store.Notify(
		ctx,
		notification.NewConfig().
			SetEntityType("GarageDoor").
			SetFieldName("IsClosed").
			SetNotifyOnChange(true),
		notification.NewCallback(gsc.OnGarageDoorStatusChanged)))

	gsc.notificationTokens = append(gsc.notificationTokens, gsc.store.Notify(
		ctx,
		notification.NewConfig().
			SetEntityType("GarageDoor").
			SetFieldName("Moving").
			SetNotifyOnChange(true).
			SetContextFields(
				"Closing",
				"PercentClosed",
				"TimeToOpen",
				"TimeToClose",
			),
		notification.NewCallback(gsc.OnGarageDoorMoving)))
}

func (gsc *GarageStatusCalculator) OnGarageDoorStatusChanged(ctx context.Context, notification data.Notification) {
	isClosed := notification.GetCurrent().GetValue().GetBool()
	door := binding.NewEntity(ctx, gsc.store, notification.GetCurrent().GetEntityId())

	if isClosed {
		door.GetField("Closing").WriteBool(ctx, false)
		door.GetField("Moving").WriteBool(ctx, false)
		door.GetField("PercentClosed").WriteInt(ctx, 100)
	} else {
		door.GetField("Moving").WriteBool(ctx, true)
	}
}

func (gsc *GarageStatusCalculator) OnGarageDoorMoving(ctx context.Context, notification data.Notification) {
	moving := notification.GetCurrent().GetValue().GetBool()
	closing := notification.GetContext(0).GetValue().GetBool()
	percentClosed := notification.GetContext(1).GetValue().GetInt()
	timeToOpen := notification.GetContext(2).GetValue().GetInt()
	timeToClose := notification.GetContext(3).GetValue().GetInt()

	if timeToOpen == 0 || timeToClose == 0 {
		log.Warn("TimeToOpen and/or TimeToClose is 0 for door %s", notification.GetCurrent().GetEntityId())
		return
	}

	if moving {
		gsc.movingGarageDoorContext[notification.GetCurrent().GetEntityId()] = MovingGarageDoorContext{
			InitialPercentClosed: percentClosed,
			PercentClosed:        percentClosed,
			Closing:              closing,
			TotalTimeToOpen:      timeToOpen,
			TotalTimeToClose:     timeToClose,
			ButtonPressTime:      notification.GetCurrent().GetWriteTime(),
		}
	} else {
		delete(gsc.movingGarageDoorContext, notification.GetCurrent().GetEntityId())
	}
}

func (gsc *GarageStatusCalculator) OnSchemaUpdated(ctx context.Context, n data.Notification) {
	gsc.Reinitialize(ctx)
}

func (gsc *GarageStatusCalculator) OnBecameLeader(ctx context.Context) {
	gsc.isLeader = true

	gsc.Reinitialize(ctx)
}

func (gsc *GarageStatusCalculator) OnLostLeadership(ctx context.Context) {
	gsc.isLeader = false

	for _, token := range gsc.notificationTokens {
		token.Unbind(ctx)
	}
}

func (gsc *GarageStatusCalculator) DoWork(ctx context.Context) {
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

		door := binding.NewEntity(ctx, gsc.store, doorId)
		movingGarageDoor.PercentClosed = int64(percentClosed)
		door.GetField("PercentClosed").WriteInt(ctx, movingGarageDoor.PercentClosed)

		if isApproximatelyEqual(percentClosed, 0.0, 1.0/float64(movingGarageDoor.TotalTimeToOpen)) ||
			isApproximatelyEqual(percentClosed, 100.0, 1.0/float64(movingGarageDoor.TotalTimeToClose)) {
			door.GetField("Moving").WriteBool(ctx, false)
		}
	}
}
