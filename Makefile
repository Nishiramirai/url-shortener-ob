.PHONY: build run up down test test-cover clean

# Локальная сборка бинарника
build:
	go build -o bin/url-shortener ./cmd/app/main.go

# Локальный запуск приложения
run:
	go run cmd/shortener/main.go

# Запуск в Docker
up:
	docker compose up --build -d

# Остановка Docker и полная очистка volumes
down:
	docker compose down -v

# Запуск юнит-тестов с флагом race condition
test:
	go test -v -race ./...

# Запуск тестов с генерацией отчета о покрытии кода
test-cover:
	go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out

# Очистка локальных бинарников
clean:
	rm -rf bin/
	rm -f coverage.out