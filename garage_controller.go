package main

import (
	"context"

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
		store: store,
	}
}

func (gc *GarageController) Init(context.Context, app.Handle) {

}

func (gc *GarageController) Deinit(context.Context) {

}

func (gc *GarageController) Reinitialize(ctx context.Context) {
	for _, token := range gc.notificationTokens {
		token.Unbind(ctx)
	}

	gc.notificationTokens = []data.NotificationToken{}

	if !gc.isLeader {
		return
	}

	gc.notificationTokens = append(gc.notificationTokens, gc.store.Notify(
		ctx,
		notification.NewConfig().
			SetEntityType("Root").
			SetFieldName("SchemaUpdateTrigger"),
		notification.NewCallback(gc.OnSchemaUpdated)))

	gc.notificationTokens = append(gc.notificationTokens, gc.store.Notify(
		ctx,
		notification.NewConfig().
			SetEntityType("GarageDoor").
			SetFieldName("ToggleTrigger"),
		notification.NewCallback(gc.OnToggleTrigger)))
}

func (gc *GarageController) OnSchemaUpdated(ctx context.Context, n data.Notification) {
	gc.Reinitialize(ctx)
}

func (gc *GarageController) OnBecameLeader(ctx context.Context) {
	gc.isLeader = true
	gc.Reinitialize(ctx)
}

func (gc *GarageController) OnLostLeadership(ctx context.Context) {
	gc.isLeader = false

	for _, token := range gc.notificationTokens {
		token.Unbind(ctx)
	}

	gc.notificationTokens = []data.NotificationToken{}
}

func (gc *GarageController) DoWork(context.Context) {
}

func (gc *GarageController) OnStatusDeviceChanged(ctx context.Context, n data.Notification) {
	gc.Reinitialize(ctx)
}

func (gc *GarageController) OnToggleTrigger(ctx context.Context, n data.Notification) {
	door := binding.NewEntity(ctx, gc.store, n.GetCurrent().GetEntityId())
	door.GetField("ToggleTriggerFn").WriteInt(ctx)

	closing := n.GetContext(0).GetValue().GetBool()
	moving := n.GetContext(1).GetValue().GetBool()

	moving = !moving

	if moving {
		closing = !closing
	}

	door.GetField("Closing").WriteBool(ctx, closing)

	if !moving {
		door.GetField("Moving").WriteBool(ctx, moving, n.GetContext(1).GetWriteTime())
	} else {
		door.GetField("Moving").WriteBool(ctx, moving)
	}
}
