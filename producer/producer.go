package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/mitrich772/go-order-service/producer/generate"
	"github.com/segmentio/kafka-go"
)

func main() {
	kafkaPort := flag.String("port", "9092", "Порт Kafka")
	flag.Parse()
	fmt.Printf("Порт для Kafka: %s\n", *kafkaPort)
	writer := &kafka.Writer{
		Addr:     kafka.TCP("localhost:" + *kafkaPort),
		Topic:    "orders",
		Balancer: &kafka.LeastBytes{},
	}
	defer writer.Close()

	order := generate.MakeOrder()
	//order.OrderUID = ""

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
	fmt.Printf("Id: %s\n", order.OrderUID)
	log.Println("Случайный заказ успешно отправлен в Kafka")
}
