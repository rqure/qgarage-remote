package main

import (
	"os"

	qdb "github.com/rqure/qdb/src"
	"github.com/rqure/qlib/pkg/app"
	"github.com/rqure/qlib/pkg/app/workers"
	"github.com/rqure/qlib/pkg/data/store"
)

func getDatabaseAddress() string {
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
	db := store.NewWeb(store.WebConfig{
		Address: getDatabaseAddress(),
	})

	storeWorker := workers.NewStore(db)
	webServiceWorker := qdb.NewWebServiceWorker(getWebServiceAddress())
	leadershipWorker := workers.NewLeadership(db)
	schemaValidator := leadershipWorker.GetEntityFieldValidator()
	garageController := NewGarageController(db)
	ttsController := NewTTSController(db)
	garageStatusCalculator := NewGarageStatusCalculator(db)

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
	storeWorker.SchemaUpdated.Connect(garageController.OnSchemaUpdated)
	storeWorker.SchemaUpdated.Connect(ttsController.OnSchemaUpdated)
	storeWorker.SchemaUpdated.Connect(garageStatusCalculator.OnSchemaUpdated)

	leadershipWorker.BecameLeader().Connect(garageController.OnBecameLeader)
	leadershipWorker.BecameLeader().Connect(ttsController.OnBecameLeader)
	leadershipWorker.BecameLeader().Connect(garageStatusCalculator.OnBecameLeader)
	leadershipWorker.LosingLeadership().Connect(garageController.OnLostLeadership)
	leadershipWorker.LosingLeadership().Connect(ttsController.OnLostLeadership)
	leadershipWorker.LosingLeadership().Connect(garageStatusCalculator.OnLostLeadership)

	// Create a new application configuration
	config := qdb.ApplicationConfig{
		Name: "garage",
		Workers: []qdb.IWorker{
			storeWorker,
			webServiceWorker,
			leadershipWorker,
			garageController,
			ttsController,
			garageStatusCalculator,
		},
	}

	app := app.NewApplication(config)

	app.Execute()
}
