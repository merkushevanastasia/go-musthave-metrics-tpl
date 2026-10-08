package service

import (
	"context"

	dto "github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/dto/server"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/repository"
)

var metricRepository = repository.MetricRepository{}

func ProcessGauge(ctx context.Context, dto dto.GaugeMetricDto) {
	metricRepository.UpdateGauge(ctx, dto.Name, dto.Value)
}

func ProcessCounter(ctx context.Context, dto dto.CounterMetricDto) {
	metricRepository.UpdateCounter(ctx, dto.Name, dto.Value)
}
