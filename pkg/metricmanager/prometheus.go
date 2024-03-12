package metricmanager

import (
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// metrics

var (
	MetricPrefix    = "streamey_"
	writeLoopMetric = promauto.NewCounter(prometheus.CounterOpts{
		Name: MetricPrefix + "write_loop",
		Help: "Write loop counter.",
	})
	writeBytesMetric = promauto.NewCounter(prometheus.CounterOpts{
		Name: MetricPrefix + "bytes_written",
		Help: "Write loop counter.",
	})
	readBytesMetric = promauto.NewCounter(prometheus.CounterOpts{
		Name: MetricPrefix + "bytes_read",
		Help: "Read loop counter.",
	})
	validationErrorMetric = promauto.NewCounter(prometheus.CounterOpts{
		Name: MetricPrefix + "validation_error",
		Help: "Validation error counter.",
	})
	audioParseErrorMetric = promauto.NewCounter(prometheus.CounterOpts{
		Name: MetricPrefix + "parse_audio_error",
		Help: "Validation error counter.",
	})
)

func IncWriteLoopCounter() {
	writeLoopMetric.Inc()
}

func IncBytesWrittenCounter(bytes int) {
	writeBytesMetric.Add(float64(bytes))
}

func IncBytesReadCounter(bytes int) {
	readBytesMetric.Add(float64(bytes))
}

func IncValidationErrorCounter() {
	validationErrorMetric.Inc()
}

func IncParseAudioErrorCounter() {
	audioParseErrorMetric.Inc()
}

// metric server

func Listen(port int) {
	http.Handle("/metrics", promhttp.Handler())
	portString := ":" + strconv.Itoa(port)
	http.ListenAndServe(portString, nil)
}
