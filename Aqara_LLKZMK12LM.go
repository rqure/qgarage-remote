package main

import (
	"time"

	qdb "github.com/rqure/qdb/src"
	"google.golang.org/protobuf/types/known/anypb"
)

type Aqara_LLKZMK12LM struct {
}

func (d *Aqara_LLKZMK12LM) GetModel() string {
	return "AqaraLLKZMK12LM"
}

func (d *Aqara_LLKZMK12LM) Open(controlDeviceEntityId string, writeRequests chan *qdb.DatabaseRequest) {
	a, err := anypb.New(&qdb.Int{Raw: 0})
	if err != nil {
		qdb.Warn("[Aqara_LLKZMK12LM::Open] Failed to create anypb: %s", err)
		return
	}

	go func() {
		writeRequests <- &qdb.DatabaseRequest{
			Id:    controlDeviceEntityId,
			Field: "StateOnTrigger",
			Value: a,
		}

		<-time.After(100 * time.Millisecond)

		writeRequests <- &qdb.DatabaseRequest{
			Id:    controlDeviceEntityId,
			Field: "StateOffTrigger",
			Value: a,
		}
	}()
}

func (d *Aqara_LLKZMK12LM) Close(controlDeviceEntityId string, writeRequests chan *qdb.DatabaseRequest) {
	a, err := anypb.New(&qdb.Int{Raw: 0})
	if err != nil {
		qdb.Warn("[Aqara_LLKZMK12LM::Close] Failed to create anypb: %s", err)
		return
	}

	go func() {
		writeRequests <- &qdb.DatabaseRequest{
			Id:    controlDeviceEntityId,
			Field: "StateOnTrigger",
			Value: a,
		}

		<-time.After(100 * time.Millisecond)

		writeRequests <- &qdb.DatabaseRequest{
			Id:    controlDeviceEntityId,
			Field: "StateOffTrigger",
			Value: a,
		}
	}()
}
