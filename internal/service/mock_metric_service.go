package service

import (
	"context"

	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/model/dto"
	"github.com/stretchr/testify/mock"
)

type MockMetricService struct {
	mock.Mock
}

// UpdateMetric имитирует обновление метрики
func (m *MockMetricService) UpdateMetric(ctx context.Context, dto *dto.MetricDto) {
	m.Called(ctx, dto)
}

// GetMetric имитирует получение метрики
func (m *MockMetricService) GetMetric(ctx context.Context, metricName string) (dto.MetricDto, error) {
	args := m.Called(ctx, metricName)
	return args.Get(0).(dto.MetricDto), args.Error(1)
}

// GetAll имитирует получение списка всех метрик (Был пропущен в старом моке)
func (m *MockMetricService) GetAll(ctx context.Context) []dto.MetricDto {
	args := m.Called(ctx)

	// Безопасно проверяем на nil, если в .Return() передали пустое значение
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]dto.MetricDto)
}

// StartStoreTicker имитирует запуск таймера сброса метрик на диск
func (m *MockMetricService) StartStoreTicker(ctx context.Context) {
	m.Called(ctx)
}

// RestoreMetrics имитирует восстановление метрик при старте
func (m *MockMetricService) RestoreMetrics() error {
	args := m.Called()
	return args.Error(0)
}
