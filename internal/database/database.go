package database

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
)

// ------------------- Структуры -------------------

// Config содержит настройки подключения к базе данных.
type Config struct {
	User     string
	Password string
	DBName   string
	SSLMode  string
	Host     string
	Port     string
}

// Order представляет заказ с вложенными Delivery, Payment и Items.
type Order struct {
	OrderUID          string    `gorm:"primaryKey;type:varchar(36)" json:"order_uid" validate:"required"`
	CustomerID        string    `gorm:"type:varchar(50);not null" json:"customer_id" validate:"required"`
	Locale            string    `gorm:"type:char(2)" json:"locale" validate:"required"`
	DeliveryService   string    `gorm:"type:varchar(50)" json:"delivery_service" validate:"required"`
	ShardKey          string    `gorm:"type:varchar(10)" json:"shardkey" validate:"required"`
	SmID              int16     `gorm:"type:smallint" json:"sm_id" validate:"gt=0"`
	DateCreated       time.Time `gorm:"type:timestamptz" json:"date_created" validate:"required,lte"`
	OofShard          string    `gorm:"type:varchar(10)" json:"oof_shard" validate:"required"`
	TrackNumber       string    `gorm:"type:varchar(50)" json:"track_number" validate:"required"`
	Entry             string    `gorm:"type:varchar(20)" json:"entry" validate:"required"`
	InternalSignature string    `gorm:"type:varchar(255)" json:"internal_signature" validate:"omitempty,max=255"`

	Delivery Delivery `gorm:"foreignKey:OrderUID;references:OrderUID" json:"delivery" validate:"required"`
	Payment  Payment  `gorm:"foreignKey:OrderUID;references:OrderUID" json:"payment" validate:"required"`
	Items    []Item   `gorm:"foreignKey:OrderUID;references:OrderUID" json:"items" validate:"required,min=1,dive"`
}

// Delivery содержит информацию о доставке заказа.
type Delivery struct {
	DeliveryID uint   `gorm:"primaryKey;autoIncrement;type:bigserial" json:"delivery_id"`
	OrderUID   string `gorm:"type:varchar(36);uniqueIndex" json:"order_uid"`
	Name       string `gorm:"type:varchar(100)" json:"name" validate:"required"`
	Phone      string `gorm:"type:varchar(20)" json:"phone" validate:"required,e164"`
	Zip        string `gorm:"type:varchar(20)" json:"zip" validate:"required"`
	City       string `gorm:"type:varchar(50)" json:"city" validate:"required"`
	Address    string `gorm:"type:varchar(200)" json:"address" validate:"required"`
	Region     string `gorm:"type:varchar(50)" json:"region" validate:"required"`
	Email      string `gorm:"type:varchar(100)" json:"email" validate:"required,email"`
}

// Payment содержит информацию о платеже заказа.
type Payment struct {
	PaymentID    uint    `gorm:"primaryKey;autoIncrement;type:bigserial" json:"payment_id"`
	OrderUID     string  `gorm:"type:varchar(36);uniqueIndex" json:"order_uid"`
	Transaction  string  `gorm:"type:varchar(36)" json:"transaction" validate:"required"`
	RequestID    string  `gorm:"type:varchar(50)" json:"request_id"`
	Currency     string  `gorm:"type:char(3)" json:"currency" validate:"required,len=3"`
	Provider     string  `gorm:"type:varchar(50)" json:"provider" validate:"required"`
	Amount       float64 `gorm:"type:numeric(12,2)" json:"amount" validate:"gte=0"`
	PaymentDT    int64   `gorm:"type:bigint" json:"payment_dt"`
	Bank         string  `gorm:"type:varchar(50)" json:"bank"`
	DeliveryCost float64 `gorm:"type:numeric(12,2)" json:"delivery_cost" validate:"gte=0"`
	GoodsTotal   float64 `gorm:"type:numeric(12,2)" json:"goods_total" validate:"gte=0"`
	CustomFee    float64 `gorm:"type:numeric(12,2)" json:"custom_fee" validate:"gte=0"`
}

// Item представляет товар в заказе.
type Item struct {
	ItemID      uint    `gorm:"primaryKey;autoIncrement;type:bigserial" json:"item_id"`
	OrderUID    string  `gorm:"type:varchar(36);index" json:"order_uid"`
	ChrtID      int64   `gorm:"type:bigint" json:"chrt_id" validate:"gt=0"`
	TrackNumber string  `gorm:"type:varchar(50)" json:"track_number" validate:"required"`
	Price       float64 `gorm:"type:numeric(12,2)" json:"price" validate:"gte=0"`
	RID         string  `gorm:"column:rid;type:varchar(36)" json:"rid"`
	Name        string  `gorm:"type:varchar(200)" json:"name" validate:"required"`
	Sale        float64 `gorm:"type:numeric(5,2)" json:"sale" validate:"gte=0"`
	Size        string  `gorm:"type:varchar(10)" json:"size"`
	TotalPrice  float64 `gorm:"type:numeric(12,2)" json:"total_price" validate:"gte=0"`
	NmID        int64   `gorm:"type:bigint" json:"nm_id" validate:"gt=0"`
	Brand       string  `gorm:"type:varchar(100)" json:"brand"`
	Status      int16   `gorm:"type:smallint" json:"status"`
}

