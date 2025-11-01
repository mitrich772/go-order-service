package main

import (
	"context"
	"html/template"
	"log"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/mitrich772/go-order-service/internal/cache"
	"github.com/mitrich772/go-order-service/internal/database"
	"github.com/mitrich772/go-order-service/internal/kafka"
	"github.com/mitrich772/go-order-service/internal/web"
)

func main() {
	log.Println("Программа запущена")

	// Параметры для БД читаются из .env (если заданы), иначе используются значения по умолчанию.
	cfg := database.Config{
		User:     getenv("DB_USER", "serviceuser"),
		Password: getenv("DB_PASSWORD", "123"),
		DBName:   getenv("DB_NAME", "order_management"),
		SSLMode:  "disable",
		Host:     getenv("DB_HOST", "localhost"),
		Port:     getenv("DB_PORT", "5432"),
	}

	// --- Получаем соединение с gorm ---
	gorm := database.ConnectDB(cfg)
	defer database.Close(gorm)

	// --- Создаем обертку для работы с gorm ---
	database := database.NewGormDatabase(gorm, 3, 500*time.Millisecond)

	// --- Создание OrderStore ---
	var store cache.OrderStore
	storeCap, err := strconv.Atoi(getenv("CACHE_SIZE", "50"))
	if err != nil {
		log.Printf("Ошибка перевода storeCap %v", err)
		storeCap = 50
	}
	if getenv("ENABLE_CACHE", "true") == "true" {
		store = cache.NewDBWithCacheStore(database, storeCap)
	} else {
		store = cache.NewDBStore(database)
	}

	// --- Web ---
	tpl := template.Must(template.ParseFiles("templates/index.html"))
	webPort := getenv("PORT", "3000")

	web.Start(store, tpl, webPort)

	// --- Kafka ---
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	consumer := kafka.NewConsumer(
		store,
		[]string{getenv("KAFKA_BROKERS", "localhost:9092")},
		getenv("KAFKA_TOPIC", "orders"),
		getenv("KAFKA_GROUP", "order-service"),
		getenv("KAFKA_DLQ_TOPIC", "orders-dlq"),
	)
	consumer.Start(ctx)
	log.Printf("Kafka brokers %v\n", consumer.Brokers)
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

// getenv возвращает значение переменной окружения key
// fallback если переменная не установлена или пуста
func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
