package metricmanager

import (
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// metric server

func Listen(port int) {
	http.Handle("/metrics", promhttp.Handler())
	portString := ":" + strconv.Itoa(port)
	http.ListenAndServe(portString, nil)
}
