package main

import (
	"encoding/json"
	"flag"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type searchEvent struct {
	Query     string `json:"query"`
	UserID    string `json:"user_id"`
	IP        string `json:"ip"`
	Timestamp int64  `json:"timestamp"`
}

func main() {
	amqpURL := flag.String("url", "amqp://guest:guest@localhost:5672/", "RabbitMQ URL")
	queue := flag.String("queue", "search_logs", "queue name")
	query := flag.String("query", "iphone 15", "search query")
	user := flag.String("user", "demo-user", "user id")
	count := flag.Int("n", 1, "messages to publish")
	flag.Parse()

	conn, err := amqp.Dial(*amqpURL)
	if err != nil {
		log.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("channel: %v", err)
	}
	defer ch.Close()

	if _, err := ch.QueueDeclare(*queue, true, false, false, false, nil); err != nil {
		log.Fatalf("declare queue: %v", err)
	}

	now := time.Now().Unix()
	for i := 0; i < *count; i++ {
		ev := searchEvent{
			Query:     *query,
			UserID:    *user,
			Timestamp: now,
		}
		body, _ := json.Marshal(ev)
		if err := ch.Publish("", *queue, false, false, amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		}); err != nil {
			log.Fatalf("publish: %v", err)
		}
	}
	log.Printf("published %d message(s) to queue %q", *count, *queue)
}
