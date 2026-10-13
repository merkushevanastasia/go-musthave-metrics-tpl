package service

import (
	"context"
	"testing"

	dto "github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/dto/server"
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

func TestMetricServiceProcessGauge(t *testing.T) {

	tests := []struct {
		name     string
		inputDto dto.GaugeMetricDto
	}{
		{
			name: "Успешная обработка Gauge Alloc",
			inputDto: dto.GaugeMetricDto{
				Name:  "Alloc",
				Value: 12345.67,
			},
		},
	}

	for _, tt := range tests {
		ctx := context.Background()

		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockMetricRepository)
			s := NewMetricService(mockRepo)
			mockRepo.On("UpdateGauge", ctx, tt.inputDto.Name, tt.inputDto.Value).Return().Once()
			s.ProcessGauge(ctx, tt.inputDto)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestMetricServiceProcessCounter(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		inputDto dto.CounterMetricDto
	}{
		{
			name: "Успешная обработка Counter PollCount",
			inputDto: dto.CounterMetricDto{
				Name:  "PollCount",
				Value: 5,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockMetricRepository)
			s := NewMetricService(mockRepo)
			mockRepo.On("UpdateCounter", ctx, tt.inputDto.Name, tt.inputDto.Value).Return().Once()
			s.ProcessCounter(ctx, tt.inputDto)
			mockRepo.AssertExpectations(t)
		})
	}
}
