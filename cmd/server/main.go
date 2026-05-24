package main

import (
	"log"
	"time"

	"top-queries-counter/internal/app"
	"top-queries-counter/internal/config"
	"top-queries-counter/internal/store"
)

func main() {
	log.Println("Starting search trends service...")

	_ = config.Load()

	s := store.New()
	p := app.NewProcessor(s)

	p.Process(app.SearchEvent{Query: "iphone", UserID: "user1", Timestamp: time.Now().Unix()})
	p.Process(app.SearchEvent{Query: "iphone", UserID: "user1", Timestamp: time.Now().Unix()})

	top := s.GetTop(10)
	log.Printf("Current top (size: %2d): %v\n", len(top), top)
	log.Println("Sliding window and anti-spam ready (Commit 2)")
}
