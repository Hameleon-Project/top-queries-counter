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
}

func New() *Storage {
	return &Storage{
		history: make(map[string][]int64),
	}
}

func (s *Storage) Add(query string, timestamp int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.history[query] = append(s.history[query], timestamp)
}

func (s *Storage) GetTop(limit int) []Item {
	s.mu.Lock()
	defer s.mu.Unlock()

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

	sort.Slice(list, func(i, j int) bool {
		if list[i].Count == list[j].Count {
			return list[i].Query < list[j].Query
		}
		return list[i].Count > list[j].Count
	})

	if len(list) > limit {
		return list[:limit]
	}
	return list
}
