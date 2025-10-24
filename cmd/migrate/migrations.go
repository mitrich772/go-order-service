package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {

	action := flag.String("action", "up", "up | down | step")
	steps := flag.Int("n", 1, "Количество шагов для step")
	flag.Parse()

	user := getenv("DB_USER", "serviceuser")
	password := getenv("DB_PASSWORD", "123")
	host := getenv("DB_HOST", "localhost")
	port := getenv("DB_PORT", "5432")
	dbname := getenv("DB_NAME", "order_management")
	sslmode := getenv("DB_SSLMODE", "disable")

	dsn := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=%s",
		user, password, host, port, dbname, sslmode,
	)

	m, err := migrate.New("file://migrations", dsn)
	if err != nil {
		log.Fatalf("Ошибка создания мигратора: %v", err)
	}

	switch *action {
	case "up":
		err = m.Up()
	case "down":
		err = m.Down()
	case "step":
		err = m.Steps(*steps)
	default:
		log.Fatalf("Неизвестное действие: %s", *action)
	}

	if err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Ошибка миграции: %v", err)
	}

	log.Println("Миграция завершена")
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
