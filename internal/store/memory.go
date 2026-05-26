package store

import (
	"sort"
	"strings"
	"sync"
	"time"
)

const WindowSeconds = 300

type Item struct {
	Query string `json:"query"`
	Count int    `json:"count"`
}

type Storage struct {
	mu      sync.RWMutex
	history map[string][]int64

	stopMu    sync.RWMutex
	stopWords map[string]struct{}

	cacheMu sync.RWMutex
	cache   []Item
}

func New() *Storage {
	return &Storage{
		history:   make(map[string][]int64),
		stopWords: make(map[string]struct{}),
		cache:     make([]Item, 0),
	}
}

func NormalizeQuery(query string) string {
	return strings.TrimSpace(strings.ToLower(query))
}

func (s *Storage) IsStopWord(query string) bool {
	query = NormalizeQuery(query)
	s.stopMu.RLock()
	defer s.stopMu.RUnlock()
	_, exists := s.stopWords[query]
	return exists
}

func (s *Storage) AddStopWord(word string) {
	word = NormalizeQuery(word)
	if word == "" {
		return
	}
	s.stopMu.Lock()
	s.stopWords[word] = struct{}{}
	s.stopMu.Unlock()
	s.PurgeQuery(word)
}

func (s *Storage) RemoveStopWord(word string) {
	word = NormalizeQuery(word)
	s.stopMu.Lock()
	delete(s.stopWords, word)
	s.stopMu.Unlock()
}

func (s *Storage) ListStopWords() []string {
	s.stopMu.RLock()
	defer s.stopMu.RUnlock()
	out := make([]string, 0, len(s.stopWords))
	for w := range s.stopWords {
		out = append(out, w)
	}
	sort.Strings(out)
	return out
}

func (s *Storage) PurgeQuery(query string) {
	query = NormalizeQuery(query)
	s.mu.Lock()
	delete(s.history, query)
	s.mu.Unlock()
	s.UpdateCache()
}

func (s *Storage) Add(query string, timestamp int64) {
	query = NormalizeQuery(query)
	if query == "" || s.IsStopWord(query) {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.history[query] = append(s.history[query], timestamp)
}

func (s *Storage) UpdateCache() {
	s.mu.Lock()
	boundary := time.Now().Unix() - WindowSeconds
	list := make([]Item, 0, len(s.history))

	for query, timestamps := range s.history {
		if s.IsStopWord(query) {
			delete(s.history, query)
			continue
		}

		validIdx := 0
		for _, ts := range timestamps {
			if ts >= boundary {
				timestamps[validIdx] = ts
				validIdx++
			}
		}
		s.history[query] = timestamps[:validIdx]

		if validIdx > 0 {
			list = append(list, Item{Query: query, Count: validIdx})
		} else {
			delete(s.history, query)
		}
	}
	s.mu.Unlock()

	sort.Slice(list, func(i, j int) bool {
		if list[i].Count == list[j].Count {
			return list[i].Query < list[j].Query
		}
		return list[i].Count > list[j].Count
	})

	s.cacheMu.Lock()
	s.cache = list
	s.cacheMu.Unlock()
}

func (s *Storage) TopQueries(limit int) []Item {
	if limit <= 0 {
		limit = 10
	}

	s.cacheMu.RLock()
	defer s.cacheMu.RUnlock()

	if len(s.cache) == 0 {
		return []Item{}
	}
	if len(s.cache) <= limit {
		out := make([]Item, len(s.cache))
		copy(out, s.cache)
		return out
	}
	out := make([]Item, limit)
	copy(out, s.cache[:limit])
	return out
}

func (s *Storage) Top(limit int) []Item {
	return s.TopQueries(limit)
}
