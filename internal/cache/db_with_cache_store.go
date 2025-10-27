package cache

import (
	"errors"

	"github.com/mitrich772/go-order-service/internal/database"
)

// DBWithCacheStore — хранилище заказов с кэшем в памяти и базой данных. Реализует OrderStore.
type DBWithCacheStore struct {
	db    database.Database
	cache Cache
}

// NewDBWithCacheStore создает новый DBWithCacheStore с указанной емкостью кэша.
func NewDBWithCacheStore(db database.Database, storeCap int) *DBWithCacheStore {
	return &DBWithCacheStore{
		db:    db,
		cache: NewOrderCaheFromDB(db, storeCap),
	}
}

// Save сохраняет заказ в базе данных и обновляет кэш.
func (s *DBWithCacheStore) Save(order *database.Order) error {
	if order == nil {
		return errors.New("order is nil")
	}
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

