# Top Queries Counter

Сервис «Популярные поисковые запросы» для виджета на главной маркетплейса: поток событий из **RabbitMQ**, подсчёт **Top-N** за **5 минут**, отдача по быстрому **HTTP/JSON** API.

## Быстрый старт

```bash
docker compose up --build -d
make publish
curl "http://localhost:8080/api/top?n=10"
```

- API: `http://localhost:8080`
- RabbitMQ UI: `http://localhost:15672` (guest/guest)

Локально: RabbitMQ на `localhost:5672`, затем `make run`.

## HTTP API

| Метод | Путь | Описание |
|--------|------|----------|
| GET | `/health` | Healthcheck |
| GET | `/api/top?n=10` | Top-N за 5 минут (`n` или `limit`, макс. `MAX_TOP_N`) |
| GET | `/api/stoplist` | Стоп-слова |
| POST | `/api/stoplist` | `{"word":"..."}` — добавить и убрать из топа |
| DELETE | `/api/stoplist` | `{"word":"..."}` — удалить из стоп-листа |
| GET | `/metrics` | Prometheus |

```bash
curl "http://localhost:8080/api/top?n=5"
curl -X POST http://localhost:8080/api/stoplist -H "Content-Type: application/json" -d '{"word":"casino"}'
curl http://localhost:8080/metrics
```

## Контракт сообщений в брокере

Очередь: `search_logs` (`QUEUE_NAME`).

```json
{
  "query": "iphone 15",
  "user_id": "user-42",
  "ip": "203.0.113.10",
  "timestamp": 1716729600
}
```

| Поле | Обязательность | Назначение |
|------|----------------|------------|
| `query` | да | Текст поиска (trim + lower) |
| `user_id` | да* | Идентификатор пользователя |
| `ip` | да* | Резервный ID (боты/парсеры) |
| `timestamp` | нет | Unix **секунды**; `0` = время приёма |

\* Нужен хотя бы один из `user_id` / `ip` — иначе событие отбрасывается (защита от анонимной накрутки).

**Поля для бизнес-логики:** `query` + время — для топа; `user_id`/`ip` — для антиспама; лишние поля не требуются.

## Архитектура

```
RabbitMQ → Processor → antispam.Guard → Storage (in-memory)
                              ↓ UpdateCache (1/сек)
                         sorted cache → GET /api/top
```

- **Окно 5 мин:** `map[query][]timestamp`, очистка по `boundary = now - 300`.
- **Top-N:** метод `Storage.TopQueries(n)` — чтение готового отсортированного кэша (`RLock`), O(limit).
- **Read-heavy:** запись в map; чтение API не пересчитывает весь топ.
- **Стоп-лист:** `map[string]struct{}` + HTTP; при добавлении слова — `PurgeQuery` и пересчёт кэша.
- **Аномалии** (`internal/antispam`):
  - cooldown `(user_id, query)` — 5 с;
  - лимит событий с одного IP в минуту (парсеры);
  - лимит событий на один `query` в минуту (всплеск/накрутка конкурентов).
- **Без managed-облака:** RabbitMQ и приложение в `docker-compose`.
- **Старт:** пустое состояние (как в ТЗ).



## Компромиссы и неточности ТЗ

| В ТЗ | Решение |
|------|---------|
| Нет формата сообщений | JSON-контракт выше |
| «Защита от ботов» без метрик | Три уровня в `antispam`, пороги через env |
| Точность «5 минут» | События по `timestamp`; кэш 1 с — допустимо для UI |
| Read >> write | Предрасчёт топа раз в секунду |
| gRPC или HTTP | Выбран HTTP/JSON (в ТЗ — на выбор) |
| Персистентность не указана | In-memory ради latency; после рестарта — пусто |

## Тесты и нагрузка

```bash
make test
make bench
make loadtest   # hey + go test -bench (см. docs/BENCHMARKS.md)
```

## Переменные окружения

| Переменная | По умолчанию |
|------------|----------------|
| `HTTP_ADDR` | `:8080` |
| `AMQP_URL` | `amqp://guest:guest@localhost:5672/` |
| `QUEUE_NAME` | `search_logs` |
| `ANTISPAM_USER_COOLDOWN_SEC` | `5` |
| `ANTISPAM_MAX_PER_IP_PER_MIN` | `60` |
| `ANTISPAM_MAX_QUERY_PER_MIN` | `500` |
| `MAX_TOP_N` | `100` |
