package repository

import (
	"context"

	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/model/entity"
)

type MetricRepository interface {
	Update(ctx context.Context, entity entity.MetricEntity) error
	Get(ctx context.Context, metricName string) (entity.MetricEntity, error)
	GetAll(ctx context.Context) map[string]entity.MetricEntity
	Clean(ctx context.Context)
	RestoreMetrics(ctx context.Context) error
	CreateDump(ctx context.Context) error
	UpdateBatch(ctx context.Context, entities []entity.MetricEntity) error
}
