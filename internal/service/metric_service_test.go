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
