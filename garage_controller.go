package main

import (
	qdb "github.com/rqure/qdb/src"
	"github.com/rqure/qgarage/devices"
	"github.com/rqure/qgarage/events"
)

type GarageController struct {
	db                 qdb.IDatabase
	isLeader           bool
	events             chan events.IEvent
	notificationTokens []qdb.INotificationToken
}

func NewGarageController(db qdb.IDatabase) *GarageController {
	return &GarageController{
		db:     db,
		events: make(chan events.IEvent, 1024),
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
		Type:  "GarageDoor",
		Field: "StatusDevice",
	}, qdb.NewNotificationCallback(gc.OnStatusDeviceChanged)))

	gc.notificationTokens = append(gc.notificationTokens, gc.db.Notify(&qdb.DatabaseNotificationConfig{
		Type:  "GarageDoor",
		Field: "OpenTrigger",
		ContextFields: []string{
			"ControlDevice",
		},
	}, qdb.NewNotificationCallback(gc.OnOpenTrigger)))

	gc.notificationTokens = append(gc.notificationTokens, gc.db.Notify(&qdb.DatabaseNotificationConfig{
		Type:  "GarageDoor",
		Field: "CloseTrigger",
		ContextFields: []string{
			"ControlDevice",
		},
	}, qdb.NewNotificationCallback(gc.OnCloseTrigger)))

	doors := qdb.NewEntityFinder(gc.db).Find(qdb.SearchCriteria{
		EntityType: "GarageDoor",
		Conditions: []qdb.FieldConditionEval{
			qdb.NewReferenceCondition().Where("StatusDevice").IsNotEqualTo(&qdb.EntityReference{Raw: ""}),
		},
	})

	for _, door := range doors {
		statusDeviceId := door.GetField("StatusDevice").PullEntityReference()
		statusDeviceEntity := qdb.NewEntity(gc.db, statusDeviceId)
		statusDevice := devices.MakeStatusDevice(statusDeviceEntity.GetType())

		if statusDevice == nil {
			qdb.Warn("[GarageController::Reinitialize] Status device not found for door %s (%s)", door.GetId(), door.GetName())
			continue
		}

		config, callback := statusDevice.GetNotificationSettings(door, statusDeviceEntity)
		gc.notificationTokens = append(gc.notificationTokens, gc.db.Notify(config, callback))
	}
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
	for {
		select {
		case event := <-gc.events:
			switch event.GetType() {
			case events.OpenCommand:
				gc.OpenDoor(event)
			case events.CloseCommand:
				gc.CloseDoor(event)
			case events.OpenTTS:
				gc.OpenTTS(event)
			case events.CloseTTS:
				gc.CloseTTS(event)
			case events.OpenReminderTTS:
				gc.OpenReminderTTS(event)
			case events.WriteDB:
				gc.WriteDB(event)
			}
		default:
			return
		}
	}
}

func (gc *GarageController) OpenDoor(event events.IEvent) {
	if !gc.isLeader {
		return
	}
}

func (gc *GarageController) CloseDoor(event events.IEvent) {
	if !gc.isLeader {
		return
	}
}

func (gc *GarageController) OpenTTS(event events.IEvent) {
	if !gc.isLeader {
		return
	}
}

func (gc *GarageController) CloseTTS(event events.IEvent) {
	if !gc.isLeader {
		return
	}
}

func (gc *GarageController) OpenReminderTTS(event events.IEvent) {
	if !gc.isLeader {
		return
	}
}

func (gc *GarageController) WriteDB(event events.IEvent) {
	if !gc.isLeader {
		return
	}
}

func (gc *GarageController) OnStatusDeviceChanged(notification *qdb.DatabaseNotification) {
	gc.Reinitialize()
}

func (gc *GarageController) OnOpenTrigger(notification *qdb.DatabaseNotification) {

}

func (gc *GarageController) OnCloseTrigger(notification *qdb.DatabaseNotification) {

}
