package web

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/mitrich772/go-order-service/internal/cache"
	"github.com/mitrich772/go-order-service/internal/database"
)

// Server структура для работы с endpoints
type Server struct {
	Store cache.OrderStore
	Tpl   *template.Template
}

// IndexHandler рендерит главную страницу (форма для ввода ID заказа)
func (s *Server) IndexHandler(w http.ResponseWriter, r *http.Request) {
	if s.Tpl == nil {
		http.Error(w, "template not set", http.StatusInternalServerError)
		return
	}
	err := s.Tpl.Execute(w, nil)
	if err != nil {
		log.Fatalln(err)
	}
}

// OrderHandler возвращает данные заказа по ID в формате JSON
func (s *Server) OrderHandler(w http.ResponseWriter, r *http.Request) {
	uid := strings.TrimPrefix(r.URL.Path, "/order/")
	order, err := s.GetOrder(uid)
	if err != nil {
		http.Error(w, "order not found", http.StatusNotFound)
		return
	}
	writeJSON(w, order)
}

// GetOrder ищет заказ в хранилище (кэш + БД)
func (s *Server) GetOrder(uid string) (*database.Order, error) {
	if uid == "" {
		return nil, fmt.Errorf("order_uid required")
	}
	order, err := s.Store.Get(uid)
	if err != nil {
		return nil, err
	}

	return order, nil
}

// writeJSON сериализует данные в JSON и пишет в ответ
func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(v)
	if err != nil {
		log.Fatalln(err)
	}
}

// Start запускает HTTP-сервер
func Start(cacheStore cache.OrderStore, tpl *template.Template, port string) {
	mux := http.NewServeMux()
	srv := &Server{
		Store: cacheStore,
		Tpl:   tpl,
	}

	mux.HandleFunc("/", srv.IndexHandler)
	mux.HandleFunc("/order/", srv.OrderHandler)

	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	go func() {
		log.Println("Web сервер запущен на порту", port)
		if err := http.ListenAndServe(":"+port, mux); err != nil {
			log.Fatal(err)
		}
	}()
}
