package cache

import (
	"github.com/mitrich772/go-order-service/internal/database"
)

// OrderStore описывает интерфейс хранилища заказов (БД или БД+кэш).
type OrderStore interface {
	Save(order *database.Order) error
	Get(uid string) (*database.Order, error)
}

// DBStore — простое хранилище заказов только в базе данных.
type DBStore struct {
	db database.Database
}

// NewDBStore создает новый DBStore с подключением к базе данных.
func NewDBStore(db database.Database) *DBStore {
	return &DBStore{
		db: db,
	}
}

// Save сохраняет заказ в базе данных.
func (s *DBStore) Save(order *database.Order) error {
	return s.db.SaveOrder(order)
}

// Get возвращает заказ из базы данных по uid.
func (s *DBStore) Get(uid string) (*database.Order, error) {
	return s.db.GetOrder(uid)
}

// DBWithCacheStore — хранилище заказов с кэшем в памяти и базой данных.
type DBWithCacheStore struct {
	db    database.Database
	cache Cache
}

// NewDBWithCacheStore создает новый DBWithCacheStore с указанной емкостью кэша.
func NewDBWithCacheStore(db database.Database, storeCap int) *DBWithCacheStore {
	return &DBWithCacheStore{
		db:    db,
		cache: Init(db, storeCap),
	}
}

// Save сохраняет заказ в базе данных и обновляет кэш.
func (s *DBWithCacheStore) Save(order *database.Order) error {
	if err := s.db.SaveOrder(order); err != nil {
		return err
	}
	if s.cache != nil {
		s.cache.Set(order)
	}
	return nil
}

// Get возвращает заказ из кэша, если он есть, иначе из базы данных, и обновляет кэш.
func (s *DBWithCacheStore) Get(uid string) (*database.Order, error) {
	if s.cache != nil { // сначала пробуем кэш
		if order, ok := s.cache.Get(uid); ok {
			return order, nil
		}
	}

	order, err := s.db.GetOrder(uid)
	if err != nil {
		return nil, err
	}

	if s.cache != nil {
		s.cache.Set(order)
	}
	return order, nil
}
