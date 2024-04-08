package main

import (
	"fmt"

	qmq "github.com/rqure/qmq/src"
)

type StateToDoorRelayTransformer struct {
	logger qmq.Logger
}

func NewStateToDoorRelayTransformer(logger qmq.Logger) qmq.Transformer {
	return &StateToDoorRelayTransformer{
		logger: logger,
	}
}

func (t *StateToDoorRelayTransformer) Transform(i interface{}) interface{} {
	s, ok := i.(string)
	if !ok {
		t.logger.Error(fmt.Sprintf("StateToDoorRelayTransformer.Transform: invalid input type %T", i))
		return nil
	}

	j := &GarageDoorRelayJson{
		State: s,
	}

	return j
}
