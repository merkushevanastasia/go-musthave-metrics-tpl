package service

import (
	"context"
	"testing"

	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/constants"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/dto"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/repository"
	"github.com/stretchr/testify/assert"
)

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
			mockRepo := new(repository.MockMetricRepository)
			s := NewMetricServiceImpl(mockRepo)
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
			mockRepo := new(repository.MockMetricRepository)
			s := NewMetricServiceImpl(mockRepo)
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
