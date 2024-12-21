package main

import (
	"context"

	qdb "github.com/rqure/qdb/src"
	"github.com/rqure/qlib/pkg/app"
	"github.com/rqure/qlib/pkg/data"
	"github.com/rqure/qlib/pkg/data/binding"
	"github.com/rqure/qlib/pkg/data/notification"
)

type GarageController struct {
	store              data.Store
	isLeader           bool
	notificationTokens []data.NotificationToken
}

func NewGarageController(store data.Store) *GarageController {
	return &GarageController{
		db: db,
	}
}

func (gc *GarageController) Init(context.Context, app.Handle) {

}

func (gc *GarageController) Deinit(context.Context) {

}

func (gc *GarageController) Reinitialize() {
	for _, token := range gc.notificationTokens {
		token.Unbind()
	}

	gc.notificationTokens = []data.NotificationToken{}

	if !gc.isLeader {
		return
	}

	gc.notificationTokens = append(gc.notificationTokens, gc.db.Notify(
ctx,
notification.NewConfig().
SetEntityType(        "GarageDoor").
SetFieldName(        "ToggleTrigger"),
	}, notification.NewCallback(gc.OnToggleTrigger)))
}

func (gc *GarageController) OnSchemaUpdated() {
	gc.Reinitialize()
}

func (gc *GarageController) OnBecameLeader(context.Context) {
	gc.isLeader = true
	gc.Reinitialize()
}

func (gc *GarageController) OnLostLeadership(context.Context) {
	gc.isLeader = false

	for _, token := range gc.notificationTokens {
		token.Unbind()
	}

	gc.notificationTokens = []data.NotificationToken{}
}

func (gc *GarageController) DoWork(context.Context) {
}

func (gc *GarageController) OnStatusDeviceChanged(ctx context.Context, notification data.Notification) {
	gc.Reinitialize()
}

func (gc *GarageController) OnToggleTrigger(ctx context.Context, notification data.Notification) {
	door := binding.NewEntity(ctx, gc.db, notification.GetCurrent().Id)
	door.GetField("ToggleTriggerFn").WriteInt(ctx)

	closing := notification.GetContext(0).GetValue().GetBool()
	moving := notification.GetContext(1).GetValue().GetBool()

	moving = !moving

	if moving {
		closing = !closing
	}

	door.GetField("Closing").WriteBool(ctx, closing)

	if !moving {
		gc.db.Write([]*qdb.DatabaseRequest{
			{
				Id:        door.GetId(),
				Field:     "Moving",
				Value:     qdb.NewBoolValue(moving),
				WriteTime: &qdb.Timestamp{Raw: notification.Context[1].WriteTime},
			},
		})
	} else {
		door.GetField("Moving").WriteBool(ctx, moving)
	}
}
