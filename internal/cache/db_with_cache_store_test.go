package cache

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	mockcache "github.com/mitrich772/go-order-service/internal/cache/mocks"
	"github.com/mitrich772/go-order-service/internal/database"
	mockdb "github.com/mitrich772/go-order-service/internal/database/mocks"
)

// 1 Save должен записывать заказ в DB без ошибки и обновлять Cache
func TestDBWithCacheStore_Save_StoreAndCache(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mockdb.NewMockDatabase(ctrl)
	mockCache := mockcache.NewMockCache(ctrl)

	mockDB.EXPECT().GetLastNOrders(100).Return(nil, nil)

	store := NewDBWithCacheStore(mockDB, 100)
	store.cache = mockCache

	order := &database.Order{OrderUID: "test"}

	mockDB.EXPECT().SaveOrder(order).Return(nil).Times(1)
	mockCache.EXPECT().Set(order).Times(1)

	err := store.Save(order)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}
}

// 2 Если SaveOrder вернул ошибку → Cache.Set не вызывается, ошибка возвращается
func TestDBWithCacheStore_Save_DBError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mockdb.NewMockDatabase(ctrl)
	mockCache := mockcache.NewMockCache(ctrl)

	mockDB.EXPECT().GetLastNOrders(100).Return(nil, nil)

	store := NewDBWithCacheStore(mockDB, 100)
	store.cache = mockCache

	order := &database.Order{OrderUID: "test"}

	mockDB.EXPECT().SaveOrder(order).Return(errors.New("ошибка БД")).Times(1)
	mockCache.EXPECT().Set(order).Times(0)

	err := store.Save(order)
	if err == nil {
		t.Fatalf("ожидалась ошибка сохранения в БД, но получили nil")
	}
}

// 3 Get возвращает элемент из кэша, DB не вызывается
func TestDBWithCacheStore_Get_FromCache(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mockdb.NewMockDatabase(ctrl)
	mockCache := mockcache.NewMockCache(ctrl)

	mockDB.EXPECT().GetLastNOrders(100).Return(nil, nil)

	store := NewDBWithCacheStore(mockDB, 100)
	store.cache = mockCache

	orderUID := "test"
	expectedOrder := &database.Order{OrderUID: orderUID}

	mockCache.EXPECT().Get(orderUID).Return(expectedOrder, true).Times(1)

	order, err := store.Get(orderUID)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	if order != expectedOrder {
		t.Fatalf("ожидался заказ %v, получили %v", expectedOrder, order)
	}
}

// 4 Get не находит заказ, DB вызывается и находит заказ
func TestDBWithCacheStore_Get_FromDB(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mockdb.NewMockDatabase(ctrl)
	mockCache := mockcache.NewMockCache(ctrl)

	mockDB.EXPECT().GetLastNOrders(100).Return(nil, nil)

	store := NewDBWithCacheStore(mockDB, 100)
	store.cache = mockCache

	orderUID := "test"
	expectedOrder := &database.Order{OrderUID: orderUID}

	mockCache.EXPECT().Get(orderUID).Return(nil, false).Times(1)
	mockDB.EXPECT().GetOrder(orderUID).Return(expectedOrder, nil).Times(1)
	mockCache.EXPECT().Set(expectedOrder).Times(1)

	order, err := store.Get(orderUID)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	if order != expectedOrder {
		t.Fatalf("ожидался заказ %v, получили %v", expectedOrder, order)
	}
}

// 5 Get элемента нет ни в кэше, ни в DB → ошибка возвращается
func TestDBWithCacheStore_Get_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mockdb.NewMockDatabase(ctrl)
	mockCache := mockcache.NewMockCache(ctrl)

	mockDB.EXPECT().GetLastNOrders(100).Return(nil, nil)

	store := NewDBWithCacheStore(mockDB, 100)
	store.cache = mockCache

	orderUID := "missing"

	mockCache.EXPECT().Get(orderUID).Return(nil, false).Times(1)
	mockDB.EXPECT().GetOrder(orderUID).Return(nil, errors.New("не найдено")).Times(1)

	_, err := store.Get(orderUID)
	if err == nil {
		t.Fatalf("ожидалась ошибка для отсутствующего заказа, но получили nil")
	}
}

// 6 Save nil-заказ не паникует и возвращает ошибку
func TestDBWithCacheStore_Save_NilOrder(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mockdb.NewMockDatabase(ctrl)
	mockCache := mockcache.NewMockCache(ctrl)

	mockDB.EXPECT().GetLastNOrders(100).Return(nil, nil)

	store := NewDBWithCacheStore(mockDB, 100)
	store.cache = mockCache

	err := store.Save(nil)
	if err == nil {
		t.Fatalf("ожидалась ошибка при сохранении nil-заказа, но получили nil")
	}
}
