package main

import qdb "github.com/rqure/qdb/src"

type EventType int

const (
	OpenCommand EventType = iota
	CloseCommand
	DoorStatusChanged
	OpenTTS
	CloseTTS
	OpenReminderTTS
)

type IEvent interface {
	GetType() EventType
	Context() interface{}
}

type GarageController struct {
	db                 qdb.IDatabase
	isLeader           bool
	events             chan IEvent
	notificationTokens []qdb.INotificationToken
}

func NewGarageController(db qdb.IDatabase) *GarageController {
	return &GarageController{
		db:     db,
		events: make(chan IEvent, 1024),
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

	doors := qdb.NewEntityFinder(gc.db).Find(qdb.SearchCriteria{
		EntityType: "GarageDoor",
		Conditions: []qdb.FieldConditionEval{
			qdb.NewReferenceCondition().Where("StatusDevice").IsNotEqualTo(&qdb.EntityReference{Raw: ""}),
		},
	})

	for _, door := range doors {
		door.GetField("StatusDevice->")
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
}

func (gc *GarageController) DoWork() {
	for {
		select {
		case event := <-gc.events:
			switch event.GetType() {
			case OpenCommand:
				gc.OpenDoor(event)
			case CloseCommand:
				gc.CloseDoor(event)
			case DoorStatusChanged:
				gc.OnDoorStatusChanged(event)
			case OpenTTS:
				gc.OpenTTS(event)
			case CloseTTS:
				gc.CloseTTS(event)
			case OpenReminderTTS:
				gc.OpenReminderTTS(event)
			}
		default:
			return
		}
	}
}

func (gc *GarageController) OpenDoor(event IEvent) {
	if !gc.isLeader {
		return
	}
}

func (gc *GarageController) CloseDoor(event IEvent) {
	if !gc.isLeader {
		return
	}
}

func (gc *GarageController) OnDoorStatusChanged(event IEvent) {
	if !gc.isLeader {
		return
	}
}

func (gc *GarageController) OpenTTS(event IEvent) {
	if !gc.isLeader {
		return
	}
}

func (gc *GarageController) CloseTTS(event IEvent) {
	if !gc.isLeader {
		return
	}
}

func (gc *GarageController) OpenReminderTTS(event IEvent) {
	if !gc.isLeader {
		return
	}
}
