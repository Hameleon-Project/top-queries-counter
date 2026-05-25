package store

import (
	"sort"
	"sync"
	"time"
)

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

func (s *Storage) IsStopWord(query string) bool {
	s.stopMu.RLock()
	defer s.stopMu.RUnlock()
	_, exists := s.stopWords[query]
	return exists
}

func (s *Storage) AddStopWord(word string) {
	s.stopMu.Lock()
	s.stopWords[word] = struct{}{}
	s.stopMu.Unlock()
}

func (s *Storage) RemoveStopWord(word string) {
	s.stopMu.Lock()
	delete(s.stopWords, word)
	s.stopMu.Unlock()
}

func (s *Storage) Add(query string, timestamp int64) {
	if s.IsStopWord(query) {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.history[query] = append(s.history[query], timestamp)
}

func (s *Storage) UpdateCache() {
	s.mu.Lock()
	boundary := time.Now().Unix() - 300
	var list []Item

	for query, timestamps := range s.history {
		validIdx := 0
		for _, ts := range timestamps {
			if ts >= boundary {
				timestamps[validIdx] = ts
				validIdx++
			}
		}
		s.history[query] = timestamps[:validIdx]

		if validIdx > 0 {
			list = append(list, Item{
				Query: query,
				Count: validIdx,
			})
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

func (s *Storage) GetCachedTop(limit int) []Item {
	s.cacheMu.RLock()
	defer s.cacheMu.RUnlock()

	if len(s.cache) == 0 {
		return []Item{}
	}

	if len(s.cache) > limit {
		return s.cache[:limit]
	}

	res := make([]Item, len(s.cache))
	copy(res, s.cache)
	return res
}
