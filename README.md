# Go Order Service

Читает заказы из Kafka, валидирует и сохраняет в PostgreSQL, отдаёт данные через HTTP API + простую веб-страницу.

---

# Возможности

* Подписка на Kafka-топик `orders` и обработка JSON-заказов  
* Валидация заказов на логические и структурные ошибки  
* Сохранение в PostgreSQL через GORM с транзакциями и retry  
* Потокобезопасный LRU-кэш
* DLQ (`orders-dlq`) для заказов, не прошедших обработку,  
  включая случаи временной недоступности базы (с отметкой о возможности повторной обработки)  
* HTTP API `GET /order/{order_uid}`
* CLI для миграций: `cmd/migrate` (`up`, `down`, `step`)


---

# Запуск

## Через Docker

```bash
docker-compose up --build
# После поднятия:
# UI: http://localhost:3000
```
Отправить тестовый заказ (через Docker)
```
go run ./producer/producer.go -port 29092
```
## Без Docker
Запуск приложения
```bash
go run ./cmd/app
```
Отправить тестовый заказ (без Docker)
```
go run ./producer/producer.go -port 9092
```
# Миграции

Примеры:

```bash
# Применить все миграции
go run ./cmd/migrate -action up

# Откатить все
go run ./cmd/migrate -action down

# Шаги
go run ./cmd/migrate -action step -n 2
```
# Структура проекта

```
cmd/
 ├─ app/           # main: init, db, cache, kafka, web, graceful shutdown
 └─ migrate/       # CLI миграций (golang-migrate)
internal/
 ├─ database/      # модели, GormDatabase, retry, валидация
 ├─ cache/         # OrderStore интерфейс, DBStore, DBWithCacheStore, LRU
 ├─ kafka/         # consumer (segmentio/kafka-go), DLQ, обработка сообщений
 └─ web/           # HTTP handlers
producer/
 └─ producer.go    # генератор тестовых заказов (gofakeit)
static/            # статические файлы
migrations/        # SQL миграции
templates/         # HTML-шаблоны
```

---

# Архитектура и поток данных

```
Kafka (orders) 
  → Consumer (parse → validate → SaveOrder with retry) 
    → PostgreSQL (GORM)
    → Update LRU-cache (если включен)
    ↳ При ошибке → отправка в DLQ (orders-dlq, c заголовком ошибки)

HTTP запрос: GET /order/{uid} 
  → OrderStore.Get
    → Cache hit → вернуть
    → Cache miss → БД → вернуть + кэшировать
```

# Тестирование

```bash
go test ./...
```

