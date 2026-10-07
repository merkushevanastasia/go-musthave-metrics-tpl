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

func (s *MetricService) GetGaugeValue(ctx context.Context, metricName string) (float64, error) {
	value, err := s.repo.GetGauge(ctx, metricName)
	return value, err
}

func (s *MetricService) GetCounterValue(ctx context.Context, metricName string) (int64, error) {
	value, err := s.repo.GetCounter(ctx, metricName)
	return value, err
}

func (s *MetricService) GetAll(ctx context.Context) ([]dto.CounterMetricDto, []dto.GaugeMetricDto) {
	gaugeMetricEntities, counterMetricEntities := s.repo.GetAll(ctx)

	resCounterList := make([]dto.CounterMetricDto, 0)
	resGaugeList := make([]dto.GaugeMetricDto, 0)

	for k, v := range counterMetricEntities {
		resCounterList = append(resCounterList, dto.CounterMetricDto{Name: k, Value: v})
	}
	for k, v := range gaugeMetricEntities {
		resGaugeList = append(resGaugeList, dto.GaugeMetricDto{Name: k, Value: v})
	}

	return resCounterList, resGaugeList
}
