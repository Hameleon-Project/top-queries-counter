package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"top-queries-counter/internal/store"
)

type Handler struct {
	store *store.Storage
}

func NewHandler(s *store.Storage) *http.Handler {
	h := &Handler{store: s}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/top", h.GetTop)
	mux.HandleFunc("POST /api/stoplist", h.AddStopList)
	mux.HandleFunc("DELETE /api/stoplist", h.RemoveStopList)

	handler := http.Handler(mux)
	return &handler
}

func (h *Handler) GetTop(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 10
	if idx, err := strconv.Atoi(limitStr); err == nil && idx > 0 {
		limit = idx
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(h.store.GetCachedTop(limit))
}

func (h *Handler) AddStopList(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Word string `json:"word"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Word == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	h.store.AddStopWord(req.Word)
	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) RemoveStopList(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Word string `json:"word"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Word == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	h.store.RemoveStopWord(req.Word)
	w.WriteHeader(http.StatusOK)
}
