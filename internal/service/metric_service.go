package service

import (
	"context"

	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/constants"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/dto"
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

func (s *MetricService) UpdateMetric(ctx context.Context, dto *dto.MetricDto) {
	if dto.MetricType == constants.GaugeMetricType {
		s.repo.UpdateGauge(ctx, dto.MetricName, dto.Gauge)
	} else if dto.MetricType == constants.CounterMetricType {
		s.repo.UpdateCounter(ctx, dto.MetricName, dto.Counter)
	}
}

func (s *MetricService) GetMetric(ctx context.Context, metricType string, metricName string) (dto.MetricDto, error) {

	result := dto.MetricDto{
		MetricName: metricName,
		MetricType: metricType,
	}

	if metricType == constants.GaugeMetricType {
		value, err := s.repo.GetGauge(ctx, metricName)
		if err != nil {
			return dto.MetricDto{}, err
		}
		result.Gauge = value
	}
	if metricType == constants.CounterMetricType {
		value, err := s.repo.GetCounter(ctx, metricName)
		if err != nil {
			return dto.MetricDto{}, err
		}
		result.Counter = value
	}
	return result, nil
}

func (s *MetricService) GetAll(ctx context.Context) ([]dto.MetricDto, error) {
	gaugeMetricEntities, counterMetricEntities := s.repo.GetAll(ctx)

	result := make([]dto.MetricDto, 0)

	for name, value := range counterMetricEntities {
		result = append(result, dto.MetricDto{MetricType: constants.CounterMetricType, MetricName: name, Counter: value})
	}

	for name, value := range gaugeMetricEntities {
		result = append(result, dto.MetricDto{MetricType: constants.GaugeMetricType, MetricName: name, Gauge: value})
	}

	return result, nil
}
