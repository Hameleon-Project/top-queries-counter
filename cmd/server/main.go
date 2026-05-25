package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"top-queries-counter/internal/api"
	"top-queries-counter/internal/app"
	"top-queries-counter/internal/config"
	"top-queries-counter/internal/store"
)

func main() {
	log.Println("Starting search trends service...")

	cfg := config.Load()
	s := store.New()
	p := app.NewProcessor(s)

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	go func() {
		for range ticker.C {
			s.UpdateCache()
		}
	}()

	stopRabbit := make(chan struct{})
	go p.ListenRabbit(cfg.AMQPURL, "search_logs", stopRabbit)

	handler := api.NewHandler(s)
	srv := &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: *handler,
	}

	go func() {
		log.Printf("HTTP Server running on %s\n", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to listen: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	close(stopRabbit)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Service exiting gracefully")
}
