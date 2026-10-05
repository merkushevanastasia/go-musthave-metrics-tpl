package service

import (
	"context"

	dto "github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/dto/server"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/repository"
)

type MetricService struct {
	repo repository.MetricRepository
}

func NewMetricService(repo repository.MetricRepository) *MetricService {
	return &MetricService{
		repo: repo,
	}
}

func (s *MetricService) ProcessGauge(ctx context.Context, dto dto.GaugeMetricDto) {
	s.repo.UpdateGauge(ctx, dto.Name, dto.Value)
}

func (s *MetricService) ProcessCounter(ctx context.Context, dto dto.CounterMetricDto) {
	s.repo.UpdateCounter(ctx, dto.Name, dto.Value)
}
