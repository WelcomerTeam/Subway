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
			Name: "subway_interaction_count",
			Help: "Count of total interactions received",
		},
		[]string{"name", "guild_id", "user_id"},
	)

	subwaySuccessfulInteractionTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "subway_successful_interaction_count",
			Help: "Count of successful interactions received",
		},
	)

	subwayFailedInteractionTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "subway_failed_interaction_count",
			Help: "Count of failed interactions received",
		},
	)
)

// SetupPrometheus sets up prometheus.
func (subway *Subway) SetupPrometheus() error {
	subway.Logger.Info().Msgf("Serving prometheus at %s", subway.prometheusAddress)

	prometheus.MustRegister(subwayInteractionProcessingTimeName)
	prometheus.MustRegister(subwayInteractionTotal)
	prometheus.MustRegister(subwaySuccessfulInteractionTotal)
	prometheus.MustRegister(subwayFailedInteractionTotal)

	http.Handle("/metrics", promhttp.HandlerFor(
		prometheus.DefaultGatherer,
		promhttp.HandlerOpts{},
	))

	err := http.ListenAndServe(subway.prometheusAddress, nil)
	if err != nil {
		subway.Logger.Error().Str("host", subway.prometheusAddress).Err(err).Msg("Failed to serve prometheus server")

		return fmt.Errorf("failed to serve prometheus: %w", err)
	}

	return nil
}
