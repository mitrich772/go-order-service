package db

import (
	"encoding/json"
	"fmt"
	"log"
	"net/mail"
	"regexp"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// ------------------- Структуры -------------------
type Config struct {
	User     string
	Password string
	DBName   string
	SSLMode  string
	Host     string
	Port     string
}

type Order struct {
	OrderUID          string    `gorm:"primaryKey;type:varchar(36)" json:"order_uid"`
	CustomerID        string    `gorm:"type:varchar(50);not null" json:"customer_id"`
	Locale            string    `gorm:"type:char(2)" json:"locale"`
	DeliveryService   string    `gorm:"type:varchar(50)" json:"delivery_service"`
	ShardKey          string    `gorm:"type:varchar(10)" json:"shardkey"`
	SmID              int16     `gorm:"type:smallint" json:"sm_id"`
	DateCreated       time.Time `gorm:"type:timestamptz" json:"date_created"`
	OofShard          string    `gorm:"type:varchar(10)" json:"oof_shard"`
	TrackNumber       string    `gorm:"type:varchar(50)" json:"track_number"`
	Entry             string    `gorm:"type:varchar(20)" json:"entry"`
	InternalSignature string    `gorm:"type:varchar(255)" json:"internal_signature"`

	Delivery Delivery `gorm:"foreignKey:OrderUID;references:OrderUID" json:"delivery"`
	Payment  Payment  `gorm:"foreignKey:OrderUID;references:OrderUID" json:"payment"`
	Items    []Item   `gorm:"foreignKey:OrderUID;references:OrderUID" json:"items"`
}

type Delivery struct {
	DeliveryID uint   `gorm:"primaryKey;autoIncrement;type:bigserial" json:"delivery_id"`
	OrderUID   string `gorm:"type:varchar(36);uniqueIndex" json:"order_uid"`
	Name       string `gorm:"type:varchar(100)" json:"name"`
	Phone      string `gorm:"type:varchar(20)" json:"phone"`
	Zip        string `gorm:"type:varchar(20)" json:"zip"`
	City       string `gorm:"type:varchar(50)" json:"city"`
	Address    string `gorm:"type:varchar(200)" json:"address"`
	Region     string `gorm:"type:varchar(50)" json:"region"`
	Email      string `gorm:"type:varchar(100)" json:"email"`
}

type Payment struct {
	PaymentID    uint    `gorm:"primaryKey;autoIncrement;type:bigserial" json:"payment_id"`
	OrderUID     string  `gorm:"type:varchar(36);uniqueIndex" json:"order_uid"`
	Transaction  string  `gorm:"type:varchar(36)" json:"transaction"`
	RequestID    string  `gorm:"type:varchar(50)" json:"request_id"`
	Currency     string  `gorm:"type:char(3)" json:"currency"`
	Provider     string  `gorm:"type:varchar(50)" json:"provider"`
	Amount       float64 `gorm:"type:numeric(12,2)" json:"amount"`
	PaymentDT    int64   `gorm:"type:bigint" json:"payment_dt"`
	Bank         string  `gorm:"type:varchar(50)" json:"bank"`
	DeliveryCost float64 `gorm:"type:numeric(12,2)" json:"delivery_cost"`
	GoodsTotal   float64 `gorm:"type:numeric(12,2)" json:"goods_total"`
	CustomFee    float64 `gorm:"type:numeric(12,2)" json:"custom_fee"`
}

type Item struct {
	ItemID      uint    `gorm:"primaryKey;autoIncrement;type:bigserial" json:"item_id"`
	OrderUID    string  `gorm:"type:varchar(36);index" json:"order_uid"`
	ChrtID      int64   `gorm:"type:bigint" json:"chrt_id"`
	TrackNumber string  `gorm:"type:varchar(50)" json:"track_number"`
	Price       float64 `gorm:"type:numeric(12,2)" json:"price"`
	RID         string  `gorm:"type:varchar(36)" json:"rid"`
	Name        string  `gorm:"type:varchar(200)" json:"name"`
	Sale        float64 `gorm:"type:numeric(5,2)" json:"sale"`
	Size        string  `gorm:"type:varchar(10)" json:"size"`
	TotalPrice  float64 `gorm:"type:numeric(12,2)" json:"total_price"`
	NmID        int64   `gorm:"type:bigint" json:"nm_id"`
	Brand       string  `gorm:"type:varchar(100)" json:"brand"`
	Status      int16   `gorm:"type:smallint" json:"status"`
}

// ------------------- Функции -------------------
func Init(cfg Config) *gorm.DB {
	dsn := fmt.Sprintf(
		"user=%s password=%s dbname=%s sslmode=%s host=%s port=%s",
		cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode, cfg.Host, cfg.Port,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	log.Println("База данных подключена")
	AutoMigrate(db)
	return db
}

func Close(gormDB *gorm.DB) {
	sqlDB, err := gormDB.DB()
	if err != nil {
		log.Println("Ошибка получения *sql.DB:", err)
		return
	}
	sqlDB.Close()
}

func AutoMigrate(db *gorm.DB) {
	err := db.AutoMigrate(&Order{}, &Delivery{}, &Payment{}, &Item{})
	if err != nil {
		log.Fatalf("Ошибка миграции: %v", err)
	}
	log.Println("Таблицы созданы")
}

func GetAllOrders(db *gorm.DB) ([]Order, error) {
	var orders []Order
	err := db.Preload("Delivery").
		Preload("Payment").
		Preload("Items").
		Find(&orders).Error
	if err != nil {
		return nil, err
	}
	return orders, nil
}

// GetOrder возвращает заказ по ID с подгруженными связями
func GetOrder(db *gorm.DB, uid string) (*Order, error) {
	var order Order
	err := db.Preload("Delivery").
		Preload("Payment").
		Preload("Items").
		Where("order_uid = ?", uid).
		First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// OrderFromJSON парсит JSON в структуру Order
func OrderFromJSON(data []byte) (*Order, error) {
	var order Order
	if err := json.Unmarshal(data, &order); err != nil {
		return nil, err
	}
	return &order, nil
}

// SaveOrder сохраняет заказ в транзакции
func SaveOrder(db *gorm.DB, order *Order) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(order).Error; err != nil {
			return err
		}
		return nil
	})
}

