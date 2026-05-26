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
	top := s.TopQueries(10)

	if len(top) != 1 {
		t.Fatalf("expected top size 1, got %d", len(top))
	}
	if top[0].Query != "golang" {
		t.Errorf("expected query golang, got %q", top[0].Query)
	}
	if top[0].Count != 2 {
		t.Errorf("expected count 2, got %d", top[0].Count)
	}
}

func TestStorage_StopListPurgesHistory(t *testing.T) {
	s := New()
	now := time.Now().Unix()
	s.Add("casino", now)
	s.AddStopWord("casino")

	s.UpdateCache()
	if len(s.TopQueries(10)) != 0 {
		t.Fatal("stop-word must remove query from top immediately")
	}
}

func TestStorage_NormalizeQuery(t *testing.T) {
	s := New()
	now := time.Now().Unix()
	s.Add("  iPhone  ", now)
	s.Add("iphone", now)

	s.UpdateCache()
	top := s.TopQueries(10)
	if len(top) != 1 || top[0].Count != 2 {
		t.Fatalf("expected merged normalized query, got %+v", top)
	}
}

func TestStorage_TopLimit(t *testing.T) {
	s := New()
	now := time.Now().Unix()
	for i := 0; i < 5; i++ {
		s.Add("a", now)
	}
	for i := 0; i < 3; i++ {
		s.Add("b", now)
	}
	s.UpdateCache()

	top := s.TopQueries(1)
	if len(top) != 1 || top[0].Query != "a" {
		t.Fatalf("expected top-1, got %+v", top)
	}
}
