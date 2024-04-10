package main

import (
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	qmq "github.com/rqure/qmq/src"
)

type GarageProcessorConfig struct {
	PulseDurationProvider PulseDurationProvider
	TtsProvider           TtsProvider
}

type GarageProcessor struct {
	config         GarageProcessorConfig
	activeReminder atomic.Bool
}

func NewGarageProcessor(config GarageProcessorConfig) qmq.WebServiceCustomProcessor {
	if config.PulseDurationProvider == nil {
		config.PulseDurationProvider = NewEnvironmentPulseDurationProvider()
	}

	if config.TtsProvider == nil {
		config.TtsProvider = NewEnvironmentReminderProvider()
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

			if state == qmq.GarageDoorState_CLOSED && p.config.TtsProvider.GetGarageClosedMessage() != "DISABLE" {
				e.WithProducer("audio-player:tts:exchange").Push(p.config.TtsProvider.GetGarageClosedMessage())
			}

			if state == qmq.GarageDoorState_OPENED && p.config.TtsProvider.GetGarageOpenedMessage() != "DISABLE" {
				e.WithProducer("audio-player:tts:exchange").Push(p.config.TtsProvider.GetGarageOpenedMessage())
			}

			if state == qmq.GarageDoorState_OPENED && p.activeReminder.CompareAndSwap(false, true) {
				go func() {
					if p.config.TtsProvider.GetGarageOpenedReminderMessage() != "DISABLE" {
						return
					}

					<-time.After(p.config.TtsProvider.GetGarageOpenedReminderInterval())

					if w.WithSchema().Get("garage:state").(*qmq.GarageDoorState).Value != qmq.GarageDoorState_OPENED {
						return
					}

					e.WithProducer("audio-player:tts:exchange").Push(p.config.TtsProvider.GetGarageOpenedReminderMessage())
				}()
			}
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
