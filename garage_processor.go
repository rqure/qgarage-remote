package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	qmq "github.com/rqure/qmq/src"
)

type GarageProcessorConfig struct {
	PulseDurationProvider PulseDurationProvider
}

type GarageProcessor struct {
	config GarageProcessorConfig
}

func NewGarageProcessor(config GarageProcessorConfig) qmq.WebServiceCustomProcessor {
	if config.PulseDurationProvider == nil {
		config.PulseDurationProvider = NewEnvironmentPulseDurationProvider()
	}

	return &GarageProcessor{
		config: config,
	}
}

func (p *GarageProcessor) Process(e qmq.EngineComponentProvider, w qmq.WebServiceComponentProvider) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-quit:
			return
		case consumable := <-e.WithConsumer("garage:sensor:queue").Pop():
			consumable.Ack()
			sensor := consumable.Data().(*GarageDoorSensorJson)
			state := qmq.GarageDoorState_OPENED
			if sensor.Contact {
				state = qmq.GarageDoorState_CLOSED
			}
			w.WithSchema().Set("garage:state", &qmq.GarageDoorState{Value: state})
		case key := <-w.WithSchema().Ch():
			w.WithWebClientNotifier().NotifyAll([]string{key})

			switch key {
			case "garage:trigger":
				w.WithLogger().Advise("Garage door button pressed")

				e.WithProducer("garage:command:exchange").Push("ON")
				<-time.After(time.Duration(p.config.PulseDurationProvider.Get()) * time.Millisecond)
				e.WithProducer("garage:command:exchange").Push("OFF")
			}
		}
	}
}
