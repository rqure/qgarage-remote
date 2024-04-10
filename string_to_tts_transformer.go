package main

import (
	"fmt"

	qmq "github.com/rqure/qmq/src"
)

type StringToTtsTransformer struct {
	logger qmq.Logger
}

func NewStringToTtsTransformer(logger qmq.Logger) qmq.Transformer {
	return &StringToTtsTransformer{
		logger: logger,
	}
}

func (t *StringToTtsTransformer) Transform(i interface{}) interface{} {
	s, ok := i.(string)
	if !ok {
		t.logger.Error(fmt.Sprintf("StringToTtsTransformer.Transform: invalid input type %T", i))
		return nil
	}

	m := &qmq.TextToSpeechRequest{
		Text: s,
	}

	return m
}
