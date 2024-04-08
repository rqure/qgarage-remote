package main

// Example JSON:
// {"battery":100,"contact":false,"device_temperature":25,"linkquality":87,"power_outage_count":5,"voltage":3085}
type GarageDoorSensorJson struct {
	Battery           int     `json:"battery"`
	Contact           bool    `json:"contact"`
	DeviceTemperature float32 `json:"device_temperature"`
	LinkQuality       int     `json:"linkquality"`
	PowerOutageCount  int     `json:"power_outage_count"`
	Voltage           int     `json:"voltage"`
}
