package kafka

import (
	"context"
	"log"
	"time"

	"github.com/mitrich772/go-order-service/internal/cache"
	"github.com/mitrich772/go-order-service/internal/database"

	"github.com/segmentio/kafka-go"
)

// Consumer представляет Kafka consumer, который читает сообщения и сохраняет их в OrderStore.
type Consumer struct {
	store   cache.OrderStore
	reader  *kafka.Reader
	Brokers []string
	Topic   string
	GroupID string
}

// NewConsumer создает нового Kafka consumer с заданными параметрами.
func NewConsumer(cacheStore cache.OrderStore, brokers []string, topic, groupID string) *Consumer {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		GroupID:        groupID,
		Topic:          topic,
		MinBytes:       10e3, // 10KB
		MaxBytes:       10e6, // 10MB
		CommitInterval: time.Second,
	})
	return &Consumer{
		store:   cacheStore,
		reader:  r,
		Brokers: brokers,
		Topic:   topic,
		GroupID: groupID,
	}
}

// Consume читает сообщения из Kafka и вызывает handler для каждого сообщения.
func (c *Consumer) Consume(ctx context.Context, handler func(key, value []byte)) error {
	for {
		m, err := c.reader.ReadMessage(ctx)
		if err != nil {
			return err
		}
		log.Printf("consumed message: key=%s value=%s", string(m.Key), string(m.Value))
		handler(m.Key, m.Value)
	}
}

// Close закрывает Kafka reader.
func (c *Consumer) Close() error {
	return c.reader.Close()
}

// Start запускает Kafka consumer в отдельной горутине.
func (c *Consumer) Start(ctx context.Context) {
	go func() {
		err := c.Consume(ctx, func(key, value []byte) {
			order, err := database.OrderFromJSON(value)
			if err != nil {
				log.Printf("Ошибка парсинга JSON: %v", err)
				return
			}

			if err := database.ValidateOrder(order); err != nil {
				log.Printf("Ошибка валидации заказа %s: %v", order.OrderUID, err)
				return
			}

			if err := c.store.Save(order); err != nil {
				log.Printf("Ошибка сохранения заказа %s: %v", order.OrderUID, err)
			}
		})

		if err != nil {
			log.Printf("Consumer error: %v", err)
		}
	}()
}
