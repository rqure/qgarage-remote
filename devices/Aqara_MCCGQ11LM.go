package devices

import qdb "github.com/rqure/qdb/src"

type Aqara_MCCGQ11LM struct {
}

func (d *Aqara_MCCGQ11LM) GetModel() string {
	return "AqaraMCCGQ11LM"
}

func (d *Aqara_MCCGQ11LM) GetNotificationSettings(door qdb.IEntity, device qdb.IEntity) (*qdb.DatabaseNotificationConfig, qdb.INotificationCallback) {
	config := &qdb.DatabaseNotificationConfig{
		Id:             device.GetId(),
		Field:          "Contact",
		NotifyOnChange: true,
	}

	callback := qdb.NewNotificationCallback(func(notification *qdb.DatabaseNotification) {
		oldStatus := door.GetField("GarageDoorStatus").PullValue(&qdb.GarageDoorState{}).(*qdb.GarageDoorState)
		contact := &qdb.Bool{}
		err := notification.Current.Value.UnmarshalTo(contact)
		if err != nil {
			qdb.Warn("[Aqara_MCCGQ11LM::GetNotificationSettings] Failed to unmarshal contact value: %s", err)
			return
		}

		newStatus := &qdb.GarageDoorState{Raw: qdb.GarageDoorState_CLOSED}
		if !contact.Raw {
			newStatus.Raw = qdb.GarageDoorState_OPENED
		}

		if oldStatus.Raw != newStatus.Raw {
			door.GetField("GarageDoorStatus").PushValue(newStatus)
		}
	})

	return config, callback
}
