package main

import (
	"encoding/json"
	"fmt"

	qmq "github.com/rqure/qmq/src"
)

type DoorRelayToMqttTransformer struct {
	logger qmq.Logger
}

func NewDoorRelayToMqttTransformer(logger qmq.Logger) qmq.Transformer {
	return &DoorRelayToMqttTransformer{
		logger: logger,
	}
}

func (t *DoorRelayToMqttTransformer) Transform(i interface{}) interface{} {
	j, ok := i.(*GarageDoorRelayJson)
	if !ok {
		t.logger.Error(fmt.Sprintf("DoorRelayToMqttTransformer.Transform: invalid input type %T", i))
		return nil
	}

	b, err := json.Marshal(j)
	if err != nil {
		t.logger.Error(fmt.Sprintf("DoorRelayToMqttTransformer.Transform: failed to marshal garage door relay json: %v", err))
		return nil
	}

	m := &qmq.MqttMessage{
		Topic:    "zigbee2mqtt/garage-door-relay/set",
		Payload:  b,
		Qos:      0,
		Retained: false,
	}

	return m
}
