package main

import (
	"os"
	"strconv"
)

type PulseDurationProvider interface {
	Get() int
}

type EnvironmentPulseDurationProvider struct {
	pulseDuration int
}

func (p *EnvironmentPulseDurationProvider) Get() int {
	return p.pulseDuration
}

func NewEnvironmentPulseDurationProvider() PulseDurationProvider {
	pulseDuration, err := strconv.Atoi(os.Getenv("PULSE_DURATION_MILLIS"))
	if err != nil {
		pulseDuration = 200
	}

	return &EnvironmentPulseDurationProvider{
		pulseDuration: pulseDuration,
	}
}
