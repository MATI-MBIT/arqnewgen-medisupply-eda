package drivenadapters

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

// KafkaProducer handles message production to Kafka
type KafkaProducer struct {
	writer *kafka.Writer
}

// NewKafkaProducer creates a new KafkaProducer
func NewKafkaProducer(brokerAddress, topic string) *KafkaProducer {
	writer := &kafka.Writer{
		Addr:  kafka.TCP(brokerAddress),
		Topic: topic,
	}

	return &KafkaProducer{
		writer: writer,
	}
}

// SendMessage sends a message to Kafka
func (p *KafkaProducer) SendMessage(ctx context.Context, key, value string) error {
	return p.writer.WriteMessages(ctx,
		kafka.Message{
			Key:   []byte(key),
			Value: []byte(value),
		},
	)
}

// StartProducing starts producing messages periodically (for demo purposes)
func (p *KafkaProducer) StartProducing(ctx context.Context) {
	log.Println("Starting Kafka producer...")
	
	for {
		select {
		case <-ctx.Done():
			log.Println("Kafka producer stopping...")
			p.Close()
			return
		case <-time.After(1 * time.Second):
			key := fmt.Sprintf("Key-%d", time.Now().Unix())
			value := fmt.Sprintf("Hello Kafka at %s", time.Now().Format(time.RFC3339))
			
			if err := p.SendMessage(ctx, key, value); err != nil {
				log.Printf("Error sending message: %v", err)
			} else {
				log.Println("Message sent successfully")
			}
		}
	}
}

// Close closes the Kafka writer
func (p *KafkaProducer) Close() error {
	if p.writer != nil {
		return p.writer.Close()
	}
	return nil
}