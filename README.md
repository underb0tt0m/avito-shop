# avito-shop

Backend-сервис интернет-магазина мерча с JWT-авторизацией, переводом монет между пользователями, покупкой товаров и встроенным мониторингом.

## Что умеет

* авторизация по `username/password` с выдачей JWT;
* автоматическое создание нового пользователя при первом входе;
* просмотр баланса, инвентаря и истории переводов;
* перевод монет между пользователями;
* покупка товаров за монеты;
* метрики Prometheus;
* дашборды через Grafana;
* unit- и integration-тесты на PostgreSQL.

## Архитектура

```text
handler -> service -> storage
```

### Основные слои

* `cmd/handler` — HTTP-обработчики и маршруты;
* `internal/service` — бизнес-логика;
* `internal/storage/postgres` — работа с PostgreSQL;
* `internal/api_middleware` — auth и stopwatch middleware;
* `internal/jwtmanager` — создание и парсинг JWT;
* `internal/hasher` — хеширование паролей;
* `internal/prometheus_metrics` — метрики и middleware;
* `internal/storage/postgres/migrations` — SQL-миграции.

## Стек

* **Go**
* **Chi**
* **PostgreSQL**
* **pgx / pgxpool**
* **JWT**
* **bcrypt**
* **Prometheus + Grafana**
* **Docker Compose**
* **gomock**
* **testcontainers-go**
* **golangci-lint**
* **GitHub Actions**

## Инфраструктура и качество кода

* конфигурация приложения вынесена в `cmd/config.yaml`: отдельно задаются сервисный и технический порты, параметры PostgreSQL, логирование, JWT и JSON-кодек;
* для локальной разработки и проверки проекта используются команды из `Makefile`: запуск через Docker Compose, быстрые тесты, полный прогон тестов и линтинг;
* в проекте настроен `golangci-lint` для статического анализа и проверки качества кода;
* настроен GitHub Actions workflow для автоматического запуска проверок при изменениях в репозитории.

## Быстрый запуск

### 1. Подготовить `.env`

```env
DB_USER=your_username
DB_PASSWORD=your_password
JWT_SECRET=your-super-secret-key
```

### 2. Запустить проект

```bash
make docker-compose-build-up
```

### Make-команды

```bash
make docker-compose-build-up   # сборка и запуск проекта в Docker
make docker-compose-up         # запуск без пересборки
make build                     # локальная сборка бинарника
make run                       # локальный запуск приложения
make test-fast                 # быстрые тесты
make test                      # полный прогон тестов + coverage
make test-ci                   # тесты для CI
make lint                      # запуск golangci-lint
make clean                     # очистка бинарников и coverage-файлов
```

После запуска будут доступны:

* API: `http://localhost:8080/api`
* Метрики: `http://localhost:9200/metrics`
* Prometheus: `http://localhost:9090`
* Grafana: `http://localhost:3000`

Grafana по умолчанию:

* login: `admin`
* password: `admin`

## Основные эндпоинты

### `POST /api/auth`

Аутентификация пользователя и получение JWT-токена.

Если пользователя еще нет в базе данных, он создается автоматически при первом успешном запросе.

**Request body**
```json
{
  "username": "username",
  "password": "password"
}
```

**Response**

```json
{
  "token": "..."
}
```

---

### `GET /api/info`

Получение информации о пользователе:

* текущий баланс;
* инвентарь;
* история входящих и исходящих переводов.

**Headers**

```http
Authorization: Bearer <token>
```

---

### `POST /api/sendCoin`

Перевод монет другому пользователю.

**Headers**

```http
Authorization: Bearer <token>
```

**Request body**

```json
{
  "toUser": "username",
  "amount": 50
}
```

---

### `POST /api/buy/{itemID}`

Покупка товара по его идентификатору.

**Headers**

```http
Authorization: Bearer <token>
```

**Path parameters**

* `itemID` — идентификатор товара.

**Example**

```http
POST /api/buy/1
Authorization: Bearer <token>
```

> Примечание: в исходном контракте тестового задания покупка была описана через `GET /api/buy/{item}` с передачей имени товара в path-параметре. В этом проекте контракт был осознанно изменен на `POST /api/buy/{itemID}`, так как покупка изменяет состояние системы, а использование числового идентификатора товара делает API более устойчивым и однозначным.

## Тесты

### Запуск быстрых тестов

```bash
make test-fast
```

### Запуск всех тестов (в том числе интеграционных)

```bash
make test
```

Для integration tests используется `testcontainers-go`: поднимается реальный PostgreSQL-контейнер, применяются миграции и затем гоняются тесты storage-слоя.