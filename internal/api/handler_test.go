package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"top-queries-counter/internal/store"
)

func TestHandler_GetTop(t *testing.T) {
	s := store.New()
	now := time.Now().Unix()
	s.Add("alpha", now)
	s.Add("beta", now)
	s.Add("beta", now)
	s.UpdateCache()

	h := NewHandler(s, 100)
	req := httptest.NewRequest(http.MethodGet, "/api/top?n=1", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}

	var items []store.Item
	if err := json.NewDecoder(rec.Body).Decode(&items); err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].Query != "beta" {
		t.Fatalf("unexpected top: %+v", items)
	}
}

func TestHandler_MaxTopN(t *testing.T) {
	s := store.New()
	h := NewHandler(s, 5)
	req := httptest.NewRequest(http.MethodGet, "/api/top?n=100", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	var items []store.Item
	_ = json.NewDecoder(rec.Body).Decode(&items)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
}

func TestHandler_StopList(t *testing.T) {
	s := store.New()
	now := time.Now().Unix()
	s.Add("casino", now)
	s.UpdateCache()

	h := NewHandler(s, 100)
	body := `{"word":"casino"}`
	req := httptest.NewRequest(http.MethodPost, "/api/stoplist", strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("post status %d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/top?n=10", nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	var items []store.Item
	_ = json.NewDecoder(rec.Body).Decode(&items)
	if len(items) != 0 {
		t.Fatalf("stop-list must purge query from top, got %+v", items)
	}
}
