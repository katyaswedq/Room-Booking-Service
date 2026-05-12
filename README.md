Выполнение тестового задания на Авито стажировку, март 2026.

# Room Booking Service

Сервис бронирования переговорок на Go.

## Что реализовано

### Обязательная часть

- `GET /_info`
- `POST /dummyLogin`
- `GET /rooms/list`
- `POST /rooms/create`
- `POST /rooms/{roomId}/schedule/create`
- `GET /rooms/{roomId}/slots/list`
- `POST /bookings/create`
- `GET /bookings/my`
- `POST /bookings/{bookingId}/cancel`
- `GET /bookings/list`

### Дополнительная часть

- `POST /register`
- `POST /login`
- опциональное создание `conferenceLink` при создании брони
- Swagger UI
- unit-тесты
- integration-тесты
- нагрузочное тестирование

---

## Стек:

- Go
- PostgreSQL
- Docker Compose
- Chi Router, pgx, JWT, go-transaction-manager

---

## Логика создания слотов

Слоты генерируются на 7 дней вперёд, такое решение выбрано по двум причинам:

1. В ТЗ указано, что в 99.9% случаев пользователи просматривают слоты в пределах ближайших 7 дней.
2. Бронирование происходит по `slotId`, поэтому слот должен существовать как отдельная сущность со стабильным UUID, сохранённым в базе данных.

Слоты создаются атомарно вместе с расписанием в рамках одной транзакции. Это гарантирует, что не возникнет состояния, при котором расписание уже создано, а слоты ещё нет.

## Логика создания архитектуры 

Проект разделён на несколько слоёв:

- `cmd/app` — точка входа, инициализация зависимостей и запуск сервера
- `handlers` — HTTP-слой: чтение запросов, вызов usecase, формирование ответов
- `usecase` — бизнес-логика и основные сценарии работы системы
- `domain` — доменные модели и интерфейсы репозиториев
- `infrastructure` — работа с БД, JWT, конфигом

Основная идея архитектуры — не смешивать уровни ответственности.  
Handlers не содержат бизнес-логику, usecase не зависит от HTTP, repo отвечает только за доступ к данным, слой domain не проникает в handlers

## Запуск проекта 

В `docker-compose` уже заданы все необходимые значения по умолчанию, поэтому дополнительная настройка не требуется. Такое решение выбрано для упрощения запуска и проверки проекта.  

- docker compose up -d --build 
либо 
- make up

## Остановить проект

- docker compose down
либо 
- make down

## Запустить все тесты 

- go test ./...
либо 
make test

## Запустить только интеграционные тесты

- go test ./internal/tests -v 
либо 
- make integration

## Покрытие бизнес-логики

- go test ./internal/usecase/... -coverpkg=./internal/usecase/... -coverprofile=unit.out
- go tool cover -func=unit.out
либо 
- make unit-cover

## Результаты прогона нагрузочного тестирования:

- total: 1.3365 s
- average: 12.4 ms
- slowest: 113.9 ms
- p95: 17.7 ms
- p99: 26.2 ms
- throughput: 3741 req/s
- errors: 0

### Swagger

После запуска: **http://localhost:8080/swagger/**