package app

import (
	"testing"
	"time"

	"top-queries-counter/internal/antispam"
	"top-queries-counter/internal/store"
)

func newTestProcessor() *Processor {
	return NewProcessor(store.New(), antispam.New(5, 100, 100))
}

func TestProcessor_AntiSpam(t *testing.T) {
	p := newTestProcessor()
	now := time.Now().Unix()

	ev := SearchEvent{Query: "iphone", UserID: "u1", IP: "1.1.1.1", Timestamp: now}
	p.Process(ev)
	p.Process(ev)

	s := p.store
	s.UpdateCache()
	top := s.TopQueries(10)
	if len(top) != 1 || top[0].Count != 1 {
		t.Fatalf("anti-spam should allow one event, got %+v", top)
	}
}

func TestProcessor_DefaultTimestamp(t *testing.T) {
	p := newTestProcessor()
	p.Process(SearchEvent{Query: "laptop", UserID: "u2", IP: "2.2.2.2"})
	p.store.UpdateCache()

	if len(p.store.TopQueries(10)) != 1 {
		t.Fatal("expected event with default timestamp")
	}
}

func TestProcessor_EmptyQuery(t *testing.T) {
	p := newTestProcessor()
	p.Process(SearchEvent{Query: "   ", UserID: "u1", IP: "1.1.1.1"})
	p.store.UpdateCache()
	if len(p.store.TopQueries(10)) != 0 {
		t.Fatal("empty query must be ignored")
	}
}

func TestProcessor_RequiresIdentity(t *testing.T) {
	p := newTestProcessor()
	p.Process(SearchEvent{Query: "phone", Timestamp: time.Now().Unix()})
	p.store.UpdateCache()
	if len(p.store.TopQueries(10)) != 0 {
		t.Fatal("event without user_id and ip must be ignored")
	}
}

func TestProcessor_StaleTimestamp(t *testing.T) {
	p := newTestProcessor()
	old := time.Now().Unix() - 400
	p.Process(SearchEvent{Query: "old", UserID: "u1", IP: "1.1.1.1", Timestamp: old})
	p.store.UpdateCache()
	if len(p.store.TopQueries(10)) != 0 {
		t.Fatal("stale event must be ignored")
	}
}
