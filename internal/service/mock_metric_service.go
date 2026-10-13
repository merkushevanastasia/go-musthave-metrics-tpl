package service

import (
	"context"

	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/dto"
)

type MockMetricService struct {
	UpdateMetricFunc func(ctx context.Context, dto *dto.MetricDto)
	GetMetricFunc    func(ctx context.Context, metricType string, metricName string) (dto.MetricDto, error)
	GetAllFunc       func(ctx context.Context) ([]dto.MetricDto, error)
}

func (m *MockMetricService) UpdateMetric(ctx context.Context, dto *dto.MetricDto) {
	if m.UpdateMetricFunc != nil {
		m.UpdateMetricFunc(ctx, dto)
	}
}

func (m *MockMetricService) GetMetric(ctx context.Context, metricType string, metricName string) (dto.MetricDto, error) {
	if m.GetMetricFunc != nil {
		return m.GetMetricFunc(ctx, metricType, metricName)
	}
	return dto.MetricDto{}, nil
}

func (m *MockMetricService) GetAll(ctx context.Context) ([]dto.MetricDto, error) {
	if m.GetAllFunc != nil {
		return m.GetAllFunc(ctx)
	}
	return nil, nil
}
