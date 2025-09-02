package main

import (
	"context"
	"html/template"
	"log"
	"os"
	"os/signal"

	"github.com/mitrich772/go-order-service/internal/cache"
	"github.com/mitrich772/go-order-service/internal/db"
	"github.com/mitrich772/go-order-service/internal/kafka"
	"github.com/mitrich772/go-order-service/internal/web"
)

func main() {
	log.Println("Программа запущена!")

	cfg := db.Config{
		User:     getenv("DB_USER", "serviceuser"),
		Password: getenv("DB_PASSWORD", "123"),
		DBName:   getenv("DB_NAME", "order_management"),
		SSLMode:  "disable",
		Host:     getenv("DB_HOST", "localhost"),
		Port:     getenv("DB_PORT", "5432"),
	}

	gormDB := db.Init(cfg)
	defer db.Close(gormDB)

	// --- Создание OrderStore ---
	var store cache.OrderStore
	if getenv("ENABLE_CACHE", "true") == "true" {
		store = cache.NewDBWithCacheStore(gormDB)
	} else {
		store = cache.NewDBStore(gormDB)
	}

	// --- Web ---
	tpl := template.Must(template.ParseFiles("templates/index.html"))
	web.Start(store, tpl) // передаем интерфейс OrderStore

	// --- Kafka ---
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	consumer := kafka.NewConsumer(
		store,
		[]string{getenv("KAFKA_BROKERS", "localhost:9092")},
		getenv("KAFKA_TOPIC", "orders"),
		getenv("KAFKA_GROUP", "order-service"),
	)
	consumer.Start(ctx)
	// --- Graceful shutdown ---
	waitForShutdown(cancel)
}

func waitForShutdown(cancel context.CancelFunc) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	log.Println("Останавливаем сервис...")
	cancel()
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
