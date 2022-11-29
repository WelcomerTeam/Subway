package internal

import (
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	subwayInteractionProcessingTimeName = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "subway_interaction_processing_seconds",
			Help: "Time taken (in seconds) spent processing interactions",
		},
		[]string{"name", "guild_id", "user_id"},
	)

	subwayInteractionTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "subway_interaction_total",
			Help: "Total interactions received",
		},
		[]string{"name", "guild_id", "user_id"},
	)

	subwaySuccessfulInteractionTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "subway_successful_interaction_total",
			Help: "Total successful interactions received",
		},
	)

	subwayFailedInteractionTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "subway_failed_interaction_total",
			Help: "Total failed interactions received",
		},
	)
)

// SetupPrometheus sets up prometheus.
func (sub *Subway) SetupPrometheus() error {
	sub.Logger.Info().Msgf("Serving prometheus at %s", sub.prometheusAddress)

	prometheus.MustRegister(subwayInteractionProcessingTimeName)
	prometheus.MustRegister(subwayInteractionTotal)
	prometheus.MustRegister(subwaySuccessfulInteractionTotal)
	prometheus.MustRegister(subwayFailedInteractionTotal)

	prometheusMux := http.NewServeMux()
	prometheusMux.Handle("/metrics", promhttp.HandlerFor(
		prometheus.DefaultGatherer,
		promhttp.HandlerOpts{},
	))

	err := http.ListenAndServe(sub.prometheusAddress, prometheusMux)
	if err != nil {
		sub.Logger.Error().Str("host", sub.prometheusAddress).Err(err).Msg("Failed to serve prometheus server")

		return fmt.Errorf("failed to serve prometheus: %w", err)
	}

	return nil
}
