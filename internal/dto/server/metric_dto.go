package server

type MetricType int

const (
	Gauge MetricType = iota
	Counter
)

type CounterMetricDto struct {
	Name  string
	Value int64
}

type GaugeMetricDto struct {
	Name  string
	Value float64
}
