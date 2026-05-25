package store

import (
	"testing"
	"time"
)

func TestStorage_SlidingWindow(t *testing.T) {
	s := New()
	now := time.Now().Unix()

	s.Add("golang", now)
	s.Add("golang", now-10)

	s.Add("outdated", now-400)

	s.UpdateCache()
	top := s.GetCachedTop(10)

	if len(top) != 1 {
		t.Fatalf("expected top size 1, got %d", len(top))
	}

	if top[0].Query != "golang" {
		t.Errorf("expected top query to be 'golang', got '%s'", top[0].Query)
	}

	if top[0].Count != 2 {
		t.Errorf("expected count to be 2, got %d", top[0].Count)
	}
}
