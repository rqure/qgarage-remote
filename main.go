package main

import (
	"os"

	qdb "github.com/rqure/qdb/src"
)

func getDatabaseAddress() string {
	addr := os.Getenv("QDB_ADDR")
	if addr == "" {
		addr = "redis:6379"
	}

	return addr
}

func main() {
	db := qdb.NewRedisDatabase(qdb.RedisDatabaseConfig{
		Address: getDatabaseAddress(),
	})

	dbWorker := qdb.NewDatabaseWorker(db)
	leaderElectionWorker := qdb.NewLeaderElectionWorker(db)
	schemaValidator := qdb.NewSchemaValidator(db)

	schemaValidator.AddEntity("Root", "SchemaUpdateTrigger")
	schemaValidator.AddEntity("GarageController", "OpenTTS", "CloseTTS", "OpenReminderTTS", "OpenReminderInterval")
	schemaValidator.AddEntity("GarageDoor", "GarageDoorStatus", "ControlUnit", "OpenTrigger", "CloseTrigger")

	dbWorker.Signals.SchemaUpdated.Connect(qdb.Slot(schemaValidator.OnSchemaUpdated))
	leaderElectionWorker.AddAvailabilityCriteria(func() bool {
		return schemaValidator.IsValid()
	})

	dbWorker.Signals.Connected.Connect(qdb.Slot(leaderElectionWorker.OnDatabaseConnected))
	dbWorker.Signals.Disconnected.Connect(qdb.Slot(leaderElectionWorker.OnDatabaseDisconnected))

	// Create a new application configuration
	config := qdb.ApplicationConfig{
		Name: "garage",
		Workers: []qdb.IWorker{
			dbWorker,
			leaderElectionWorker,
		},
	}

	// Create a new application
	app := qdb.NewApplication(config)

	// Execute the application
	app.Execute()
}
