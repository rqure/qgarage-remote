package main

import (
	"encoding/json"

	qmq "github.com/rqure/qmq/src"
)

type MqttToDoorSensorTransformer struct {
	logger qmq.Logger
}

func NewMqttToDoorSensorTransformer(logger qmq.Logger) qmq.Transformer {
	return &MqttToDoorSensorTransformer{
		logger: logger,
	}
}

func (t *MqttToDoorSensorTransformer) Transform(i interface{}) interface{} {
	m, ok := i.(*qmq.MqttMessage)
	if !ok {
		t.logger.Error("MqttToDoorSensorTransformer.Transform: invalid input type")
		return nil
	}

	j := new(GarageDoorSensorJson)
	err := json.Unmarshal(m.Payload, j)
	if err != nil {
		t.logger.Error("MqttToDoorSensorTransformer.Transform: failed to unmarshal mqtt message to garage door sensor json")
		return nil
	}

	return j
}
