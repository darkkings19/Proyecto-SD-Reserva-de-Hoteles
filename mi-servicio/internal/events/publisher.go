package events

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
)

type Publisher interface {
	Publish(ctx context.Context, key string, event Event) error
	Close() error
}

type Config struct {
	Enabled          bool
	BootstrapServers string
	Topic            string
	ClientID         string
}

type KafkaPublisher struct {
	enabled bool
	topic   string
	writer  *kafka.Writer
}

func NewPublisher(cfg Config) Publisher {
	if !cfg.Enabled {
		log.Println("[Reservas][Kafka] Publicacion deshabilitada")
		return NoopPublisher{}
	}

	if cfg.BootstrapServers == "" || cfg.Topic == "" {
		log.Println("[Reservas][Kafka] Configuracion incompleta; publicacion deshabilitada")
		return NoopPublisher{}
	}

	return &KafkaPublisher{
		enabled: true,
		topic:   cfg.Topic,
		writer: &kafka.Writer{
			Addr:         kafka.TCP(splitBootstrapServers(cfg.BootstrapServers)...),
			Topic:        cfg.Topic,
			Balancer:     &kafka.Hash{},
			RequiredAcks: kafka.RequireOne,
			Async:        false,
			BatchTimeout: 10 * time.Millisecond,
			Transport: &kafka.Transport{
				ClientID: cfg.ClientID,
			},
		},
	}
}

func (p *KafkaPublisher) Publish(ctx context.Context, key string, event Event) error {
	if !p.enabled {
		return nil
	}

	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	if strings.TrimSpace(key) == "" {
		return errors.New("kafka message key is required")
	}

	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(key),
		Value: body,
		Headers: []kafka.Header{
			{Key: "event_type", Value: []byte(event.EventType)},
			{Key: "source_service", Value: []byte(event.SourceService)},
		},
	})
}

func (p *KafkaPublisher) Close() error {
	if p.writer == nil {
		return nil
	}
	return p.writer.Close()
}

type NoopPublisher struct{}

func (NoopPublisher) Publish(context.Context, string, Event) error {
	return nil
}

func (NoopPublisher) Close() error {
	return nil
}

func splitBootstrapServers(value string) []string {
	parts := strings.Split(value, ",")
	servers := make([]string, 0, len(parts))
	for _, part := range parts {
		server := strings.TrimSpace(part)
		if server != "" {
			servers = append(servers, server)
		}
	}
	return servers
}
