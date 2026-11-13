# go-sharded-ttl-cache

[![CI](https://github.com/zakharauvital/go-sharded-ttl-cache/actions/workflows/ci.yml/badge.svg)](https://github.com/zakharauvital/go-sharded-ttl-cache/actions/workflows/ci.yml)

A thread-safe sharded in-memory TTL cache in Go: 32 shards guarded by `sync.RWMutex`, keys distributed with `hash/maphash`, background cleanup of expired entries.

## Design

- **Sharding.** A single mutex over one map becomes a bottleneck under concurrent load. The key space is split into 32 shards, each with its own `RWMutex`, so writers to different shards never block each other and readers share a lock.
- **Hashing.** `maphash` with a per-cache random seed: fast, allocation-free, and resistant to hash flooding.
- **Expiration.** `Get` never returns an expired entry, even if cleanup has not removed it yet. A background goroutine sweeps shards on a ticker; it locks one shard at a time, so the rest of the cache stays available during cleanup.
- **Lifecycle.** Cleanup is bound to a `context.Context` and stops when the context is cancelled — no leaked goroutines.

## Structure

```
main.go                       — usage example
internal/cache/cache.go       — cache implementation
internal/cache/cache_test.go  — tests and benchmarks (package cache_test, public API only)
.github/workflows/ci.yml      — gofmt, go vet, golangci-lint, tests with -race
```

## Usage

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

c, err := cache.New(ctx, time.Second)
if err != nil {
	log.Fatal(err)
}

c.Set("key", "value", 5*time.Second)

val, ok := c.Get("key")
```

- `New(ctx, interval)` creates a cache and starts background cleanup every `interval`. Cleanup stops when `ctx` is cancelled. A non-positive `interval` returns `ErrInvalidInterval`.
- `Set(key, val, ttl)` stores a value; writing an existing key replaces the value and restarts its TTL.
- `Get(key)` returns the value and `true` if the key exists and its TTL has not expired.

## Testing

Tests are table-driven and run on `testing/synctest` with a virtual clock: TTL and cleanup timing are checked deterministically, without real sleeps or flakiness. They cover expiration boundaries (exactly at TTL, at a cleanup tick), overwrites, cleanup, invalid intervals and concurrent access under the race detector. Statement coverage is 100%.

Benchmarks: parallel `Get`, `Set`, a 90/10 read-heavy mix, and `Get` during aggressive cleanup.

## Commands

Requires Go 1.27+ (version from `go.mod`).

```sh
make                # list commands
make run            # run the example
make test           # tests with -race
make bench          # benchmarks
make test-coverage  # coverage, report in coverage.html
make lint           # gofmt, go vet and golangci-lint
make fmt            # format
make tidy           # go mod tidy
```
