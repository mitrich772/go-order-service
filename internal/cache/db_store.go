package cache

import "github.com/mitrich772/go-order-service/internal/database"

// DBStore — простое хранилище заказов только в базе данных. Реализует OrderStore.
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