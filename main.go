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

	c := cache.New(ctx, 100*time.Millisecond)

	c.Set("k0", "val1", 100*time.Millisecond)
	c.Set("k1", "val1", 1000*time.Millisecond)

	v1, ok := c.Get("k1")
	if !ok {
		log.Printf("k1 not exist\n")
	}
	log.Printf("cache: k1=%v\n", v1)

	time.Sleep(1000 * time.Millisecond)

	v1, ok = c.Get("k1")
	if !ok {
		log.Printf("k1 not exist\n")
	}
	log.Printf("cache: k1=%v\n", v1)
}
