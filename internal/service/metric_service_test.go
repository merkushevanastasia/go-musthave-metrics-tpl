package service

import (
	"context"
	"testing"

	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/constants"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/dto"
	"github.com/stretchr/testify/assert"
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

	// Безопасное приведение типов для мап, так как они могут быть nil в настройках On().Return()
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

func TestMetricServiceUpdateMetric(t *testing.T) {

	tests := []struct {
		name     string
		inputDto dto.MetricDto
	}{
		{
			name: "Успешная обработка Gauge Alloc",
			inputDto: dto.MetricDto{
				MetricName: "Alloc",
				Gauge:      12345.67,
				MetricType: constants.GaugeMetricType,
			},
		},
	}

	for _, tt := range tests {
		ctx := context.Background()

		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockMetricRepository)
			s := NewMetricService(mockRepo)
			mockRepo.On("UpdateGauge", ctx, tt.inputDto.MetricName, tt.inputDto.Gauge).Return().Once()
			s.UpdateMetric(ctx, &tt.inputDto)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestMetricServiceGet(t *testing.T) {

	tests := []struct {
		name            string
		metricName      string
		metricType      string
		expectedCounter int64
		expectedGauge   float64
	}{
		{
			name:            "Успешная обработка Gauge Alloc",
			metricName:      constants.Alloc,
			metricType:      constants.GaugeMetricType,
			expectedCounter: 0,
			expectedGauge:   12345.67,
		},
		{
			name:            "Успешная обработка Gauge Alloc",
			metricName:      constants.RandomValue,
			metricType:      constants.CounterMetricType,
			expectedCounter: 100,
			expectedGauge:   0,
		},
	}

	for _, tt := range tests {
		ctx := context.Background()

		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockMetricRepository)
			s := NewMetricService(mockRepo)
			if tt.metricType == constants.GaugeMetricType {
				mockRepo.On("GetGauge", ctx, tt.metricName).Return(tt.expectedGauge, nil)
			} else {
				mockRepo.On("GetCounter", ctx, tt.metricName).Return(tt.expectedCounter, nil)
			}
			metric, err := s.GetMetric(ctx, tt.metricType, tt.metricName)
			if err != nil {
				return
			}

			assert.Equal(t, tt.metricType, metric.MetricType)
			assert.Equal(t, tt.expectedGauge, metric.Gauge)
			assert.Equal(t, tt.expectedCounter, metric.Counter)
			assert.Equal(t, tt.metricName, metric.MetricName)
			mockRepo.AssertExpectations(t)
		})
	}
}
