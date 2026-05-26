package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	EventsProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "search_events_processed_total",
		Help: "Total search events accepted into the store",
	})
	EventsDropped = promauto.NewCounter(prometheus.CounterOpts{
		Name: "search_events_dropped_total",
		Help: "Events dropped by anti-spam or validation",
	})
	HTTPRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "HTTP requests by route and status",
	}, []string{"route", "code"})
)

func Handler() http.Handler {
	return promhttp.Handler()
}
