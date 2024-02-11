package main

import (
	"os"
	"os/signal"
	"strconv"
	"time"
)

func main() {
	service := NewWebService()
	service.Initialize()
	defer service.Deinitialize()

	tickRateMs, err := strconv.Atoi(os.Getenv("TICK_RATE_MS"))
	if err != nil {
		tickRateMs = 100
	}

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)

	ticker := time.NewTicker(time.Duration(tickRateMs) * time.Millisecond)
	for {
		select {
		case <-sigint:
			return
		case <-ticker.C:
			service.Tick()
		}
	}
}
