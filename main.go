package main

import (
	"os"

	"github.com/rqure/qlib/pkg/app"
	"github.com/rqure/qlib/pkg/app/workers"
	"github.com/rqure/qlib/pkg/data/store"
)

func getStoreAddress() string {
	addr := os.Getenv("Q_ADDR")
	if addr == "" {
		addr = "ws://webgateway:20000/ws"
	}

	return addr
}

func getWebServiceAddress() string {
	addr := os.Getenv("Q_WEB_ADDR")
	if addr == "" {
		addr = "0.0.0.0:20001"
	}

	return addr
}

func main() {
	s := store.NewWeb(store.WebConfig{
		Address: getStoreAddress(),
	})

	storeWorker := workers.NewStore(s)
	webServiceWorker := workers.NewWeb(getWebServiceAddress())
	leadershipWorker := workers.NewLeadership(s)
	schemaValidator := leadershipWorker.GetEntityFieldValidator()
	garageController := NewGarageController(s)
	ttsController := NewTTSController(s)
	garageStatusCalculator := NewGarageStatusCalculator(s)

	schemaValidator.RegisterEntityFields("GarageController",
		"OpenTTS", "CloseTTS", "OpenReminderTTS", "OpenReminderInterval")

	schemaValidator.RegisterEntityFields("GarageDoor",
		"IsClosed",
		"ToggleTrigger", "ToggleTriggerFn",
		"Closing", "Moving",
		"TimeToOpen", "TimeToClose",
		"PercentClosed")

	storeWorker.Connected.Connect(leadershipWorker.OnStoreConnected)
	storeWorker.Disconnected.Connect(leadershipWorker.OnStoreDisconnected)

	leadershipWorker.BecameLeader().Connect(garageController.OnBecameLeader)
	leadershipWorker.BecameLeader().Connect(ttsController.OnBecameLeader)
	leadershipWorker.BecameLeader().Connect(garageStatusCalculator.OnBecameLeader)
	leadershipWorker.LosingLeadership().Connect(garageController.OnLostLeadership)
	leadershipWorker.LosingLeadership().Connect(ttsController.OnLostLeadership)
	leadershipWorker.LosingLeadership().Connect(garageStatusCalculator.OnLostLeadership)

	a := app.NewApplication("garage")
	a.AddWorker(storeWorker)
	a.AddWorker(webServiceWorker)
	a.AddWorker(leadershipWorker)
	a.AddWorker(garageController)
	a.AddWorker(ttsController)
	a.AddWorker(garageStatusCalculator)
	a.Execute()
}
