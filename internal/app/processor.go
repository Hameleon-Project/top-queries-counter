package app

import (
	"sync"
	"top-queries-counter/internal/store"
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
