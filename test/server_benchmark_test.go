package test

import (
	"html/template"
	"testing"

	"github.com/mitrich772/go-order-service/internal/cache"
	"github.com/mitrich772/go-order-service/internal/database"
	"github.com/mitrich772/go-order-service/internal/web"
)

// BenchmarkGetOrder сравнивает время получения заказа
// с включённым кешем и без него.
func BenchmarkGetOrder(b *testing.B) {
	cfg := database.Config{
		User:     "serviceuser",
		Password: "123",
		DBName:   "order_management",
		SSLMode:  "disable",
		Host:     "localhost",
		Port:     "5432",
	}

	// Инициализация БД
	gorm := database.ConnectDB(cfg)
	defer database.Close(gorm)
	database := database.NewGormDatabase(gorm, 1, 0)

	// Инициализация шаблона
	tpl, err := template.ParseFiles("C:/Users/dima/Desktop/gool/templates/index.html")
	if err != nil {
		b.Fatalf("не удалось загрузить шаблон: %v", err)
	}

	uid := "b563feb7b2b84b6test"

	// --- Сервер с кешем ---
	storeWithCache := cache.NewDBWithCacheStore(database, 25)
	serverWithCache := &web.Server{
		Store: storeWithCache,
		Tpl:   tpl,
	}

	// --- Сервер без кеша ---
	storeWithoutCache := cache.NewDBStore(database)
	serverWithoutCache := &web.Server{
		Store: storeWithoutCache,
		Tpl:   tpl,
	}

	b.Run("with cache", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if _, err := serverWithCache.GetOrder(uid); err != nil {
				b.Fatalf("ошибка GetOrder с кешем: %v", err)
			}
		}
	})

	b.Run("without cache", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if _, err := serverWithoutCache.GetOrder(uid); err != nil {
				b.Fatalf("ошибка GetOrder без кеша: %v", err)
			}
		}
	})
}
