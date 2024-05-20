package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
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

	wg := sync.WaitGroup{}
	wg.Add(2)

	consumerProcessorCtx, consumerProcessorCancel := context.WithCancel(context.Background())
	schemaProcessorCtx, schemaProcessorCancel := context.WithCancel(context.Background())

	go func() {
		defer wg.Done()

		for {
			select {
			case <-consumerProcessorCtx.Done():
				return
			case consumable := <-e.WithConsumer("garage:status").Pop():
				consumable.Ack()
				sensor := consumable.Data().(*GarageDoorSensorJson)
				state := qmq.GarageDoorState_OPENED
				if sensor.Contact {
					state = qmq.GarageDoorState_CLOSED
				}

				if w.WithSchema().Get("garage:state").(*qmq.GarageDoorState).Value != state {
					w.WithSchema().Set("garage:state", &qmq.GarageDoorState{Value: state})

					if state == qmq.GarageDoorState_CLOSED && p.config.TtsProvider.GetGarageClosedMessage() != "DISABLE" {
						e.WithProducer("audio-player:cmd:play-tts").Push(p.config.TtsProvider.GetGarageClosedMessage())
					}

					if state == qmq.GarageDoorState_OPENED && p.config.TtsProvider.GetGarageOpenedMessage() != "DISABLE" {
						e.WithProducer("audio-player:cmd:play-tts").Push(p.config.TtsProvider.GetGarageOpenedMessage())
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

							e.WithProducer("audio-player:cmd:play-tts").Push(p.config.TtsProvider.GetGarageOpenedReminderMessage())
						}()
					}
				}
			}
		}
	}()

	go func() {
		defer wg.Done()

		for {
			select {
			case <-schemaProcessorCtx.Done():
				return
			case key := <-w.WithSchema().Ch():
				w.WithWebClientNotifier().NotifyAll([]string{key})

				switch key {
				case "garage:trigger":
					w.WithLogger().Advise("Garage door button pressed")

					e.WithProducer("garage:cmd:relay").Push("ON")
					<-time.After(time.Duration(p.config.PulseDurationProvider.Get()) * time.Millisecond)
					e.WithProducer("garage:cmd:relay").Push("OFF")
				}
			}
		}
	}()

	<-quit
	consumerProcessorCancel()
	schemaProcessorCancel()
	wg.Wait()
}
