package cache

import (
	"sync"

	"github.com/mitrich772/go-order-service/internal/database"
)

// OrderCache реализует интерфейс Cache и хранит кэш заказов с потокобезопасным доступом.
type OrderCache struct {
	mu      sync.RWMutex
	storage *LRU
}

// NewOrderCache создает новый OrderCache с заданной вместимостью storeCap.
func NewOrderCache(storeCap int) *OrderCache {
	return &OrderCache{
		storage: NewLru(storeCap),
	}
}

// Get возвращает заказ из кэша по uid.
func (c *OrderCache) Get(uid string) (*database.Order, bool) {
	c.mu.RLock()
	value, ok := c.storage.Get(uid)
	c.mu.RUnlock()

	if !ok {
		return nil, ok
	}

	order, ok := value.(*database.Order)
	if !ok {
		return nil, false
	}

	return order, ok
}

// Set добавляет заказ в кэш или обновляет существующий.
func (c *OrderCache) Set(order *database.Order) (exist bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	res := c.storage.Set(order.OrderUID, order)
	return res
}

// NewOrderCaheFromDB инициализирует OrderCache с данными из базы.
func NewOrderCaheFromDB(db database.Database, storeCap int) *OrderCache {
	cache := NewOrderCache(storeCap)

	orders, err := db.GetLastNOrders(storeCap)
	if err != nil {
		panic(err)
	}

	for i := range orders {
		cache.Set(&orders[i])
	}
	return cache
}
