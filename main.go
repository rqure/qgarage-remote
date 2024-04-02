package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"strconv"
	"time"

	qmq "github.com/rqure/qmq/src"
	"google.golang.org/protobuf/proto"
)

type Schema struct {
	db *qmq.QMQConnection
	kv map[string]proto.Message
}

func NewSchema() *Schema {
	s := new(Schema)
	s.kv = make(map[string]proto.Message)

	keys := []string{
		"garage:state",
		"garage:requested-state",
	}

	for _, key := range keys {
		switch key {
		case "garage:state":
			s.kv[key] = new(qmq.QMQGarageDoorState)
		case "garage:requested-state":
			s.kv[key] = new(qmq.QMQGarageDoorState)
		}
	}

	return s
}

func (s *Schema) Get(key string) proto.Message {
	v := s.kv[key]

	if v != nil {
		s.db.GetValue(key, v)
	}

	return v
}

func (s *Schema) Set(key string, value proto.Message) {
	v := s.kv[key]
	if v != nil && reflect.TypeOf(v) != reflect.TypeOf(value) {
		return
	}

	s.kv[key] = value
	s.db.SetValue(key, value)
}

func (s *Schema) GetAllData(db *qmq.QMQConnection) {
	s.db = db

	for key := range s.kv {
		s.Get(key)
	}
}

func (s *Schema) SetAllData(db *qmq.QMQConnection) {
	for key := range s.kv {
		s.Set(key, s.kv[key])
	}
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
