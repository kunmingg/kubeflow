package monitoring_util

import (
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"net"
	"net/http"
)

// Common label keys for all metrics signals
// COMPONENT: name of component outputing the metrics, eg. "tf-operator"
const COMPONENT = "component"
// KIND: each componenet can label their metrics with custom tag "kind". Suggest keeping "kind" value to be CONSTANT PER METRICS.
const KIND = "kind"
const NAMESPACE = "namespace"
// PATH: request path of the metrics, eg. "/<group>/<version>/namespaces/*/<CRD>/<verb>"
const PATH = "path"

const METRICSPORT = "8079"
const METRICSPATH = "/metrics"

var (
	// Counter metrics
	// num of requests counter vec
	requestCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "request_counter",
			Help: "Number of request_counter",
		},
		[]string{COMPONENT, KIND, NAMESPACE, PATH},
	)
	// Counter metrics for failed requests
	requestFailureCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "request_failure_counter",
		Help: "Number of request_failure_counter",
	}, []string{COMPONENT, KIND, NAMESPACE, PATH})

	// Gauge metrics
	requestGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "requests_gauge",
		Help: "Number of requests_gauge",
	}, []string{COMPONENT, KIND, NAMESPACE, PATH})

	// Gauge metrics for failed requests
	requestFailureGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "requests_failure_gauge",
		Help: "Number of requests_failure_gauge",
	}, []string{COMPONENT, KIND, NAMESPACE, PATH})

	// Linear latencies
	requestLinearLatency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "request_linear_latency",
		Help:    "A histogram of request_linear_latency",
		Buckets: prometheus.LinearBuckets(1, 1, 15),
	}, []string{COMPONENT, KIND, NAMESPACE, PATH})

	// Exponential latencies
	requestExponentialLatency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "request_exponential_latency",
		Help:    "A histogram of request_exponential_latency",
	}, []string{COMPONENT, KIND, NAMESPACE, PATH})

	serviceHeartbeat = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "service_heartbeat",
		Help: "Heartbeat signal every 10 seconds indicating pods are alive.",
	})
)

func init() {
	// Register prometheus counters
	prometheus.MustRegister(requestCounter)
	prometheus.MustRegister(requestFailureCounter)
	prometheus.MustRegister(requestGauge)
	prometheus.MustRegister(requestFailureGauge)
	prometheus.MustRegister(requestLinearLatency)
	prometheus.MustRegister(requestExponentialLatency)
	prometheus.MustRegister(serviceHeartbeat)
}

type MetricsExporter struct {
	// Add error channel so that kf_monitor user can get notified when monitoring thread is down.
	errChan     chan error
	component	string
}

func (me *MetricsExporter) ServeMetrics() {
	mux := http.NewServeMux()
	mux.Handle(METRICSPATH, promhttp.Handler())
	server := http.Server{
		Handler: mux,
	}
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", METRICSPORT))
	if err != nil {
		me.errChan <- err
		return
	}
	// Count heartbeat
	go func() {
		for {
			time.Sleep(10 * time.Second)
			serviceHeartbeat.Inc()
		}
	}()
	// Serve metrics
	go func() {
		log.Info("starting metrics server", "path", METRICSPATH)
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			me.errChan <- err
		}
	}()
}

func (me *MetricsExporter) AddRequestCounter(kind string, namespace string, path string) {
	labels := prometheus.Labels{COMPONENT: me.component, KIND: kind, NAMESPACE: namespace, PATH: path}
	requestCounter.With(labels).Inc()
}

func (me *MetricsExporter) AddRequestFailureCounter(kind string, namespace string, path string) {
	labels := prometheus.Labels{COMPONENT: me.component, KIND: kind, NAMESPACE: namespace, PATH: path}
	requestFailureCounter.With(labels).Inc()
}

func (me *MetricsExporter) AddRequestGauge(kind string, namespace string, path string, val float64) {
	labels := prometheus.Labels{COMPONENT: me.component, KIND: kind, NAMESPACE: namespace, PATH: path}
	requestGauge.With(labels).Add(val)
}

func (me *MetricsExporter) ObserveRequestLinearLatency(kind string, namespace string, path string, val float64) {
	labels := prometheus.Labels{COMPONENT: me.component, KIND: kind, NAMESPACE: namespace, PATH: path}
	requestLinearLatency.With(labels).Observe(val)
}