// ------------------ Валидация заказа ------------------

func ValidateOrder(order *Order) error {
	// Основные поля
	if err := validateBasicFields(order); err != nil {
		return err
	}

	// Доставка
	if err := validateDelivery(order.Delivery); err != nil {
		return err
	}

	// Оплата
	if err := validatePayment(order.Payment); err != nil {
		return err
	}

	// Товары
	if err := validateItems(order.Items, order.TrackNumber); err != nil {
		return err
	}

	// Проверка согласованности суммы
	if err := validateTotalAmount(order); err != nil {
		return err
	}

	return nil
}

// ------------------ Вспомогательные функции ------------------

func validateBasicFields(order *Order) error {
	if order.OrderUID == "" {
		return fmt.Errorf("orderUID пустой")
	}
	if order.CustomerID == "" {
		return fmt.Errorf("customerID пустой")
	}
	if order.Locale == "" {
		return fmt.Errorf("locale пустой")
	}
	if order.DeliveryService == "" {
		return fmt.Errorf("deliveryService пустой")
	}
	if order.SmID <= 0 {
		return fmt.Errorf("smID должно быть положительным")
	}
	if order.DateCreated.IsZero() {
		return fmt.Errorf("dateCreated пустая")
	}
	if order.DateCreated.After(time.Now()) {
		return fmt.Errorf("dateCreated в будущем")
	}
	return nil
}

func validateDelivery(d Delivery) error {
	if d.Name == "" {
		return fmt.Errorf("delivery.name пустое")
	}
	if d.Phone == "" {
		return fmt.Errorf("delivery.phone пустое")
	}
	if !isValidPhone(d.Phone) {
		return fmt.Errorf("delivery.phone некорректный формат")
	}
	if d.Zip == "" {
		return fmt.Errorf("delivery.zip пустой")
	}
	if d.City == "" {
		return fmt.Errorf("delivery.city пустой")
	}
	if d.Address == "" {
		return fmt.Errorf("delivery.address пустой")
	}
	if d.Region == "" {
		return fmt.Errorf("delivery.region пустой")
	}
	if d.Email == "" {
		return fmt.Errorf("delivery.email пустой")
	}
	if _, err := mail.ParseAddress(d.Email); err != nil {
		return fmt.Errorf("delivery.email некорректен: %v", err)
	}
	return nil
}

func validatePayment(p Payment) error {
	if p.Transaction == "" {
		return fmt.Errorf("payment.transaction пустой")
	}
	if p.Amount < 0 {
		return fmt.Errorf("payment.amount < 0")
	}
	if p.DeliveryCost < 0 {
		return fmt.Errorf("payment.deliveryCost < 0")
	}
	if p.GoodsTotal < 0 {
		return fmt.Errorf("payment.goodsTotal < 0")
	}
	if p.CustomFee < 0 {
		return fmt.Errorf("payment.customFee < 0")
	}
	if p.Currency == "" {
		return fmt.Errorf("payment.currency пустой")
	}
	return nil
}

func validateItems(items []Item, trackNumber string) error {
	if len(items) == 0 {
		return fmt.Errorf("items пустые")
	}
	for i, item := range items {
		if item.Price < 0 {
			return fmt.Errorf("items[%d].price < 0", i)
		}
		if item.TotalPrice < 0 {
			return fmt.Errorf("items[%d].totalPrice < 0", i)
		}
		if item.Sale < 0 {
			return fmt.Errorf("items[%d].sale < 0", i)
		}
		if item.ChrtID <= 0 {
			return fmt.Errorf("items[%d].chrtID некорректен", i)
		}
		if item.NmID <= 0 {
			return fmt.Errorf("items[%d].nmID некорректен", i)
		}
		if item.TrackNumber != trackNumber {
			return fmt.Errorf("items[%d].trackNumber не совпадает с заказом", i)
		}
	}
	return nil
}

func validateTotalAmount(order *Order) error {
	sum := 0.0
	for _, item := range order.Items {
		sum += item.TotalPrice
	}
	total := sum + order.Payment.DeliveryCost + order.Payment.CustomFee
	diff := total - order.Payment.Amount
	if diff < -0.01 || diff > 0.01 {
		return fmt.Errorf("payment.amount %f не совпадает с суммой товаров %f", order.Payment.Amount, total)
	}
	return nil
}

func isValidPhone(phone string) bool {
	re := regexp.MustCompile(`^\+?[0-9]{7,15}$`)
	return re.MatchString(phone)
}
