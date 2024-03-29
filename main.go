package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"time"

	qmq "github.com/rqure/qmq/src"
)

type Schema struct {
	GarageState *qmq.QMQGarageDoorState `qmq:"garage:state"`
}

// Example JSON:
// {"battery":100,"contact":false,"device_temperature":25,"linkquality":87,"power_outage_count":5,"voltage":3085}
type GarageDoorSensorJson struct {
	Battery           int     `json:"battery"`
	Contact           bool    `json:"contact"`
	DeviceTemperature float32 `json:"device_temperature"`
	LinkQuality       int     `json:"linkquality"`
	PowerOutageCount  int     `json:"power_outage_count"`
	Voltage           int     `json:"voltage"`
}

type TickHandler struct{}

func (h *TickHandler) OnTick(c qmq.WebServiceContext) {
	schema := c.Schema().(Schema)

	mqttMessage := new(qmq.QMQMqttMessage)
	popped := c.App().Consumer("garage:door-sensor:queue").Pop(mqttMessage)
	if popped == nil {
		return
	}

	sensorData := new(GarageDoorSensorJson)
	err := json.Unmarshal(mqttMessage.Payload, sensorData)
	if err != nil {
		c.App().Logger().Warn(fmt.Sprintf("Failed to unmarshal garage door sensor data: %v", err))
		return
	}

	if sensorData.Contact {
		schema.GarageState.Value = qmq.QMQGarageDoorStateEnum_GARAGE_DOOR_STATE_CLOSED
	} else {
		schema.GarageState.Value = qmq.QMQGarageDoorStateEnum_GARAGE_DOOR_STATE_OPEN
	}

	c.NotifyClients(qmq.DataUpdateResponse{
		Data: qmq.KeyValueResponse{
			Key:   "garage:state",
			Value: schema.GarageState,
		},
	})

	popped.Ack()
}

func main() {
	service := qmq.NewWebService()
	service.Initialize(new(Schema))
	service.AddTickHandler(new(TickHandler))
	defer service.Deinitialize()

	tickRateMs, err := strconv.Atoi(os.Getenv("TICK_RATE_MS"))
	if err != nil {
		tickRateMs = 100
	}

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)

	ticker := time.NewTicker(time.Duration(tickRateMs) * time.Millisecond)
	for {
		select {
		case <-sigint:
			return
		case <-ticker.C:
			service.Tick()
		}
	}
}
