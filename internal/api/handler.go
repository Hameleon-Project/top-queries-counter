package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"top-queries-counter/internal/metrics"
	"top-queries-counter/internal/store"
)

type Handler struct {
	store   *store.Storage
	maxTopN int
}

func NewHandler(s *store.Storage, maxTopN int) http.Handler {
	if maxTopN <= 0 {
		maxTopN = 100
	}
	h := &Handler{store: s, maxTopN: maxTopN}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", h.Health)
	mux.HandleFunc("GET /api/top", h.GetTop)
	mux.HandleFunc("GET /api/stoplist", h.ListStopList)
	mux.HandleFunc("POST /api/stoplist", h.AddStopList)
	mux.HandleFunc("DELETE /api/stoplist", h.RemoveStopList)
	mux.Handle("GET /metrics", metrics.Handler())
	return mux
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
	metrics.HTTPRequests.WithLabelValues("health", "200").Inc()
}

func (h *Handler) GetTop(w http.ResponseWriter, r *http.Request) {
	limit := clampLimit(parseLimit(r.URL.Query().Get("n"), r.URL.Query().Get("limit")), h.maxTopN)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(h.store.TopQueries(limit)); err != nil {
		http.Error(w, "encode error", http.StatusInternalServerError)
		metrics.HTTPRequests.WithLabelValues("top", "500").Inc()
		return
	}
	metrics.HTTPRequests.WithLabelValues("top", "200").Inc()
}

func parseLimit(nParam, limitParam string) int {
	limit := 10
	for _, raw := range []string{nParam, limitParam} {
		if raw == "" {
			continue
		}
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			return n
		}
	}
	return limit
}

func clampLimit(n, max int) int {
	if n > max {
		return max
	}
	return n
}

func (h *Handler) ListStopList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string][]string{"words": h.store.ListStopWords()})
	metrics.HTTPRequests.WithLabelValues("stoplist_get", "200").Inc()
}

func (h *Handler) AddStopList(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Word string `json:"word"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Word == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		metrics.HTTPRequests.WithLabelValues("stoplist_post", "400").Inc()
		return
	}
	h.store.AddStopWord(req.Word)
	w.WriteHeader(http.StatusCreated)
	metrics.HTTPRequests.WithLabelValues("stoplist_post", "201").Inc()
}

func (h *Handler) RemoveStopList(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Word string `json:"word"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Word == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		metrics.HTTPRequests.WithLabelValues("stoplist_delete", "400").Inc()
		return
	}
	h.store.RemoveStopWord(req.Word)
	w.WriteHeader(http.StatusOK)
	metrics.HTTPRequests.WithLabelValues("stoplist_delete", "200").Inc()
}
