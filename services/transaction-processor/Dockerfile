# Этап 1: Сборка
FROM golang:1.26.1-alpine AS builder

WORKDIR /app

# Копируем ВЕСЬ проект (все сервисы, common и go.work) целиком
COPY . .

# Переходим в папку нужного сервиса
WORKDIR /app/services/transaction-processor

# Собираем бинарник. 
# Так как мы скопировали весь корень, go.work лежит в /app и автоматически склеит все модули!
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/transaction-processor-bin ./cmd/transaction-processor/main.go

# Этап 2: Финальный образ (весит всего ~15-20 МБ)
FROM alpine:3.19

WORKDIR /app

# Копируем только готовый бинарник
COPY --from=builder /app/transaction-processor-bin .

EXPOSE 8081

CMD ["./transaction-processor-bin"]