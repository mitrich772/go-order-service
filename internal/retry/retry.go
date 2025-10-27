package retry

import (
	"fmt"
	"log"
	"time"
)
// TemporaryErrorChecker — проверяет, является ли ошибка временной
type TemporaryErrorChecker func(error) bool

// Retry выполняет функцию fn несколько раз с задержкой, если ошибка временная
func Retry[T any](maxRetries int, delay time.Duration, check TemporaryErrorChecker, fn func() (T, error)) (T, error) {
	var lastErr error
	var result T

	for i := 0; i < maxRetries; i++ {
		result, lastErr = fn()
		if lastErr == nil {
			return result, nil
		}

		// Если ошибка не временная — возвращаем сразу
		if !check(lastErr) {
			return result, lastErr
		}

		log.Printf("⚠️ Retry попытка %d/%d не удалась: %v", i+1, maxRetries, lastErr)
		time.Sleep(delay)
	}

	return result, fmt.Errorf("операция не удалась после %d попыток: %w", maxRetries, lastErr)
}