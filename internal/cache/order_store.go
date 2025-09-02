package cache

import (
	"github.com/mitrich772/go-order-service/internal/db"

	"gorm.io/gorm"
)

// OrderStore описывает интерфейс хранилища заказов (БД или БД+кэш)
type OrderStore interface {
	Save(order *db.Order) error
	Get(uid string) (*db.Order, error)
}

// DBStore — простое хранилище заказов только в базе данных
type DBStore struct {
	db *gorm.DB
}

func NewDBStore(db *gorm.DB) *DBStore {
	return &DBStore{
		db: db,
	}
}

func (s *DBStore) Save(order *db.Order) error {
	return db.SaveOrder(s.db, order)
}

func (s *DBStore) Get(uid string) (*db.Order, error) {
	return db.GetOrder(s.db, uid)
}

// DBWithCacheStore — хранилище заказов с кэшем в памяти и базой данных
type DBWithCacheStore struct {
	db    *gorm.DB
	cache *OrderCache
}

func NewDBWithCacheStore(db *gorm.DB) *DBWithCacheStore {
	return &DBWithCacheStore{
		db:    db,
		cache: Init(db),
	}
}

func (s *DBWithCacheStore) Save(order *db.Order) error {
	if err := db.SaveOrder(s.db, order); err != nil {
		return err
	}
	if s.cache != nil {
		s.cache.Set(order)
	}
	return nil
}

func (s *DBWithCacheStore) Get(uid string) (*db.Order, error) {
	if s.cache != nil { // сначала пробуем кэш, не получилось? тогда пойдем в базу
		if order, ok := s.cache.Get(uid); ok {
			return order, nil
		}
	}

	order, err := db.GetOrder(s.db, uid)
	if err != nil {
		return nil, err
	}

	if s.cache != nil {
		s.cache.Set(order)
	}
	return order, nil
}
