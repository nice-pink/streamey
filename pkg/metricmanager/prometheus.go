package metricmanager

import (
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// metrics

var ()

type MetricManager struct {
	writeLoopMetric       prometheus.Counter
	writeBytesMetric      prometheus.Counter
	readBytesMetric       prometheus.Counter
	validationErrorMetric prometheus.Counter
	audioParseErrorMetric prometheus.Counter
}

func NewMetricManager(prefix string, url string) *MetricManager {
	mm := MetricManager{}

	//
	mm.writeLoopMetric = promauto.NewCounter(prometheus.CounterOpts{
		Name:        prefix + "write_loop",
		Help:        "Write loop counter.",
		ConstLabels: prometheus.Labels{"url": url},
	})
	mm.writeBytesMetric = promauto.NewCounter(prometheus.CounterOpts{
		Name:        prefix + "bytes_written",
		Help:        "Write loop counter.",
		ConstLabels: prometheus.Labels{"url": url},
	})
	mm.readBytesMetric = promauto.NewCounter(prometheus.CounterOpts{
		Name:        prefix + "bytes_read",
		Help:        "Read loop counter.",
		ConstLabels: prometheus.Labels{"url": url},
	})
	mm.validationErrorMetric = promauto.NewCounter(prometheus.CounterOpts{
		Name:        prefix + "validation_error",
		Help:        "Validation error counter.",
		ConstLabels: prometheus.Labels{"url": url},
	})
	mm.audioParseErrorMetric = promauto.NewCounter(prometheus.CounterOpts{
		Name:        prefix + "parse_audio_error",
		Help:        "Validation error counter.",
		ConstLabels: prometheus.Labels{"url": url},
	})

	return &mm
}

func (m *MetricManager) IncWriteLoopCounter() {
	m.writeLoopMetric.Inc()
}

func (m *MetricManager) IncBytesWrittenCounter(bytes int) {
	m.writeBytesMetric.Add(float64(bytes))
}

func (m *MetricManager) IncBytesReadCounter(bytes int) {
	m.readBytesMetric.Add(float64(bytes))
}

func (m *MetricManager) IncValidationErrorCounter() {
	m.validationErrorMetric.Inc()
}

func (m *MetricManager) IncParseAudioErrorCounter() {
	m.audioParseErrorMetric.Inc()
}

// metric server

func Listen(port int) {
	http.Handle("/metrics", promhttp.Handler())
	portString := ":" + strconv.Itoa(port)
	http.ListenAndServe(portString, nil)
}
