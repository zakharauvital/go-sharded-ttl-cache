package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"
	"time"

	"cachettl/internal/cache"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	c, err := cache.New(ctx, 100*time.Millisecond)
	if err != nil {
		log.Fatal(err)
	}

	c.Set("short", "val1", 100*time.Millisecond)
	c.Set("long", "val2", time.Second)

	time.Sleep(500 * time.Millisecond)

	for _, key := range []string{"short", "long"} {
		val, ok := c.Get(key)
		log.Printf("cache: %s=%q found=%t", key, val, ok)
	}
}
