package main

import (
	"os"
	"os/signal"
	"strconv"
	"time"

	qmq "github.com/rqure/qmq/src"
)

type Schema struct {
	GarageState *qmq.QMQGarageDoorStateEnum `qmq:"garage:state"`
}

type TickHandler struct{}

func (h *TickHandler) onTick(c qmq.WebServiceContext) {
	schema := c.Schema().(Schema)

	mqttMessage := new(qmq.QMQMqttMessage)
	popped := c.App().Consumer("qmq2mqtt:exchange:zigbee2mqtt/garage-door-sensor").Pop(mqttMessage)
	if popped != nil {
		c.NotifyClients(qmq.DataUpdateResponse{
			Data: qmq.KeyValueResponse{
				Key:   "garage:state",
				Value: schema.GarageState.String(),
			},
		})
		popped.Ack()
	}
}

func main() {
	service := qmq.NewWebService()
	service.Initialize()
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
