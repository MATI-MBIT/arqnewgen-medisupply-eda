package drivenadapters

import (
	"errors"
	"testing"
)

func TestIsUnknownTopicOrPartitionError(t *testing.T) {
	testCases := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "exact match with error code",
			err:      errors.New("[3] Unknown Topic Or Partition: the request is for a topic or partition that does not exist on this broker"),
			expected: true,
		},
		{
			name:     "lowercase version",
			err:      errors.New("[3] unknown topic or partition: the request is for a topic or partition that does not exist on this broker"),
			expected: true,
		},
		{
			name:     "UnknownTopicOrPartition format",
			err:      errors.New("kafka: UnknownTopicOrPartition"),
			expected: true,
		},
		{
			name:     "generic unknown topic message",
			err:      errors.New("unknown topic or partition"),
			expected: true,
		},
		{
			name:     "topic does not exist message",
			err:      errors.New("topic or partition that does not exist"),
			expected: true,
		},
		{
			name:     "different error",
			err:      errors.New("connection refused"),
			expected: false,
		},
		{
			name:     "different kafka error",
			err:      errors.New("[1] OffsetOutOfRange"),
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := isUnknownTopicOrPartitionError(tc.err)
			if result != tc.expected {
				t.Errorf("Expected %v, got %v for error: %v", tc.expected, result, tc.err)
			}
		})
	}
}

func TestBatchEventPublisherAdapterCreation(t *testing.T) {
	brokerAddress := "localhost:9092"
	topic := "test-topic"
	
	adapter := NewBatchEventPublisherAdapter(brokerAddress, topic)
	
	if adapter == nil {
		t.Fatal("Expected adapter to be created, got nil")
	}
	
	if adapter.topic != topic {
		t.Errorf("Expected topic %s, got %s", topic, adapter.topic)
	}
	
	if adapter.brokerAddress != brokerAddress {
		t.Errorf("Expected broker address %s, got %s", brokerAddress, adapter.brokerAddress)
	}
	
	if adapter.writer == nil {
		t.Error("Expected writer to be initialized")
	}
	
	// Clean up
	adapter.Close()
}