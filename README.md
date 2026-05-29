# url-shortener-ob

Сервис для сокращения URL-ссылок. Генерирует уникальный 10-символьный токен из `[a-zA-Z0-9_]` для каждого длинного URL. Поддерживает два типа хранилища: in-memory и PostgreSQL.

## Быстрый запуск

```bash
cp .env.example .env
make up
```

Сервер запустится на `:8080`. По умолчанию работает с in-memory хранилищем.

Чтобы изменить хранилище на PostgreSQL, нужно заменить в `.env`:

```env
STORAGE_TYPE=postgres

```

## API

### `POST /` — сократить ссылку

Принимает оригинальный URL в JSON-теле запроса.

```json
// Request
{
  "url": "https://example.com/very-long-url"
}

// Response 201 Created — создан новый токен
{
  "result": "aB3_xK9mZq"
}

// Response 200 OK — URL уже был сокращён ранее, возвращён существующий токен
{
  "result": "aB3_xK9mZq"
}

```

* **400 Bad Request** — неверный формат URL или синтаксическая ошибка в JSON.
* **507 Insufficient Storage** — достигнут лимит `MEMORY_STORAGE_LIMIT` (для in-memory).

### `GET /:short` — редирект на оригинальный URL

Принимает 10-символьный токен и выполняет временное перенаправление.

```http
// Запрос: GET /aB3_xK9mZq

// Response 307 Temporary Redirect
// Заголовок Location содержит оригинальный URL:
Location: https://example.com/very-long-url

```

* **400 Bad Request** — длина токена не равна 10 символам или содержит недопустимые символы.
* **404 Not Found** — токен не найден в системе.

## Команды Makefile

| Команда | Что делает |
| --- | --- |
| `make up` | Собрать и запустить проект в Docker |
| `make down` | Остановить Docker и удалить связанные volumes |
| `make build` | Собрать локальный бинарник |
| `make run` | Запустить сервис локально (без Docker) |
| `make test` | Запустить Unit-тесты с race detector |
| `make line` | Запустить линтер golangci-lint |

## Переменные окружения

Все настройки задаются через `.env` файл или переменные окружения.

```env
# Окружение: local, dev, prod
ENV=local

# Выбор хранилища: memory или postgres
STORAGE_TYPE=memory
COMPOSE_PROFILES=${STORAGE_TYPE}

# Лимит для memory хранилища (макс. количество ссылок)
MEMORY_STORAGE_LIMIT=500000

# Настройки HTTP-сервера
APP_PORT=8080
HTTP_TIMEOUT=4s
HTTP_IDLE_TIMEOUT=60s

# Настройки для подключения к PostgreSQL (требуются при STORAGE_TYPE=postgres)
DB_USER=username
DB_PASSWORD=password
DB_HOST=localhost
DB_PORT=5432
DB_NAME=shortener_db

```

### Описание переменных

| Переменная | По умолчанию | Обязательная | Описание |
| --- | --- | --- | --- |
| `ENV` | `local` | нет | Режим: `local`/`dev` (обычный лог), `prod` (JSON-лог) |
| `STORAGE_TYPE` | — | **да** | Хранилище: `memory` или `postgres` |
| `MEMORY_STORAGE_LIMIT` | `500000` | нет | Лимит емкости In-Memory хранилища |
| `APP_PORT` | `8080` | нет | Порт HTTP-сервера |
| `HTTP_TIMEOUT` | `4s` | нет | Таймаут чтения/записи HTTP-запроса |
| `HTTP_IDLE_TIMEOUT` | `60s` | нет | Таймаут простоя Keep-Alive соединений |
| `DB_USER` | — | нет* | Пользователь БД (обязательно при `STORAGE_TYPE=postgres`) |
| `DB_PASSWORD` | — | нет* | Пароль БД (обязательно при `STORAGE_TYPE=postgres`) |
| `DB_HOST` | `localhost` | нет | Хост PostgreSQL |
| `DB_PORT` | `5432` | нет | Порт PostgreSQL |
| `DB_NAME` | — | нет* | Имя базы данных (обязательно при `STORAGE_TYPE=postgres`) |

## Архитектура проекта


```
├── cmd/
│   └── shortener/
│       └── main.go             # Точка входа, инициализация конфигов, логгера и DI-контейнера
├── internal/
│   ├── config/
│   │   └── config.go           # Конфигурация приложения (чтение ENV/флагов)
│   ├── handler/
│   │   ├── handler.go          # Инициализация базового хендлера и роутинга (Gin)
│   │   ├── middleware/
│   │   │   └── logger.go       # Middleware для логирования HTTP-запросов (slog)
│   │   ├── url.go              # HTTP-хендлеры для POST/GET методов
│   │   └── url_test.go         # Unit-тесты для HTTP-слоя (httptest + Gin)
│   ├── service/
│   │   ├── errors.go           # Бизнес-ошибки сервисного слоя
│   │   ├── service.go          # Бизнес-логика сокращения ссылок и оркестрация коллизий
│   │   ├── service_test.go     # Unit-тесты для проверки генерации и логики повторов
│   │   └── token.go            # Алгоритм генерации уникальных токенов (Rejection Sampling)
│   └── repository/
│       ├── errors.go           # Общие ошибки для всех реализаций хранилищ
│       ├── memory/
│       │   ├── memory.go       # Самостоятельная потокобезопасность (In-Memory) + OOM защита
│       │   └── memory_test.go  # Нагрузочные тесты In-Memory хранилища (race detector)
│       └── postgres/
│           ├── db.go           # Инициализация и подключение к PostgreSQL
│           └── postgres.go     # Реализация хранилища для PostgreSQL (pgx/sql)
├── migrations/                 # SQL-миграции для базы данных (схема таблиц и индексы)
├── api/
│   └── openapi.yaml            # Спецификация API сервиса в формате OpenAPI/Swagger
├── Dockerfile                  # Контейнеризация Go-приложения (Multi-stage build)
├── docker-compose.yaml         # Инфраструктурный манифест для локального окружения
├── Makefile                    # Скрипты автоматизации (build, run, test, lint)
└── README.md                   # Документация проекта

```

### Ключевые особенности реализации:

* **Генерация токенов:** Токен генерируется криптографически стойким методом через `crypto/rand`. В случае коллизии на уровне хранилища сервис делает до 5 итераций перегенерации (`Retries`), прежде чем вернуть ошибку.
* **Идемпотентность POST:** На один и тот же оригинальный URL всегда возвращается один и тот же токен. Это гарантирует дедупликацию данных и защиту от бесконечного раздувания хранилища.

