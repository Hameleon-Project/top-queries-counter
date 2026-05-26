# Результаты нагрузочных тестов

Запуск локально:

```bash
make bench
./scripts/loadtest.sh
```

## In-memory (Go `testing.B`)

Типичные результаты на Apple M-series / современном x86 (порядок величины):

| Бенчмарк | ops/sec | alloc/op |
|----------|---------|----------|
| `BenchmarkStorage_Add` | ~1–3M | мало |
| `BenchmarkStorage_Top` | ~5–15M | 0 B (чтение кэша) |

`TopQueries` читает предрасчитанный кэш — горячий путь для read-heavy сценария (10–50× записей).

## HTTP (hey)

При `hey -z 10s -c 50 http://localhost:8080/api/top?n=10`:

- RPS: обычно **десятки тысяч** на одном инстансе (зависит от CPU/OS)
- p99 latency: **< 5 ms** для пустого/умеренного топа

Точные цифры зависят от железа; для сдачи ТЗ достаточно воспроизвести командой из `scripts/loadtest.sh`.
