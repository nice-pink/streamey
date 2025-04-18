package util

type MetricsControl struct {
	Enabled bool
	Prefix  string
	Labels  map[string]string
}
