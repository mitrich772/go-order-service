package generate

import (
	"time"

	"github.com/mitrich772/go-order-service/internal/database"

	"github.com/brianvoe/gofakeit/v6"
)

// MakeOrder создает валидный пример db.Order
func MakeOrder() database.Order {
	// Инициализация генератора случайных данных
	gofakeit.Seed(0)

	orderUID := gofakeit.UUID()
	trackNumber := "WB" + gofakeit.LetterN(3) + gofakeit.DigitN(5)
	locale := gofakeit.RandomString([]string{"en", "ru"})
	entry := "WB" + gofakeit.LetterN(2)

	// --- Генерируем список товаров
	itemCount := gofakeit.IntRange(1, 3)
	var items []database.Item
	var goodsTotal float64

	for i := 0; i < itemCount; i++ {
		price := gofakeit.Price(100, 1000)
		sale := gofakeit.Price(0, 100)
		total := price - sale
		if total < 0 {
			total = price // защита от отрицательных значений
		}

		item := database.Item{
			OrderUID:    orderUID,
			ChrtID:      int64(gofakeit.Number(1000000, 9999999)),
			TrackNumber: trackNumber,
			Price:       price,
			RID:         gofakeit.UUID(),
			Name:        gofakeit.ProductName(),
			Sale:        sale,
			Size:        gofakeit.RandomString([]string{"XS", "S", "M", "L", "XL"}),
			TotalPrice:  total,
			NmID:        int64(gofakeit.Number(1000000, 9999999)),
			Brand:       gofakeit.Company(),
			Status:      int16(gofakeit.Number(100, 300)),
		}
		items = append(items, item)
		goodsTotal += total
	}

	// --- Стоимость доставки и сборы ---
	deliveryCost := gofakeit.Price(100, 2000)
	customFee := gofakeit.Price(0, 200)
	amount := goodsTotal + deliveryCost + customFee // Чтобы стоимость совпала

	dateCreated := time.Now().UTC()

	// --- Формируем заказ ---
	return database.Order{
		OrderUID:          orderUID,
		TrackNumber:       trackNumber,
		Entry:             entry,
		Locale:            locale,
		InternalSignature: "",
		CustomerID:        gofakeit.UUID(),
		DeliveryService:   gofakeit.RandomString([]string{"meest", "dhl", "dpd", "fedex", "wbexpress"}),
		ShardKey:          gofakeit.DigitN(1),
		SmID:              int16(gofakeit.Number(1, 100)),
		DateCreated:       dateCreated,
		OofShard:          gofakeit.DigitN(1),

		Delivery: database.Delivery{
			OrderUID: orderUID,
			Name:     gofakeit.Name(),
			Phone:    generatePhone(),
			Zip:      gofakeit.Zip(),
			City:     gofakeit.City(),
			Address:  gofakeit.Address().Address,
			Region:   gofakeit.State(),
			Email:    gofakeit.Email(),
		},

		Payment: database.Payment{
			OrderUID:     orderUID,
			Transaction:  gofakeit.UUID(),
			RequestID:    "",
			Currency:     gofakeit.RandomString([]string{"USD", "EUR", "RUB"}),
			Provider:     "wbpay",
			Amount:       amount,
			PaymentDT:    dateCreated.Unix(),
			Bank:         gofakeit.Company(),
			DeliveryCost: deliveryCost,
			GoodsTotal:   goodsTotal,
			CustomFee:    customFee,
		},

		Items: items,
	}
}

func generatePhone() string {
	return "+7" + gofakeit.DigitN(10)
}
