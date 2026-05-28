# url-shortener-ob

Сервис для сокращения URL-ссылок. Генерирует уникальный 10-символьный токен из `[a-zA-Z0-9_]` для каждого длинного URL. Поддерживает два типа хранилища: in-memory и PostgreSQL.

## Быстрый запуск

```bash
cp .env.example .env
make up
```

Сервер на `:8080`. По умолчанию работает с in-memory хранилищем.

Чтобы изменить хранилище на PostgreSQL, нужно заменить в `.env`:
```
STORAGE_TYPE=postgres
```


## API

### `POST /` — сократить ссылку

```json
// Request
{"url": "https://example.com/very-long-url"}

// Response 201 — создан новый токен
{"result": "aB3_xK9mZq"}

// Response 200 — URL уже был сокращён, вернули существующий токен
{"result": "aB3_xK9mZq"}
```

### `GET /:short` — получить оригинальный URL

```json
// GET /aB3_xK9mZq
// Response 200
{"original_url": "https://example.com/very-long-url"}

// Response 404
{"error": "url not found"}
```

## Команды Makefile

| Команда       | Что делает                                     |
|---------------|------------------------------------------------|
| `make up`     | Собрать и запустить в Docker                   |
| `make down`   | Остановить Docker и почистить volumes          |
| `make build`  | Собрать локальный бинарник                     |
| `make run`    | Запустить локально (без Docker)                |
| `make test`   | Запустить тесты с race detector                |
| `make lint`   | Запустить golangci-lint                        |

## Переменные окружения

Все настройки задаются через `.env` файл или переменные окружения. Вот полный список:

```
# Окружение: local, dev, prod
ENV=local

# Выбор хранилища: memory или postgres
STORAGE_TYPE=postgres
COMPOSE_PROFILES=${STORAGE_TYPE}

# Настройки HTTP-сервера
HTTP_ADDRESS=:8080
HTTP_TIMEOUT=4s
HTTP_IDLE_TIMEOUT=60s

# Настройки для подключения к PostgreSQL
DB_USER=username
DB_PASSWORD=password
DB_HOST=localhost
DB_PORT=5432
DB_NAME=shortener_db
```

### Описание переменных

| Переменная        | По умолчанию  | Обязательная | Описание                                              |
|-------------------|--------------|--------------|-------------------------------------------------------|
| `ENV`             | `local`      | нет          | Режим работы: `local`/`dev` (текстовый лог), `prod` (JSON-лог) |
| `STORAGE_TYPE`    | —            | **да**       | Хранилище: `memory` (в памяти) или `postgres`          |
| `HTTP_ADDRESS`    | `:8080`      | нет          | Адрес и порт HTTP-сервера                             |
| `HTTP_TIMEOUT`    | `4s`         | нет          | Таймаут чтения/записи запроса                         |
| `HTTP_IDLE_TIMEOUT` | `60s`      | нет          | Таймаут простоя соединения                            |
| `DB_USER`         | —            | **да**       | Пользователь PostgreSQL (если `STORAGE_TYPE=postgres`) |
| `DB_PASSWORD`     | —            | **да**       | Пароль PostgreSQL                                     |
| `DB_HOST`         | `localhost`  | нет          | Хост PostgreSQL                                       |
| `DB_PORT`         | `5432`       | нет          | Порт PostgreSQL                                       |
| `DB_NAME`         | —            | **да**       | Имя базы PostgreSQL                                   |

Самый простой способ начать — скопировать `.env.example`:

```bash
cp .env.example .env
```

Для in-memory режима достаточно указать только `STORAGE_TYPE=memory`, остальное не важно.

## Архитектура

```
cmd/shortener/main.go          — точка входа
internal/
├── config/config.go           — конфиг из .env / env vars
├── handler/handler.go         — HTTP-обработчики
├── handler/middleware/        — логирование запросов
├── service/service.go         — бизнес-логика
└── repository/
    ├── errors.go              — общие ошибки
    ├── memory/memory.go       — in-memory хранилище
    └── postgres/
        ├── db.go              — подключение и миграции
        └── postgres.go        — PostgreSQL хранилище
migrations/                    — SQL-миграции
api/openapi.yaml               — OpenAPI-спецификация
```

Токен генерируется через `crypto/rand`. При совпадении (коллизии) делает до 5 попыток перегенерации. На один оригинальный URL — всегда один токен (дедупликация на уровне хранилища).

