package database

import (
	"encoding/json"
	"time"
)

// ------------------- Интерфейс БД -------------------

// Database описывает набор операций для работы с заказами.
type Database interface {
	GetLastNOrders(n int) ([]Order, error)
	GetAllOrders() ([]Order, error)
	GetOrder(uid string) (*Order, error)
	SaveOrder(order *Order) error
}

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

// OrderFromJSON преобразует JSON-данные в структуру Order
func OrderFromJSON(data []byte) (*Order, error) {
	var order Order
	if err := json.Unmarshal(data, &order); err != nil {
		return nil, err
	}
	return &order, nil
}
