# ---- Build stage ----
FROM golang:1.25-alpine AS build

RUN apk add --no-cache git ca-certificates

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /bin/fitlife ./cmd/bot


# ---- Runtime stage ----
FROM alpine:3.18

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=build /bin/fitlife /bin/fitlife

# ВАЖНО: копируем фронтенд-сборку
COPY --from=build /app/frontend/dist /app/frontend/dist

COPY .env /app/.env

EXPOSE 8080

CMD ["/bin/fitlife"]