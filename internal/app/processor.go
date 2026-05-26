package app

import (
	"encoding/json"
	"log"
	"time"

	"top-queries-counter/internal/antispam"
	"top-queries-counter/internal/metrics"
	"top-queries-counter/internal/store"

	amqp "github.com/rabbitmq/amqp091-go"
)

type SearchEvent struct {
	Query     string `json:"query"`
	UserID    string `json:"user_id"`
	IP        string `json:"ip"`
	Timestamp int64  `json:"timestamp"`
}

const windowSeconds = 300

type Processor struct {
	store *store.Storage
	guard *antispam.Guard
}

func NewProcessor(s *store.Storage, guard *antispam.Guard) *Processor {
	return &Processor{store: s, guard: guard}
}

func (p *Processor) ListenRabbit(amqpURL, queueName string, stopChan <-chan struct{}) {
	backoff := 2 * time.Second
	for {
		select {
		case <-stopChan:
			log.Println("Stopping RabbitMQ consumer...")
			return
		default:
		}

		err := p.consume(amqpURL, queueName, stopChan)
		if err == nil {
			return
		}
		log.Printf("RabbitMQ consumer stopped: %v", err)

		select {
		case <-stopChan:
			return
		case <-time.After(backoff):
		}
		if backoff < 30*time.Second {
			backoff *= 2
		}
	}
}

func (p *Processor) consume(amqpURL, queueName string, stopChan <-chan struct{}) error {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return err
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		return err
	}

	msgs, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		return err
	}

	log.Printf("RabbitMQ consumer connected, queue=%s", queueName)

	for {
		select {
		case <-stopChan:
			return nil
		case d, ok := <-msgs:
			if !ok {
				return amqp.ErrClosed
			}
			if len(d.Body) == 0 {
				continue
			}
			var ev SearchEvent
			if err := json.Unmarshal(d.Body, &ev); err != nil {
				log.Printf("invalid message JSON: %v", err)
				metrics.EventsDropped.Inc()
				continue
			}
			p.Process(ev)
		}
	}
}

func (p *Processor) Process(ev SearchEvent) {
	query := store.NormalizeQuery(ev.Query)
	if query == "" {
		metrics.EventsDropped.Inc()
		return
	}

	ts := ev.Timestamp
	if ts == 0 {
		ts = time.Now().Unix()
	}

	now := time.Now().Unix()
	if ts < now-windowSeconds || ts > now+60 {
		metrics.EventsDropped.Inc()
		return
	}

	userID := ev.UserID
	ip := ev.IP
	if userID == "" && ip == "" {
		metrics.EventsDropped.Inc()
		return
	}

	if !p.guard.Allow(userID, query, ip, ts) {
		metrics.EventsDropped.Inc()
		return
	}

	p.store.Add(query, ts)
	metrics.EventsProcessed.Inc()
}
