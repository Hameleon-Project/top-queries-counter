package antispam

import (
	"sync"
	"time"
)

type Guard struct {
	mu sync.Mutex

	userQueryLast map[string]int64
	ipBucket      map[string]*bucket
	queryBucket   map[string]*bucket

	userCooldownSec int64
	maxPerIPMin     int
	maxQueryMin     int
}

type bucket struct {
	minute int64
	count  int
}

func New(userCooldownSec int64, maxPerIPMin, maxQueryMin int) *Guard {
	if userCooldownSec <= 0 {
		userCooldownSec = 5
	}
	if maxPerIPMin <= 0 {
		maxPerIPMin = 60
	}
	if maxQueryMin <= 0 {
		maxQueryMin = 500
	}
	return &Guard{
		userQueryLast:   make(map[string]int64),
		ipBucket:        make(map[string]*bucket),
		queryBucket:     make(map[string]*bucket),
		userCooldownSec: userCooldownSec,
		maxPerIPMin:     maxPerIPMin,
		maxQueryMin:     maxQueryMin,
	}
}

func (g *Guard) Allow(userID, query, ip string, ts int64) bool {
	if ts == 0 {
		ts = time.Now().Unix()
	}
	if ip == "" {
		ip = "unknown"
	}
	if userID == "" {
		userID = ip
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	uqKey := userID + ":" + query
	if last, ok := g.userQueryLast[uqKey]; ok && ts-last < g.userCooldownSec {
		return false
	}
	g.userQueryLast[uqKey] = ts

	if !g.inc(g.ipBucket, ip, ts, g.maxPerIPMin) {
		return false
	}
	if !g.inc(g.queryBucket, query, ts, g.maxQueryMin) {
		return false
	}

	if len(g.userQueryLast) > 100_000 {
		g.evict(ts)
	}
	return true
}

func (g *Guard) inc(m map[string]*bucket, key string, ts int64, limit int) bool {
	minute := ts / 60
	b, ok := m[key]
	if !ok || b.minute != minute {
		m[key] = &bucket{minute: minute, count: 1}
		return true
	}
	b.count++
	return b.count <= limit
}

func (g *Guard) evict(now int64) {
	for k, t := range g.userQueryLast {
		if now-t > 120 {
			delete(g.userQueryLast, k)
		}
	}
	cutoff := now/60 - 2
	for k, b := range g.ipBucket {
		if b.minute < cutoff {
			delete(g.ipBucket, k)
		}
	}
	for k, b := range g.queryBucket {
		if b.minute < cutoff {
			delete(g.queryBucket, k)
		}
	}
}
