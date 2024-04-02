package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"time"

	qmq "github.com/rqure/qmq/src"
	"google.golang.org/protobuf/proto"
)

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

type GarageSensorNotificationProcessor struct{}

func (h *GarageSensorNotificationProcessor) OnTick(c qmq.WebServiceContext) {
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

	state := qmq.QMQGarageDoorStateEnum_GARAGE_DOOR_STATE_OPEN
	if sensorData.Contact {
		state = qmq.QMQGarageDoorStateEnum_GARAGE_DOOR_STATE_CLOSED
	}

	c.Schema().Set("garage:state", &qmq.QMQGarageDoorState{
		Value: state,
	})

	c.App().Logger().Advise(fmt.Sprintf("Garage door state changed to: %s", state.String()))

	c.NotifyClients([]string{"garage:state"})
}

type GarageCommandHandler struct{}

func (h *GarageCommandHandler) OnSet(c qmq.WebServiceContext, key string, value proto.Message) {
	if key != "garage:requested-state" {
		return
	}

	c.NotifyClients([]string{key})

	c.App().Logger().Advise(fmt.Sprintf("Garage door requested state changed to: %v", value))
	// c.App().Producer("garage:command:exchange").Push(&qmq.QMQMqttMessage{
	// 	Topic: "garage/command",
	// })
}

func main() {
	os.Setenv("QMQ_ADDR", "localhost:6379")

	service := qmq.NewWebService()
	service.Initialize(qmq.NewSchema(map[string]proto.Message{
		"garage:state":           new(qmq.QMQGarageDoorState),
		"garage:requested-state": new(qmq.QMQGarageDoorState),
	}))
	service.App().AddConsumer("garage:sensor:queue").Initialize()
	service.App().AddProducer("garage:command:exchange").Initialize(10)
	service.AddTickHandler(new(GarageSensorNotificationProcessor))
	service.AddSetHandler(new(GarageCommandHandler))
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
