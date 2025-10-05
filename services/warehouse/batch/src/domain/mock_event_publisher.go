package domain

import "log"

// MockBatchEventPublisher is a mock implementation of BatchEventPublisher for testing
type MockBatchEventPublisher struct {
	PublishedEvents []*BatchEvent
	ShouldFail      bool
	FailureError    error
}

// NewMockBatchEventPublisher creates a new mock event publisher
func NewMockBatchEventPublisher() *MockBatchEventPublisher {
	return &MockBatchEventPublisher{
		PublishedEvents: make([]*BatchEvent, 0),
		ShouldFail:      false,
	}
}

// PublishBatchEvent implements the BatchEventPublisher interface
func (m *MockBatchEventPublisher) PublishBatchEvent(event *BatchEvent) error {
	if m.ShouldFail {
		if m.FailureError != nil {
			return m.FailureError
		}
		return &MockPublishError{Message: "mock publish failure"}
	}
	
	m.PublishedEvents = append(m.PublishedEvents, event)
	log.Printf("Mock: Published batch event %s for batch %s", event.EventType, event.BatchID)
	return nil
}

// GetPublishedEvents returns all published events
func (m *MockBatchEventPublisher) GetPublishedEvents() []*BatchEvent {
	return m.PublishedEvents
}

// GetEventCount returns the number of published events
func (m *MockBatchEventPublisher) GetEventCount() int {
	return len(m.PublishedEvents)
}

// GetEventsByType returns events of a specific type
func (m *MockBatchEventPublisher) GetEventsByType(eventType BatchEventType) []*BatchEvent {
	var events []*BatchEvent
	for _, event := range m.PublishedEvents {
		if event.EventType == eventType {
			events = append(events, event)
		}
	}
	return events
}

// Reset clears all published events
func (m *MockBatchEventPublisher) Reset() {
	m.PublishedEvents = make([]*BatchEvent, 0)
	m.ShouldFail = false
	m.FailureError = nil
}

// SetShouldFail configures the mock to fail on next publish
func (m *MockBatchEventPublisher) SetShouldFail(shouldFail bool, err error) {
	m.ShouldFail = shouldFail
	m.FailureError = err
}

// MockPublishError represents a mock publish error
type MockPublishError struct {
	Message string
}

func (e *MockPublishError) Error() string {
	return e.Message
}