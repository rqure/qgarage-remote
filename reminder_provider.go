package main

import (
	"os"
	"strconv"
	"time"
)

type ReminderProvider interface {
	GetInterval() time.Duration
	GetMessage() string
}

type EnvironmentReminderProvider struct {
	interval int
	message  string
}

func (p *EnvironmentReminderProvider) GetInterval() time.Duration {
	return time.Duration(p.interval) * time.Minute
}

func (p *EnvironmentReminderProvider) GetMessage() string {
	return p.message
}

func NewEnvironmentReminderProvider() ReminderProvider {
	interval, err := strconv.Atoi(os.Getenv("REMINDER_INTERVAL_MINUTES"))
	if err != nil {
		interval = 5
	}

	message := os.Getenv("REMINDER_MESSAGE")
	if message == "" {
		message = "Reminder: Garage door is open"
	}

	return &EnvironmentReminderProvider{
		interval: interval,
		message:  message,
	}
}
