package prometheus

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/agreyfox/eshop/system/logs"
	"github.com/go-zoo/bone"

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

func Pong(histogram *prometheus.HistogramVec) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		//monitoring how long it takes to respond
		start := time.Now()
		defer r.Body.Close()
		code := 500

		defer func() {
			httpDuration := time.Since(start)
			histogram.WithLabelValues(fmt.Sprintf("%d", code)).Observe(httpDuration.Seconds())
		}()

		code = http.StatusBadRequest // if req is not GET
		if r.Method == "GET" {
			code = http.StatusOK
			greet := fmt.Sprint("Pong \n")
			w.Write([]byte(greet))
		} else {
			w.WriteHeader(code)
		}
	}
}

// to run prometheus service
func Run(port string, mainMux *bone.Mux) bool {
	logger.Infof("Eshop Prometheus Service start at %s\n", port)
	rand.Seed(time.Now().Unix())
	//server := http.NewServeMux()
	server := bone.New()
	server.Handle("/metrics", promhttp.Handler())
	ht := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "ping_seconds",
		Help:    "Time take to ping someone",
		Buckets: []float64{1, 2, 5, 6, 10}, //defining small buckets as this app should not take more than 1 sec to respond
	}, []string{"code"}) // this will be partitioned by the HTTP code.
	mainMux.Handle("/ping", Pong(ht))

	prometheus.MustRegister(counter)
	prometheus.MustRegister(gauge)
	prometheus.MustRegister(histogram)
	prometheus.MustRegister(summary)
	prometheus.Register(ht)

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
