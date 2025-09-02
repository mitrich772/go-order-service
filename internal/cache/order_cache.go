package cache

import (
	"github.com/mitrich772/go-order-service/internal/db"
	"sync"

	"gorm.io/gorm"
)

// Кэш заказов
type OrderCache struct {
	mu    sync.RWMutex
	store map[string]*db.Order
}

// Создать новый кэш
func NewOrderCache() *OrderCache {
	return &OrderCache{
		store: make(map[string]*db.Order),
	}
}

// Получить заказ из кэша
func (c *OrderCache) Get(uid string) (*db.Order, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	order, ok := c.store[uid]

	return order, ok
}

// Добавить заказ в кэш
func (c *OrderCache) Set(order *db.Order) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.store[order.OrderUID] = order
}

// Создать новый кэш c бд
func Init(gormDB *gorm.DB) *OrderCache {
	cache := NewOrderCache()
	orders, err := db.GetAllOrders(gormDB)
	if err != nil {
		panic(err)
	}
	for i := range orders {
		cache.Set(&orders[i])
	}
	return cache
}
