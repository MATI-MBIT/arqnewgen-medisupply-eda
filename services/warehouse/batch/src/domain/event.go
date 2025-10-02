package domain

import "time"

// Event represents a domain event
type Event struct {
	Key       string
	Value     string
	Timestamp time.Time
}

// EventHandler defines the contract for handling events
type EventHandler interface {
	HandleEvent(event Event) error
}