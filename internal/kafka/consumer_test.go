package kafka

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	mock_cache "github.com/mitrich772/go-order-service/internal/cache/mocks"
	"github.com/mitrich772/go-order-service/internal/database"
	"github.com/mitrich772/go-order-service/producer/generate"
)

// Проверяет: валидное сообщение → парсится → проходит валидацию → Save вызывается 1 раз
func TestConsumer_HandleMessage_SavesOrder(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := mock_cache.NewMockOrderStore(ctrl)

	consumer := Consumer{
		Store: mockStore,
	}
	order := generate.MakeOrder()
	correctUID := order.OrderUID

	payload, err := json.Marshal(order)
	if err != nil {
		t.Fatalf("ошибка маршалинга: %v", err)
	}

	mockStore.EXPECT().
		Save(gomock.AssignableToTypeOf(&database.Order{})).
		DoAndReturn(func(o *database.Order) error {
			if o.OrderUID != correctUID {
				t.Fatalf("неверный UID: %s", o.OrderUID)
			}
			return nil
		}).
		Times(1)

	errH := consumer.HandleMessage(payload)
	if errH != nil {
		t.Fatalf("неожиданная ошибка: %v", errH)
	}
}

// Проверяет: JSON валидный синтаксически, но не проходит Validate → Save не вызывается
func TestConsumer_HandleInvalidJSON_NoSave(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := mock_cache.NewMockOrderStore(ctrl)

	consumer := Consumer{
		Store: mockStore,
	}

	payload := []byte(`{"order_uid":"123"}`)

	mockStore.EXPECT().
		Save(gomock.Any()).
		Times(0)

	err := consumer.HandleMessage(payload)
	if err == nil {
		t.Fatalf("ожидалась ошибка валидации, но получили nil")
	}
}

// Проверяет: JSON невалиден синтаксически → парсинг падает → Save не вызывается
func TestConsumer_HandleMalformedJSON_NoSave(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := mock_cache.NewMockOrderStore(ctrl)

	consumer := Consumer{
		Store: mockStore,
	}

	payload := []byte(`{"order_uid":123`) // пропущена скобка

	mockStore.EXPECT().
		Save(gomock.Any()).
		Times(0)

	err := consumer.HandleMessage(payload)
	if err == nil {
		t.Fatalf("ожидалась ошибка парсинга, но получили nil")
	}
}

// Проверяет: Save возвращает ошибку → метод не паникует → ошибка возвращается наружу
func TestConsumer_HandleMessage_SaveError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := mock_cache.NewMockOrderStore(ctrl)

	consumer := Consumer{
		Store: mockStore,
	}

	order := generate.MakeOrder()

	payload, err := json.Marshal(order)
	if err != nil {
		t.Fatalf("ошибка маршалинга: %v", err)
	}

	mockStore.EXPECT().
		Save(gomock.Any()).
		Return(errors.New("ошибка БД"))

	errH := consumer.HandleMessage(payload)
	if errH == nil {
		t.Fatalf("ожидалась ошибка сохранения, но получили nil")
	}
}
