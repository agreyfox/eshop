package prometheus

import (
	"math/rand"
	"net/http"
	"time"

	"github.com/agreyfox/eshop/system/logs"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

var (
	counter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "golang",
			Name:      "eshop_counter",
			Help:      "This is my counter",
		})

	gauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "golang",
			Name:      "eshop_gauge",
			Help:      "This is my gauge",
		})

	histogram = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: "golang",
			Name:      "eshop_histogram",
			Help:      "This is my histogram",
		})

	summary = prometheus.NewSummary(
		prometheus.SummaryOpts{
			Namespace: "golang",
			Name:      "eshop_summary",
			Help:      "This is my summary",
		})

	logger *zap.SugaredLogger = logs.Log.Sugar()
)

// to run prometheus service
func run(port string) bool {
	logger.Info("Eshop Prometheus Service start at %s\n", port)
	rand.Seed(time.Now().Unix())
	server := http.NewServeMux()
	server.Handle("/metrics", promhttp.Handler())

	prometheus.MustRegister(counter)
	prometheus.MustRegister(gauge)
	prometheus.MustRegister(histogram)
	prometheus.MustRegister(summary)

	go func() {
		for {
			counter.Add(rand.Float64() * 5)
			gauge.Add(rand.Float64()*15 - 5)
			histogram.Observe(rand.Float64() * 10)
			summary.Observe(rand.Float64() * 10)
			time.Sleep(time.Second)
		}
	}()
	logger.Fatal(http.ListenAndServe(port, server))
	return true
}
