package cache

import (
	"fmt"
	"sync"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/mitrich772/go-order-service/internal/database"
	mockdb "github.com/mitrich772/go-order-service/internal/database/mocks"
)

// Проверяет, что OrderCache корректно сохраняет и возвращает *database.Order
func TestOrderCache_SetAndGet(t *testing.T) {
	cache := NewOrderCache(10)
	order := &database.Order{OrderUID: "123"}

	exist := cache.Set(order)
	if exist {
		t.Error("ожидалось exist=false для нового элемента")
	}

	got, ok := cache.Get("123")
	if !ok {
		t.Fatal("ожидалось ok=true, но получено false")
	}
	if got == nil || got.OrderUID != "123" {
		t.Errorf("получен неверный заказ: %+v", got)
	}
}

// Проверяет потокобезопасность работы Set/Get
func TestOrderCache_ConcurrentAccess(t *testing.T) {
	cache := NewOrderCache(100)
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			uid := fmt.Sprintf("order-%d", i)
			cache.Set(&database.Order{OrderUID: uid})
			cache.Get(uid)
		}(i)
	}
	wg.Wait()
}

// Проверяет инициализацию OrderCache из базы данных
func TestNewOrderCaheFromDB(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mockdb.NewMockDatabase(ctrl)
	orders := []database.Order{
		{OrderUID: "order-1"},
		{OrderUID: "order-2"},
	}

	mockDB.EXPECT().
		GetLastNOrders(10).
		Return(orders, nil).
		Times(1)

	cache := NewOrderCaheFromDB(mockDB, 10)

	for _, o := range orders {
		if _, ok := cache.Get(o.OrderUID); !ok {
			t.Errorf("заказ %s не найден в кэше", o.OrderUID)
		}
	}
}
