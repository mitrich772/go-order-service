# ----------------------
# Stage 1: Build
# ----------------------
FROM golang:1.25-alpine AS builder

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем файлы модуля и загружаем зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь код приложения
COPY . .

# Собираем бинарник
RUN go build -o app ./cmd/app

# ----------------------
# Stage 2: Runtime
# ----------------------
FROM alpine:latest

# Рабочая директория для контейнера
WORKDIR /root/

# Копируем бинарник из билд-стейджа
COPY --from=builder /app/app .

# Копируем шаблоны и статику
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/static ./static

# Открываем порт веб-сервера
EXPOSE 3000

# Запуск приложения
CMD ["./app"]
