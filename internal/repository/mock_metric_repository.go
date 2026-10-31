package repository

import (
	"context"

	entity "github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/model/entity"
	"github.com/stretchr/testify/mock"
)

// MockMetricRepository является моком для интерфейса MetricRepository
type MockMetricRepository struct {
	mock.Mock
}

// Update имитирует обновление метрики в репозитории
func (m *MockMetricRepository) Update(ctx context.Context, entity entity.MetricEntity) {
	m.Called(ctx, entity)
}

// Get GetMetric Get имитирует получение метрики по имени
func (m *MockMetricRepository) Get(ctx context.Context, metricName string) (entity.MetricEntity, error) {
	args := m.Called(ctx, metricName)

	// Возвращаем объект и ошибку, которые мы настроим в тесте
	return args.Get(0).(entity.MetricEntity), args.Error(1)
}

// GetAll имитирует получение всех метрик
func (m *MockMetricRepository) GetAll(ctx context.Context) map[string]entity.MetricEntity {
	args := m.Called(ctx)

	// Безопасное приведение типа для map
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(map[string]entity.MetricEntity)
}
