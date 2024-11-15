package main

import (
	qdb "github.com/rqure/qdb/src"
)

type GarageController struct {
	db                 qdb.IDatabase
	isLeader           bool
	notificationTokens []qdb.INotificationToken
}

func NewGarageController(db qdb.IDatabase) *GarageController {
	return &GarageController{
		db: db,
	}
}

func (gc *GarageController) Init() {

}

func (gc *GarageController) Deinit() {

}

func (gc *GarageController) Reinitialize() {
	for _, token := range gc.notificationTokens {
		token.Unbind()
	}

	gc.notificationTokens = []qdb.INotificationToken{}

	if !gc.isLeader {
		return
	}

	gc.notificationTokens = append(gc.notificationTokens, gc.db.Notify(&qdb.DatabaseNotificationConfig{
		Type:          "GarageDoor",
		Field:         "ToggleTrigger",
		ContextFields: []string{"Closing", "Moving"},
	}, qdb.NewNotificationCallback(gc.OnToggleTrigger)))
}

func (gc *GarageController) OnSchemaUpdated() {
	gc.Reinitialize()
}

func (gc *GarageController) OnBecameLeader() {
	gc.isLeader = true
	gc.Reinitialize()
}

func (gc *GarageController) OnLostLeadership() {
	gc.isLeader = false

	for _, token := range gc.notificationTokens {
		token.Unbind()
	}

	gc.notificationTokens = []qdb.INotificationToken{}
}

func (gc *GarageController) DoWork() {
}

func (gc *GarageController) OnStatusDeviceChanged(notification *qdb.DatabaseNotification) {
	gc.Reinitialize()
}

func (gc *GarageController) OnToggleTrigger(notification *qdb.DatabaseNotification) {
	door := qdb.NewEntity(gc.db, notification.GetCurrent().Id)
	door.GetField("ToggleTriggerFn").PushInt()

	closing := qdb.ValueCast[*qdb.Bool](notification.Context[0].Value).Raw
	moving := qdb.ValueCast[*qdb.Bool](notification.Context[1].Value).Raw

	moving = !moving

	if moving {
		closing = !closing
	}

	door.GetField("Closing").PushBool(closing)

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
		door.GetField("Moving").PushBool(moving)
	}
}
