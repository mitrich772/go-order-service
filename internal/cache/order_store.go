package cache

import (
	"github.com/mitrich772/go-order-service/internal/database"
)

// OrderStore описывает интерфейс хранилища заказов (БД или БД+кэш).
type OrderStore interface {
	Save(order *database.Order) error
	Get(uid string) (*database.Order, error)
}

