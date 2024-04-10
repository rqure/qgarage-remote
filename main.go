package main

import (
	qmq "github.com/rqure/qmq/src"
	"google.golang.org/protobuf/proto"
)

type NameProvider struct{}

func (n *NameProvider) Get() string {
	return "garage"
}

type TransformerProviderFactory struct{}

func (t *TransformerProviderFactory) Create(components qmq.EngineComponentProvider) qmq.TransformerProvider {
	transformerProvider := qmq.NewDefaultTransformerProvider()
	transformerProvider.Set("consumer:garage:sensor:queue", []qmq.Transformer{
		qmq.NewMessageToAnyTransformer(components.WithLogger()),
		qmq.NewAnyToMqttTransformer(components.WithLogger()),
		NewMqttToDoorSensorTransformer(components.WithLogger()),
	})
	transformerProvider.Set("producer:garage:command:exchange", []qmq.Transformer{
		NewStateToDoorRelayTransformer(components.WithLogger()),
		NewDoorRelayToMqttTransformer(components.WithLogger()),
		qmq.NewMqttToAnyTransformer(components.WithLogger()),
		qmq.NewAnyToMessageTransformer(components.WithLogger()),
	})
	return transformerProvider
}

func main() {
	engine := qmq.NewDefaultEngine(qmq.DefaultEngineConfig{
		NameProvider:               &NameProvider{},
		TransformerProviderFactory: &TransformerProviderFactory{},
		EngineProcessor: qmq.NewWebServiceEngineProcessor(qmq.WebServiceEngineProcessorConfig{
			WebServiceCustomProcessor: NewGarageProcessor(GarageProcessorConfig{}),
			SchemaMapping: map[string]proto.Message{
				"garage:state":   &qmq.GarageDoorState{},
				"garage:trigger": &qmq.Int{},
			},
		}),
	})
	engine.Run()
}
