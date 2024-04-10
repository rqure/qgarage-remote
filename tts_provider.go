package main

import (
	"os"
	"strconv"
	"time"
)

type TtsProvider interface {
	GetGarageOpenedReminderInterval() time.Duration
	GetGarageOpenedReminderMessage() string
	GetGarageOpenedMessage() string
	GetGarageClosedMessage() string
}

type EnvironmentReminderProvider struct {
	interval                    int
	garageOpenedReminderMessage string
	garageOpenedMessage         string
	garageClosedMessage         string
}

func (p *EnvironmentReminderProvider) GetGarageOpenedReminderInterval() time.Duration {
	return time.Duration(p.interval) * time.Minute
}

func (p *EnvironmentReminderProvider) GetGarageOpenedReminderMessage() string {
	return p.garageOpenedReminderMessage
}

func (p *EnvironmentReminderProvider) GetGarageOpenedMessage() string {
	return p.garageOpenedMessage
}

func (p *EnvironmentReminderProvider) GetGarageClosedMessage() string {
	return p.garageClosedMessage
}

func NewEnvironmentReminderProvider() TtsProvider {
	interval, err := strconv.Atoi(os.Getenv("GARAGE_OPENED_REMINDER_INTERVAL_MINUTES"))
	if err != nil {
		interval = 5
	}

	garageOpenedReminderMessage := os.Getenv("GARAGE_OPENED_REMINDER_MESSAGE")
	if garageOpenedReminderMessage == "" {
		garageOpenedReminderMessage = "Reminder: Garage door is open"
	}

	garageOpenedMessage := os.Getenv("GARAGE_OPENED_MESSAGE")
	if garageOpenedMessage == "" {
		garageOpenedMessage = "Garage door has opened"
	}

	garageClosedMessage := os.Getenv("GARAGE_CLOSED_MESSAGE")
	if garageClosedMessage == "" {
		garageClosedMessage = "Garage door has closed"
	}

	return &EnvironmentReminderProvider{
		interval:                    interval,
		garageOpenedReminderMessage: garageOpenedReminderMessage,
		garageOpenedMessage:         garageOpenedMessage,
		garageClosedMessage:         garageClosedMessage,
	}
}
