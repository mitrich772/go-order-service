package web

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	mockcache "github.com/mitrich772/go-order-service/internal/cache/mocks"
	"github.com/mitrich772/go-order-service/internal/database"
)

func TestServer_GetOrder(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := mockcache.NewMockOrderStore(ctrl)

	srv := Server{Store: mockStore}

	// Позитивный случай
	mockStore.EXPECT().Get("123").Return(&database.Order{OrderUID: "123"}, nil)
	order, err := srv.GetOrder("123")
	if err != nil || order.OrderUID != "123" {
		t.Fatalf("expected order 123, got %v, err: %v", order, err)
	}

	// uid пустой
	_, err = srv.GetOrder("")
	if err == nil {
		t.Fatal("expected error for empty uid")
	}

	// заказ не найден
	mockStore.EXPECT().Get("999").Return(nil, fmt.Errorf("not found"))
	_, err = srv.GetOrder("999")
	if err == nil {
		t.Fatal("expected error for missing order")
	}
}
func TestOrderHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := mockcache.NewMockOrderStore(ctrl)
	srv := Server{Store: mockStore, Tpl: template.Must(template.New("index").Parse("ok"))}

	// успешный запрос
	mockStore.EXPECT().Get("123").Return(&database.Order{OrderUID: "123"}, nil)
	req := httptest.NewRequest("GET", "/order/123", nil)
	w := httptest.NewRecorder()
	srv.OrderHandler(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var order database.Order
	if err := json.NewDecoder(w.Body).Decode(&order); err != nil {
		t.Fatal(err)
	}
	if order.OrderUID != "123" {
		t.Fatalf("expected UID 123, got %s", order.OrderUID)
	}

	// заказ не найден
	mockStore.EXPECT().Get("999").Return(nil, fmt.Errorf("not found"))
	req = httptest.NewRequest("GET", "/order/999", nil)
	w = httptest.NewRecorder()
	srv.OrderHandler(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}

	// пустой UID
	req = httptest.NewRequest("GET", "/order/", nil)
	w = httptest.NewRecorder()
	srv.OrderHandler(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}
func TestIndexHandler(t *testing.T) {
	tpl := template.Must(template.New("index").Parse("<html>ok</html>"))
	srv := Server{Tpl: tpl}

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	srv.IndexHandler(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if body := w.Body.String(); body != "<html>ok</html>" {
		t.Fatalf("unexpected body: %s", body)
	}

	// шаблон не установлен
	srv2 := Server{}
	req = httptest.NewRequest("GET", "/", nil)
	w = httptest.NewRecorder()
	srv2.IndexHandler(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}
func TestWriteJSON(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{"foo": "bar"}
	writeJSON(w, data)
	if w.Header().Get("Content-Type") != "application/json" {
		t.Fatal("Content-Type not set to application/json")
	}

	var result map[string]string
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	if result["foo"] != "bar" {
		t.Fatalf("expected foo=bar, got %v", result)
	}
}
