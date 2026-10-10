package service

import (
	"context"

	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/constants"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/dto"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/repository"
)

type MetricService interface {
	UpdateMetric(ctx context.Context, dto *dto.MetricDto)
	GetMetric(ctx context.Context, metricType string, metricName string) (dto.MetricDto, error)
	GetAll(ctx context.Context) ([]dto.MetricDto, error)
}

type MetricServiceImpl struct {
	repo repository.MetricRepository
}

func NewMetricServiceImpl(repo repository.MetricRepository) *MetricServiceImpl {
	return &MetricServiceImpl{
		repo: repo,
	}
}

func (s *MetricServiceImpl) UpdateMetric(ctx context.Context, dto *dto.MetricDto) {
	if dto.MetricType == constants.GaugeMetricType {
		s.repo.UpdateGauge(ctx, dto.MetricName, dto.Gauge)
	} else if dto.MetricType == constants.CounterMetricType {
		s.repo.UpdateCounter(ctx, dto.MetricName, dto.Counter)
	}
}

func (s *MetricServiceImpl) GetMetric(ctx context.Context, metricType string, metricName string) (dto.MetricDto, error) {

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

func (s *MetricServiceImpl) GetAll(ctx context.Context) ([]dto.MetricDto, error) {
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
