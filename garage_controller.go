package main

import (
	qdb "github.com/rqure/qdb/src"
	"github.com/rqure/qgarage/devices"
)

type GarageController struct {
	db                 qdb.IDatabase
	isLeader           bool
	writeRequests      chan *qdb.DatabaseRequest
	notificationTokens []qdb.INotificationToken
}

func NewGarageController(db qdb.IDatabase) *GarageController {
	return &GarageController{
		db:            db,
		writeRequests: make(chan *qdb.DatabaseRequest, 1024),
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
		case writeRequest := <-gc.writeRequests:
			if !gc.isLeader {
				continue
			}

			gc.db.Write([]*qdb.DatabaseRequest{writeRequest})
		default:
			return
		}
	}
}

func (gc *GarageController) OnStatusDeviceChanged(notification *qdb.DatabaseNotification) {
	gc.Reinitialize()
}

func (gc *GarageController) OnOpenTrigger(notification *qdb.DatabaseNotification) {
	controlDeviceEntityRef := &qdb.EntityReference{}
	err := notification.Context[0].Value.UnmarshalTo(controlDeviceEntityRef)
	if err != nil {
		qdb.Error("[GarageController::OnOpenTrigger] Unable to unmarshal control device entity id: %s", err)
		return
	}

	controlDeviceEntity := qdb.NewEntity(gc.db, controlDeviceEntityRef.Raw)
	controlDevice := devices.MakeControlDevice(controlDeviceEntity.GetType())
	if controlDevice == nil {
		qdb.Warn("[GarageController::OnOpenTrigger] Control device not found for entity %s (%s)", controlDeviceEntity.GetId(), controlDeviceEntity.GetName())
		return
	}

	controlDevice.Open(controlDeviceEntityRef.Raw, gc.writeRequests)
}

func (gc *GarageController) OnCloseTrigger(notification *qdb.DatabaseNotification) {
	controlDeviceEntityRef := &qdb.EntityReference{}
	err := notification.Context[0].Value.UnmarshalTo(controlDeviceEntityRef)
	if err != nil {
		qdb.Error("[GarageController::OnCloseTrigger] Unable to unmarshal control device entity id: %s", err)
		return
	}

	controlDeviceEntity := qdb.NewEntity(gc.db, controlDeviceEntityRef.Raw)
	controlDevice := devices.MakeControlDevice(controlDeviceEntity.GetType())
	if controlDevice == nil {
		qdb.Warn("[GarageController::OnCloseTrigger] Control device not found for entity %s (%s)", controlDeviceEntity.GetId(), controlDeviceEntity.GetName())
		return
	}

	controlDevice.Close(controlDeviceEntityRef.Raw, gc.writeRequests)
}
