package store

import (
	"testing"
	"time"
)

func BenchmarkStorage_Add(b *testing.B) {
	s := New()
	now := time.Now().Unix()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Add("query", now)
	}
}

func BenchmarkStorage_Top(b *testing.B) {
	s := New()
	now := time.Now().Unix()
	for i := 0; i < 1000; i++ {
		s.Add("q"+string(rune('a'+i%26)), now)
	}
	s.UpdateCache()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.Top(10)
	}
}
