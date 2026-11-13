# go-sharded-ttl-cache

Потокобезопасный шардированный in-memory кэш с TTL на Go: 32 шарда с RWMutex, распределение ключей через maphash, фоновая очистка просроченных записей.

## Структура

```
main.go                       — пример использования
internal/cache/cache.go       — реализация кэша
internal/cache/cache_test.go  — тесты и бенчмарки (package cache_test, только публичный API)
```

## Использование

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

c := cache.New(ctx, time.Second)

c.Set("key", "value", 5*time.Second)

val, ok := c.Get("key")
```

- `New(ctx, interval)` создаёт кэш и запускает фоновую очистку с периодом `interval`. Очистка останавливается при отмене `ctx`.
- `Set(key, val, ttl)` записывает значение; повторная запись перезаписывает значение и отсчитывает TTL заново.
- `Get(key)` возвращает значение и `true`, если ключ есть и его TTL не истёк. Просроченная запись не отдаётся, даже если очистка ещё не успела её удалить.

## Запуск

```sh
go run .
```

## Тесты

Требуется Go 1.27+ (версия из `go.mod`).

```sh
go test -race ./...
go test -run=^$ -bench=. -benchmem ./...
```