// NewValidator создает новый валидатор с кастомными проверками
func NewValidator() *validator.Validate {
	v := validator.New()

	// Проверка дата не из будущего
	err := v.RegisterValidation("notfuture", func(fl validator.FieldLevel) bool {
		t, ok := fl.Field().Interface().(time.Time)
		return ok && !t.IsZero() && !t.After(time.Now())
	})
	if err != nil {
		log.Fatalf("failed to register validation: %v", err)
	}

	// Проверка согласованности track_number в items и заказе
	v.RegisterStructValidation(orderStructLevelValidation, Order{})
	// Проверка сумм: total(items) + delivery + custom_fee = payment.amount
	v.RegisterStructValidation(validateOrderTotal, Order{})

	return v
}

// У всех items должен совпадать track_number с заказом
func orderStructLevelValidation(sl validator.StructLevel) {
	order := sl.Current().Interface().(Order)
	for i, item := range order.Items {
		if item.TrackNumber != order.TrackNumber {
			sl.ReportError(item.TrackNumber, fmt.Sprintf("Items[%d].TrackNumber", i), "track_number", "trackmatch", order.TrackNumber)
		}
	}
}

// Суммы должны совпадать: все(item.TotalPrice) + payment.DeliveryCost + payment.CustomFee = Payment.Amount
func validateOrderTotal(sl validator.StructLevel) {
	order := sl.Current().Interface().(Order)

	sum := 0.0
	for _, item := range order.Items {
		sum += item.TotalPrice
	}
	total := sum + order.Payment.DeliveryCost + order.Payment.CustomFee
	diff := total - order.Payment.Amount
	if diff < -0.01 || diff > 0.01 {
		sl.ReportError(order.Payment.Amount, "Payment.Amount", "payment.amount", "totalmatch", fmt.Sprintf("%.2f", total))
	}
}

// ValidateOrder проверяет заказ на несоответствия
func ValidateOrder(order *Order) error {
	v := NewValidator()
	if err := v.Struct(order); err != nil {
		if ve, ok := err.(validator.ValidationErrors); ok {
			return formatValidationErrors(ve)
		}
		return err
	}
	return nil
}

// Проходит ошибки валидатора и выводит их
func formatValidationErrors(errs validator.ValidationErrors) error {
	var out []string
	for _, e := range errs {
		field := e.StructNamespace()
		switch e.Tag() {
		case "required":
			out = append(out, fmt.Sprintf("%s: обязательное поле", field))
		case "email":
			out = append(out, fmt.Sprintf("%s: некорректный email", field))
		case "e164":
			out = append(out, fmt.Sprintf("%s: телефон должен быть в формате +71234567890", field))
		case "gt", "gte":
			out = append(out, fmt.Sprintf("%s: значение должно быть >= 0", field))
		case "len":
			out = append(out, fmt.Sprintf("%s: длина должна быть %s", field, e.Param()))
		case "min":
			out = append(out, fmt.Sprintf("%s: минимум %s элементов", field, e.Param()))
		case "notfuture":
			out = append(out, fmt.Sprintf("%s: дата не может быть в будущем", field))
		case "trackmatch":
			out = append(out, fmt.Sprintf("%s: track_number не совпадает с заказом (%s)", field, e.Param()))
		default:
			out = append(out, fmt.Sprintf("%s: ошибка валидации (%s)", field, e.Tag()))
		}
	}
	return fmt.Errorf("%s", strings.Join(out, "; "))
}

// OrderFromJSON преобразует JSON-данные в структуру Order
func OrderFromJSON(data []byte) (*Order, error) {
	var order Order
	if err := json.Unmarshal(data, &order); err != nil {
		return nil, err
	}
	return &order, nil
}

// ------------------- Интерфейс БД -------------------

// Database описывает набор операций для работы с заказами.
type Database interface {
	// GetLastNOrders возвращает последние N заказов с подгруженными зависимостями
	GetLastNOrders(n int) ([]Order, error)
	// GetAllOrders возвращает все заказы с подгруженными зависимостями
	GetAllOrders() ([]Order, error)
	// GetOrder возвращает заказ по UID с подгруженными зависимостями
	GetOrder(uid string) (*Order, error)
	// SaveOrder сохраняет заказ и связанные данные в транзакции
	SaveOrder(order *Order) error
}
