package service

import (
	"context"
	"testing"

	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/constants"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/mapper"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/model/dto"
	dto2 "github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/model/entity"
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
				ID:    "Alloc",
				Value: floatPtr(12345.67),
				MType: constants.GaugeMetricType,
			},
		},
	}

	for _, tt := range tests {
		ctx := context.Background()

		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(repository.MockMetricRepository)
			s := NewMetricServiceImpl(mockRepo)
			entity := mapper.MapToEntity(tt.inputDto)
			mockRepo.On("Update", ctx, entity).Return().Once()
			s.UpdateMetric(ctx, &tt.inputDto)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestMetricServiceGet(t *testing.T) {

	tests := []struct {
		name         string
		metricEntity dto2.MetricEntity
	}{
		{
			name: "Успешная обработка Gauge Alloc",
			metricEntity: dto2.MetricEntity{
				ID:    constants.Alloc,
				MType: constants.GaugeMetricType,
				Value: floatPtr(12345.67),
			},
		},
		{
			name: "Успешная обработка Gauge Alloc",
			metricEntity: dto2.MetricEntity{
				ID:    constants.RandomValue,
				MType: constants.CounterMetricType,
				Delta: intPtr(0),
			},
		},
	}

	for _, tt := range tests {
		ctx := context.Background()

		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(repository.MockMetricRepository)
			s := NewMetricServiceImpl(mockRepo)
			mockRepo.On("Get", ctx, tt.metricEntity.ID).Return(tt.metricEntity, nil)
			metric, err := s.GetMetric(ctx, tt.metricEntity.ID)
			if err != nil {
				return
			}

			assert.Equal(t, tt.metricEntity.MType, metric.MType)
			assert.Equal(t, tt.metricEntity.Value, metric.Value)
			assert.Equal(t, tt.metricEntity.Delta, metric.Delta)
			assert.Equal(t, tt.metricEntity.ID, metric.ID)
			mockRepo.AssertExpectations(t)
		})
	}
}
func floatPtr(v float64) *float64 {
	return new(v)
}
func intPtr(v int64) *int64 {
	return new(v)
}
