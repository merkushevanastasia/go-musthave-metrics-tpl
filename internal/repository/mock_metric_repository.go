package repository

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type MockMetricRepository struct {
	mock.Mock
}

func (m *MockMetricRepository) UpdateGauge(ctx context.Context, name string, value float64) {
	m.Called(ctx, name, value)
}

func (m *MockMetricRepository) UpdateCounter(ctx context.Context, name string, value int64) {
	m.Called(ctx, name, value)
}

func (m *MockMetricRepository) GetGauge(ctx context.Context, name string) (float64, error) {
	args := m.Called(ctx, name)
	// args.Get(0) возвращает float64, args.Error(1) возвращает ошибку
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockMetricRepository) GetCounter(ctx context.Context, name string) (int64, error) {
	args := m.Called(ctx, name)
	// args.Get(0) возвращает int64, args.Error(1) возвращает ошибку
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockMetricRepository) GetAll(ctx context.Context) (map[string]float64, map[string]int64) {
	args := m.Called(ctx)

	var gauges map[string]float64
	if args.Get(0) != nil {
		gauges = args.Get(0).(map[string]float64)
	}

	var counters map[string]int64
	if args.Get(1) != nil {
		counters = args.Get(1).(map[string]int64)
	}

	return gauges, counters
}
