````markdown
# Go Order Service

Сервис для чтения заказов из Kafka, сохранения в PostgreSQL и кэширования в памяти.

## Описание

Сервис выполняет:

- Подписку на Kafka-топик `orders` и получение сообщений с заказами.
- Сохранение заказов в PostgreSQL (`orders`, `items`, `payments`, `deliveries`).
- Кэширование заказов в памяти для быстрого доступа.
- Управление миграциями базы через CLI.

## Требования

- Go 1.21+
- PostgreSQL
- Kafka
- Docker
````

## Продюсер заказов (для теста)

### Локально
```bash
go run producer/producer.go -port 9092
```
### Если запускаете через Docker, producer заказов нужно запускать отдельно. 
```bash
go run producer/producer.go -port 29092
```
## Управление миграциями

```bash
# Применить все миграции
go run ./cmd/migrate -action up  

# Откатить все миграции
go run ./cmd/migrate -action down  

# Применить/откатить конкретное количество шагов
go run ./cmd/migrate -action step -n 2
```

## Запуск приложения

```bash
# Локально
go run cmd/app/main.go

## Запуск через Docker
docker-compose up --build
```
При локальном запуске сервис использует `localhost`.  
При запуске через docker использует то что в .env


## Структура проекта
```
cmd/
 ├─ app/           # Основное приложение
 └─ migrate/       # CLI для миграций базы
internal/
 ├─ database/      # Работа с PostgreSQL
 ├─ cache/         # Кэширование заказов
 ├─ kafka/         # Консьюмер Kafka
 └─ web/           # Web-сервер
producer/
 └─ producer.go    # Генератор случайных заказов для теста
migrations/        # SQL миграции для базы
templates/         # HTML-шаблоны для веб-интерфейса
```
