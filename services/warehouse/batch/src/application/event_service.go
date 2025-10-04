package application

import (
	"log"

	"github.com/MATI-MBIT/arqnewgen-medisupply-eda/simple-service/batch/src/domain"
)

// EventService handles business logic for events
type EventService struct{}

// NewEventService creates a new EventService
func NewEventService() *EventService {
	return &EventService{}
}

// HandleEvent processes the received event
func (s *EventService) HandleEvent(event domain.Event) error {
	// Business logic for processing the event
	log.Printf("Processing event: Key=%s, Value=%s, Timestamp=%s", 
		event.Key, event.Value, event.Timestamp.Format("2006-01-02 15:04:05"))
	
	// Add your business logic here
	// For example: validate, transform, persist, etc.
	
	return nil
}