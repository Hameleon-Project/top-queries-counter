package app

import (
	"encoding/json"
	"log"
	"sync"

	"top-queries-counter/internal/store"

	amqp "github.com/rabbitmq/amqp091-go"
)

type SearchEvent struct {
	Query     string `json:"query"`
	UserID    string `json:"user_id"`
	IP        string `json:"ip"`
	Timestamp int64  `json:"timestamp"`
}

type Processor struct {
	store    *store.Storage
	mu       sync.Mutex
	antiSpam map[string]int64
}

func NewProcessor(s *store.Storage) *Processor {
	return &Processor{
		store:    s,
		antiSpam: make(map[string]int64),
	}
}

func (p *Processor) ListenRabbit(amqpURL, queueName string, stopChan chan struct{}) {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		log.Fatalf("failed to connect to rabbitmq: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("failed to open a channel: %v", err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("failed to declare a queue: %v", err)
	}

	msgs, err := ch.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("failed to register a consumer: %v", err)
	}

	log.Printf("Started RabbitMQ consumer on queue: %s", queueName)

	for {
		select {
		case d := <-msgs:
			if len(d.Body) == 0 {
				continue
			}
			var ev SearchEvent
			if err := json.Unmarshal(d.Body, &ev); err != nil {
				log.Printf("failed to unmarshal amqp message: %v", err)
				continue
			}
			p.Process(ev)
		case <-stopChan:
			log.Println("Stopping RabbitMQ consumer gracefully...")
			return
		}
	}
}

func (p *Processor) Process(ev SearchEvent) {
	if ev.Query == "" {
		return
	}

	userID := ev.UserID
	if userID == "" {
		userID = ev.IP
	}
	key := userID + ":" + ev.Query

	p.mu.Lock()
	lastTime, exists := p.antiSpam[key]
	if exists && ev.Timestamp-lastTime < 5 {
		p.mu.Unlock()
		return
	}
	p.antiSpam[key] = ev.Timestamp

	if len(p.antiSpam) > 50000 {
		for k, ts := range p.antiSpam {
			if ev.Timestamp-ts > 10 {
				delete(p.antiSpam, k)
			}
		}
	}
	p.mu.Unlock()

	p.store.Add(ev.Query, ev.Timestamp)
}
