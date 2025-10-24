package database

import "gorm.io/gorm"

// GormDatabase реализует интерфейс Database через gorm
type GormDatabase struct {
	db *gorm.DB
}

// NewGormDatabase создает новый GormDatabase с указанным подключением gorm
func NewGormDatabase(db *gorm.DB) *GormDatabase {
	return &GormDatabase{db: db}
}

// GetLastNOrders возвращает последние N заказов с подгруженными зависимостями
func (r *GormDatabase) GetLastNOrders(n int) ([]Order, error) {
	var orders []Order
	err := r.db.Preload("Delivery").
		Preload("Payment").
		Preload("Items").
		Limit(n).
		Find(&orders).Error
	return orders, err
}

// GetAllOrders возвращает все заказы с подгруженными зависимостями
func (r *GormDatabase) GetAllOrders() ([]Order, error) {
	var orders []Order
	err := r.db.Preload("Delivery").
		Preload("Payment").
		Preload("Items").
		Find(&orders).Error
	return orders, err
}

// GetOrder возвращает заказ по UID с подгруженными зависимостями
func (r *GormDatabase) GetOrder(uid string) (*Order, error) {
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
}

// SaveOrder сохраняет заказ и связанные данные в транзакции
func (r *GormDatabase) SaveOrder(order *Order) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		return tx.Create(order).Error
	})
}
