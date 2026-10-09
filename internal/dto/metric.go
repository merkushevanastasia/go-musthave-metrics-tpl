package dto

import (
	"sync"

	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/constants"
)

type MetricCollection struct {
	Mu      sync.RWMutex
	Metrics map[string]MetricDto
}

type MetricDto struct {
	MetricName string
	MetricType string
	Gauge      float64
	Counter    int64
}

func NewMetricCollection() *MetricCollection {
	return &MetricCollection{Metrics: make(map[string]MetricDto)}
}

func (mc *MetricCollection) AddCounter(metricName string, value int64) {
	metricValue := mc.Metrics[metricName]
	metricValue.Counter += value
	metricValue.MetricType = constants.CounterMetricType
	mc.Metrics[metricName] = metricValue
}

func (mc *MetricCollection) AddGauge(metricName string, value float64) {
	metricValue := mc.Metrics[metricName]
	metricValue.Gauge = value
	metricValue.MetricType = constants.GaugeMetricType
	mc.Metrics[metricName] = metricValue
}
