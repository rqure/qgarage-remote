package main

import (
	"time"

	qdb "github.com/rqure/qdb/src"
)

type Aqara_LLKZMK12LM struct {
	id          string
	PulseLength int64
}

func (d *Aqara_LLKZMK12LM) New(e *qdb.Entity) IControlDevice {
	return &Aqara_LLKZMK12LM{
		id:          e.GetId(),
		PulseLength: e.GetField("PulseLength").PullInt(),
	}
}

func (d *Aqara_LLKZMK12LM) GetModel() string {
	return "AqaraLLKZMK12LM"
}

func (d *Aqara_LLKZMK12LM) Open(writeRequests chan *qdb.DatabaseRequest) {
	if d.PulseLength <= 0 {
		qdb.Warn("[Aqara_LLKZMK12LM::Open] PulseDuration is 0")
		return
	}

	if d.id == "" {
		qdb.Warn("[Aqara_LLKZMK12LM::Open] id is empty")
		return
	}

	go func() {
		writeRequests <- &qdb.DatabaseRequest{
			Id:    d.id,
			Field: "StateOnTrigger",
			Value: qdb.NewIntValue(0),
		}

		<-time.After(time.Duration(d.PulseLength) * time.Millisecond)

		writeRequests <- &qdb.DatabaseRequest{
			Id:    d.id,
			Field: "StateOffTrigger",
			Value: qdb.NewIntValue(0),
		}
	}()
}

func (d *Aqara_LLKZMK12LM) Close(writeRequests chan *qdb.DatabaseRequest) {
	if d.PulseLength <= 0 {
		qdb.Warn("[Aqara_LLKZMK12LM::Close] PulseDuration is 0")
		return
	}

	if d.id == "" {
		qdb.Warn("[Aqara_LLKZMK12LM::Close] id is empty")
		return
	}

	go func() {
		writeRequests <- &qdb.DatabaseRequest{
			Id:    d.id,
			Field: "StateOnTrigger",
			Value: qdb.NewIntValue(0),
		}

		<-time.After(time.Duration(d.PulseLength) * time.Millisecond)

		writeRequests <- &qdb.DatabaseRequest{
			Id:    d.id,
			Field: "StateOffTrigger",
			Value: qdb.NewIntValue(0),
		}
	}()
}
