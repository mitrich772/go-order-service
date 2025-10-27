package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"time"

	"github.com/mitrich772/go-order-service/producer/generate"
	"github.com/segmentio/kafka-go"
)

func main() {
	kafkaPort := flag.String("port", "9092", "Порт Kafka")
	flag.Parse()

	writer := &kafka.Writer{
		Addr:     kafka.TCP("localhost:" + *kafkaPort),
		Topic:    "orders",
		Balancer: &kafka.LeastBytes{},
	}
	defer writer.Close()

	order := generate.MakeOrder()

	jsonData, err := json.Marshal(order)
	if err != nil {
		log.Fatal("Ошибка маршалинга JSON:", err)
	}

	err = writer.WriteMessages(context.Background(),
		kafka.Message{
			Key:   []byte(order.OrderUID),
			Value: jsonData,
			Time:  time.Now(),
		},
	)
	if err != nil {
		log.Fatal("Ошибка при отправке сообщения:", err)
	}

	log.Println("Случайный заказ успешно отправлен в Kafka")
}
