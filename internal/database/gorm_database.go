package database

import (
	"errors"
	"time"

	"github.com/mitrich772/go-order-service/internal/retry"
	"gorm.io/gorm"
)

// Функция проверки временных ошибок
func isTemporaryGormError(err error) bool {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Это логическая ошибка — Retry не нужен
		return false
	}
	// Здесь можно добавить проверку на таймауты, разрывы соединения и т.п.
	return true
}
// GormDatabase реализует интерфейс Database через gorm
type GormDatabase struct {
	db *gorm.DB
}

// NewGormDatabase создает новый GormDatabase с указанным подключением gorm
func NewGormDatabase(db *gorm.DB) *GormDatabase {
	return &GormDatabase{db: db}
}

// GetLastNOrders возвращает последние N заказов с подгруженными зависимостями с Retry
func (r *GormDatabase) GetLastNOrders(n int) ([]Order, error) {
	return retry.Retry[[]Order](3, 500*time.Millisecond, isTemporaryGormError, func() ([]Order, error) {
		var orders []Order
		err := r.db.Preload("Delivery").
			Preload("Payment").
			Preload("Items").
			Limit(n).
			Find(&orders).Error
		return orders, err
	})
}

// GetAllOrders возвращает все заказы с подгруженными зависимостями с Retry
func (r *GormDatabase) GetAllOrders() ([]Order, error) {
	return retry.Retry[[]Order](3, 500*time.Millisecond, isTemporaryGormError, func() ([]Order, error) {
		var orders []Order
		err := r.db.Preload("Delivery").
			Preload("Payment").
			Preload("Items").
			Find(&orders).Error
		return orders, err
	})
}

// GetOrder возвращает заказ по UID с подгруженными зависимостями с Retry
func (r *GormDatabase) GetOrder(uid string) (*Order, error) {
	return retry.Retry[*Order](3, 500*time.Millisecond, isTemporaryGormError, func() (*Order, error) {
		var order Order
		err := r.db.Preload("Delivery").
			Preload("Payment").
			Preload("Items").
			Where("order_uid = ?", uid).
			First(&order).Error
		if err != nil {
			return nil, err
		}
		return &order, nil
	})
}

// SaveOrder сохраняет заказ и связанные данные в транзакции
func (r *GormDatabase) SaveOrder(order *Order) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		return tx.Create(order).Error
	})
}
