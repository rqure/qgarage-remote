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
	db       qdb.IDatabase
	isLeader bool
	events   chan IEvent
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
}

func (gc *GarageController) CloseDoor(event IEvent) {
}

func (gc *GarageController) OnDoorStatusChanged(event IEvent) {
}

func (gc *GarageController) OpenTTS(event IEvent) {
}

func (gc *GarageController) CloseTTS(event IEvent) {
}

func (gc *GarageController) OpenReminderTTS(event IEvent) {
}
