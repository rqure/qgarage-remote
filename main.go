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
	GarageState          *qmq.QMQGarageDoorState `qmq:"garage:state"`
	GarageRequestedState *qmq.QMQGarageDoorState `qmq:"garage:requested-state"`
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

type SetHandler struct{}

type TickHandler struct{}

func (h *TickHandler) OnTick(c qmq.WebServiceContext) {
	schema := c.Schema().(*Schema)

	mqttMessage := new(qmq.QMQMqttMessage)
	popped := c.App().Consumer("garage:sensor:queue").Pop(mqttMessage)
	if popped == nil {
		return
	}
	defer popped.Ack()

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

	c.App().Logger().Advise(fmt.Sprintf("Garage door state changed to: %s", schema.GarageState.Value.String()))

	c.NotifyClients(qmq.DataUpdateResponse{
		Data: qmq.KeyValueResponse{
			Key:   "garage:state",
			Value: schema.GarageState,
		},
	})
}

func (h *SetHandler) OnSet(c qmq.WebServiceContext, key string, value interface{}) {
	schema := c.Schema().(*Schema)

	if key != "garage:requested-state" {
		return
	}

	if schema.GarageState.Value == value.(qmq.QMQGarageDoorStateEnum) {
		return
	}

	c.App().Logger().Advise(fmt.Sprintf("Garage door requested state changed to: %s", value.(qmq.QMQGarageDoorStateEnum).String()))
	// c.App().Producer("garage:command:exchange").Push(&qmq.QMQMqttMessage{
	// 	Topic: "garage/command",
	// })
}

func main() {
	os.Setenv("QMQ_ADDR", "localhost:6379")

	service := qmq.NewWebService()
	service.Initialize(new(Schema))
	service.App().AddConsumer("garage:sensor:queue").Initialize()
	service.App().AddProducer("garage:command:exchange").Initialize(1)
	service.AddTickHandler(new(TickHandler))
	service.AddSetHandler(new(SetHandler))
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
