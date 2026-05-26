package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"top-queries-counter/internal/antispam"
	"top-queries-counter/internal/api"
	"top-queries-counter/internal/app"
	"top-queries-counter/internal/config"
	"top-queries-counter/internal/store"
)

func main() {
	log.Println("Starting search trends service...")

	cfg := config.Load()
	s := store.New()
	guard := antispam.New(cfg.AntispamUserCooldown, cfg.AntispamMaxPerIPMin, cfg.AntispamMaxQueryMin)
	p := app.NewProcessor(s, guard)

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	go func() {
		for range ticker.C {
			s.UpdateCache()
		}
	}()

	stopRabbit := make(chan struct{})
	go p.ListenRabbit(cfg.AMQPURL, cfg.QueueName, stopRabbit)

	srv := &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: api.NewHandler(s, cfg.MaxTopN),
	}

	go func() {
		log.Printf("HTTP server listening on %s", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down...")

	close(stopRabbit)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("HTTP shutdown error: %v", err)
	}
	log.Println("Service stopped")
}
