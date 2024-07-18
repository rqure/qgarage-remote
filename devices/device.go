package devices

import (
	qdb "github.com/rqure/qdb/src"
)

// A control device for a garage door would allow us
// to send commands to the device to open or close the door
type IControlDevice interface {
	GetModel() string

	// The channel is used to send commands to the device
	// These would normally be in the form of database writes
	Open(controlDeviceEntityId string, writeRequests chan *qdb.DatabaseRequest)
	Close(controlDeviceEntityId string, writeRequests chan *qdb.DatabaseRequest)
}

// A status device for a garage door would report
// the current status of the door. ie. whether it is
// open or closed
type IStatusDevice interface {
	GetModel() string

	// door is used to update the door entity in the database
	// device is used to setup status change notifications for the device
	// returns the notification config and callback
	GetNotificationSettings(door qdb.IEntity, device qdb.IEntity) (*qdb.DatabaseNotificationConfig, qdb.INotificationCallback)
}

func GetAllStatusDevices() []IStatusDevice {
	return []IStatusDevice{
		&Aqara_MCCGQ11LM{},
	}
}

func GetAllControlDevices() []IControlDevice {
	return []IControlDevice{
		&Aqara_LLKZMK12LM{},
	}
}

func MakeStatusDevice(model string) IStatusDevice {
	devices := GetAllStatusDevices()
	for _, device := range devices {
		if device.GetModel() == model {
			return device
		}
	}
	return nil
}

func MakeControlDevice(model string) IControlDevice {
	devices := GetAllControlDevices()
	for _, device := range devices {
		if device.GetModel() == model {
			return device
		}
	}
	return nil
}
