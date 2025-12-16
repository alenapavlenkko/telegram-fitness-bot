# ---- Build stage ----
FROM golang:1.25-alpine AS build

# Устанавливаем зависимости
RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Копируем файлы проекта
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Сборка бота
RUN go build -o /bin/fitlife ./cmd/bot

# ---- Runtime stage ----
FROM alpine:3.18

# Устанавливаем сертификаты
RUN apk add --no-cache ca-certificates

# Копируем скомпилированный бинарник
COPY --from=build /bin/fitlife /bin/fitlife

# Копируем .env
COPY .env /app/.env

# Запуск приложения
CMD ["/bin/fitlife"]
