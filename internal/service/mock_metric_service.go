package service

import (
	"context"

	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/model/dto"
	"github.com/stretchr/testify/mock"
)

type MockMetricService struct {
	mock.Mock
	UpdateMetricFunc func(ctx context.Context, dto *dto.MetricDto)
	GetMetricFunc    func(ctx context.Context, metricName string) (dto.MetricDto, error)
	GetAllFunc       func(ctx context.Context) []dto.MetricDto
}

func (m *MockMetricService) UpdateMetric(ctx context.Context, dto *dto.MetricDto) {
	if m.UpdateMetricFunc != nil {
		m.UpdateMetricFunc(ctx, dto)
	}
}

func (m *MockMetricService) GetMetric(ctx context.Context, metricName string) (dto.MetricDto, error) {
	if m.GetMetricFunc != nil {
		return m.GetMetricFunc(ctx, metricName)
	}
	return dto.MetricDto{}, nil
}

func (m *MockMetricService) StartStoreTicker(ctx context.Context) {
	m.Called(ctx)
}

// RestoreMetrics имитирует восстановление метрик при старте
func (m *MockMetricService) RestoreMetrics() error {
	args := m.Called()
	return args.Error(0)
}
