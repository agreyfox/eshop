package prometheus

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/agreyfox/eshop/system/logs"
	"github.com/go-zoo/bone"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

var (
	ApiCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "egpal",
			Name:      "egpal_api_Request",
			Help:      " 使用lqcms后台内容调用计数器,每次不同页面调用不同内容",
		},
		[]string{"type", "method"})
	OrderCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "egpal",
			Name:      "egpal_order_Counter",
			Help:      "创建订单数,用户尝试下单即会产生一个订单号",
		},
		[]string{"method"})
	RegisterCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "egpal",
			Name:      "register_Number",
			Help:      "统计系统运行以来的注册人数总和",
		})
	PaypalOrderCreateCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "egpal",
			Name:      "Paypal_notification_of_order",
			Help:      "理论上，一个notification或创建一个paypal订单",
		})

	PaypalNotifiyCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "egpal",
			Name:      "Paypal_all_notification",
			Help:      " 发送通知次数,查看paypal和egpal连接通讯是否正常",
		},
		[]string{"method"})

	Histogram = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: "egpal",
			Name:      "http_request_histogram",
			Help:      "egpal http request ",
		})

	summary = prometheus.NewSummary(
		prometheus.SummaryOpts{
			Namespace: "golang",
			Name:      "eshop_summary",
			Help:      "This is my summary",
		})

	logger *zap.SugaredLogger = logs.Log.Sugar()
)

func init() {
	prometheus.MustRegister(ApiCounter)
	prometheus.MustRegister(OrderCounter)
	prometheus.MustRegister(RegisterCounter)
	prometheus.MustRegister(PaypalNotifiyCounter)
	prometheus.MustRegister(PaypalOrderCreateCounter)
	//prometheus.MustRegister(Histogram)

}
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
	server.Handle("/metrics", prometheus.Handler())

	ht := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "ping_seconds",
		Help:    "pong 的响应时间",
		Buckets: []float64{1, 2, 5, 6, 10}, //defining small buckets as this app should not take more than 1 sec to respond
	}, []string{"code"}) // this will be partitioned by the HTTP code.
	mainMux.Handle("/ping", Pong(ht))

	/* prometheus.MustRegister(counter)
	prometheus.MustRegister(gauge)
	prometheus.MustRegister(histogram)
	prometheus.MustRegister(summary) */
	prometheus.Register(ht)
	/*
		go func() {
			for {
				counter.Add(rand.Float64() * 5)
				gauge.Add(rand.Float64()*15 - 5)
				histogram.Observe(rand.Float64() * 10)
				summary.Observe(rand.Float64() * 10)
				time.Sleep(time.Second)
			}
		}() */
	logger.Fatal(http.ListenAndServe(port, server))
	return true
}
