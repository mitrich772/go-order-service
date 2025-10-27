package database

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
)

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
