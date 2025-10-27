package cache

import (
	"github.com/mitrich772/go-order-service/internal/database"
)

// Cache описывает интерфейс кеша
type Cache interface {
	Get(uid string) (*database.Order, bool)
	Set(order *database.Order) (exist bool)
}


