package kafka

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/mitrich772/go-order-service/internal/cache"
	"github.com/mitrich772/go-order-service/internal/database"

	"github.com/segmentio/kafka-go"
)

// Consumer представляет Kafka consumer, который читает сообщения и сохраняет их в OrderStore.
type Consumer struct {
	Store     cache.OrderStore
	reader    *kafka.Reader
	dlqWriter *kafka.Writer
	Brokers   []string
	Topic     string
	GroupID   string
}

// NewConsumer создает нового Kafka consumer с заданными параметрами.
func NewConsumer(cacheStore cache.OrderStore, brokers []string, topic, groupID, dlqTopic string) *Consumer {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		GroupID:        groupID,
		Topic:          topic,
		MinBytes:       10e3,
		MaxBytes:       10e6,
		CommitInterval: time.Second,
	})

	w := &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    dlqTopic,
		Balancer: &kafka.LeastBytes{},
	}

	return &Consumer{
		Store:     cacheStore,
		reader:    r,
		dlqWriter: w,
		Brokers:   brokers,
		Topic:     topic,
		GroupID:   groupID,
	}
}

// Consume читает сообщения из Kafka и вызывает handler для каждого сообщения.
func (c *Consumer) Consume(ctx context.Context, handler func(key, value []byte)) error {
	for {
		m, err := c.reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return err
		}
		log.Printf("consumed message: key=%s value=%s\n\n", string(m.Key), string(m.Value))
		handler(m.Key, m.Value)
	}
}

// Close закрывает Kafka reader.
func (c *Consumer) Close() error {
	return c.reader.Close()
}

// Функция для отправки сообщения в DLQ
// err — ошибка из-за которой сообщение не удалось обработать.
// retryable — флаг показывающий, можно ли повторно обработать сообщение.
func (c *Consumer) sendToDLQ(value []byte, err error, retryable bool) error {
	headers := []kafka.Header{
		{Key: "error.class", Value: []byte(fmt.Sprintf("%T", err))},
		{Key: "error.message", Value: []byte(err.Error())},
		{Key: "retryable", Value: []byte(fmt.Sprintf("%v", retryable))},
		{Key: "ts.failed", Value: []byte(fmt.Sprintf("%d", time.Now().UnixMilli()))},
	}

	errWrite := c.dlqWriter.WriteMessages(context.Background(), kafka.Message{
		Value:   value,
		Headers: headers,
	})

	if errWrite != nil {
		return fmt.Errorf("ошибка отправки в DLQ: %w", errWrite)
	}

	log.Printf("Сообщение успешно отправлено в DLQ")
	return nil
}

// HandleMessage обрабатывает одно сообщение из Kafka
func (c *Consumer) HandleMessage(value []byte) error {
	order, err := database.OrderFromJSON(value)
	if err != nil { // Не парсится
		return c.sendToDLQ(value, err, false)
	}

	if err := database.ValidateOrder(order); err != nil { // Не валидируется
		return c.sendToDLQ(value, err, false)
	}

	if err := c.Store.Save(order); err != nil { // Если retry в бд не пробьется
		return c.sendToDLQ(value, err, true)
	}

	return nil
}

// Start пытается запустить Kafka consumer в отдельной горутине
func (c *Consumer) Start(ctx context.Context) {
	go func() {
		for {
			err := c.Consume(ctx, func(_, value []byte) {
				if err := c.HandleMessage(value); err != nil {
					log.Printf("Ошибка обработки сообщения: %v", err)
				}
			})

			if err != nil {
				log.Printf("Consumer error: %v. Повтор через 10 секунд", err)
			}

			select {
			case <-ctx.Done():
				log.Println("Consumer ctx cancel")
				return
			case <-time.After(10 * time.Second):
			}
		}
	}()
}
