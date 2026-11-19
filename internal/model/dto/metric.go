package dto

import "sync"

type MetricCollection struct {
	Mu      sync.RWMutex
	Metrics map[string]MetricDto
}

func NewMetricDtoCollection() *MetricCollection {
	return &MetricCollection{Metrics: make(map[string]MetricDto)}
}

func (m *MetricCollection) Add(entity MetricDto) {
	m.Metrics[entity.ID] = entity
}

func (m *MetricCollection) GetCurrent(mName string) int64 {
	dto, present := m.Metrics[mName]
	if !present {
		return 0
	}
	return *dto.Delta
}

type MetricDto struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}
