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

type Schema struct {
	GarageState          *qmq.QMQGarageDoorState
	GarageRequestedState *qmq.QMQGarageDoorState
}

func (s *Schema) Get(key string) proto.Message {
	switch key {
	case "garage:state":
		return s.GarageState
	case "garage:requested-state":
		return s.GarageRequestedState
	}
	return nil
}

func (s *Schema) Set(key string, value proto.Message) {
	switch key {
	case "garage:state":
		s.GarageState = value.(*qmq.QMQGarageDoorState)
	case "garage:requested-state":
		s.GarageRequestedState = value.(*qmq.QMQGarageDoorState)
	}
}

func (s *Schema) GetAllData(db *qmq.QMQConnection) {
	s.GarageState = new(qmq.QMQGarageDoorState)
	s.GarageRequestedState = new(qmq.QMQGarageDoorState)

	db.GetValue("garage:state", s.GarageState)
	db.GetValue("garage:requested-state", s.GarageRequestedState)
}

func (s *Schema) SetAllData(db *qmq.QMQConnection) {
	db.SetValue("garage:state", s.GarageState)
	db.SetValue("garage:requested-state", s.GarageRequestedState)
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

type SetHandler struct{}

func (h *SetHandler) OnSet(c qmq.WebServiceContext, key string, value proto.Message) {
	c.App().Logger().Advise(fmt.Sprintf("Garage door requested state changed to: %v", value))
	// c.App().Producer("garage:command:exchange").Push(&qmq.QMQMqttMessage{
	// 	Topic: "garage/command",
	// })
}

func main() {
	os.Setenv("QMQ_ADDR", "localhost:6379")

	service := qmq.NewWebService()
	service.Initialize(new(Schema))
	service.App().AddConsumer("garage:sensor:queue").Initialize()
	service.App().AddProducer("garage:command:exchange").Initialize(10)
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
