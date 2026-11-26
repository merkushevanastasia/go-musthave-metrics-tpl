package repository

import (
	"context"
	"maps"
	"sync"

	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/constants"
	customerrors "github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/error"
	entity "github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/model/entity"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/pkg/logs"
)

type MemStorage struct {
	Mu      sync.RWMutex
	Metrics map[string]entity.MetricEntity
}

type MetricMemoryRepositoryImpl struct {
	Storage MemStorage
}

func NewMetricMemoryRepositoryImpl() *MetricMemoryRepositoryImpl {
	return &MetricMemoryRepositoryImpl{
		Storage: MemStorage{
			Metrics: make(map[string]entity.MetricEntity),
		},
	}
}

func (repository *MetricMemoryRepositoryImpl) Update(ctx context.Context, metricEntity entity.MetricEntity) error {
	repository.Storage.Mu.Lock()
	defer repository.Storage.Mu.Unlock()

	repository.update(ctx, metricEntity)
	return nil
}

func (repository *MetricMemoryRepositoryImpl) update(ctx context.Context, metricEntity entity.MetricEntity) {
	log := logs.GetLoggerFromContext(ctx)

	log.Debug().
		Str("id", metricEntity.ID).
		Str("type", metricEntity.MType).
		Msg("Обновление метрики в хранилище")

	if metricEntity.MType == constants.CounterMetricType {
		current, isPresent := repository.Storage.Metrics[metricEntity.ID]
		if isPresent {
			var newDelta int64
			if current.Delta != nil {
				newDelta += *current.Delta
			}
			if metricEntity.Delta != nil {
				newDelta += *metricEntity.Delta
			}
			metricEntity.Delta = &newDelta
		}
	}
	repository.Storage.Metrics[metricEntity.ID] = metricEntity

	log.Debug().
		Str("id", metricEntity.ID).
		Int("total_metrics", len(repository.Storage.Metrics)).
		Msg("Метрика обновлена в хранилище")
}

func (repository *MetricMemoryRepositoryImpl) Get(ctx context.Context, metricName string) (entity.MetricEntity, error) {
	log := logs.GetLoggerFromContext(ctx)
	log.Debug().Str("имя метрики", metricName).Msg("Запрос значения метрики из хранилища")

	repository.Storage.Mu.RLock()
	defer repository.Storage.Mu.RUnlock()

	currentValue, isPresent := repository.Storage.Metrics[metricName]
	if isPresent {
		return currentValue, nil
	}

	log.Debug().Str("имя метрики", metricName).Msg("Такой метрики не существует")
	return currentValue, customerrors.ErrMetricNotFound
}

func (repository *MetricMemoryRepositoryImpl) GetAll(ctx context.Context) map[string]entity.MetricEntity {
	log := logs.GetLoggerFromContext(ctx)
	log.Debug().Msg("Получение всех метрик из хранилища")

	repository.Storage.Mu.RLock()
	defer repository.Storage.Mu.RUnlock()

	copyMap := make(map[string]entity.MetricEntity, len(repository.Storage.Metrics))
	maps.Copy(copyMap, repository.Storage.Metrics)

	return copyMap
}

func (repository *MetricMemoryRepositoryImpl) Clean(ctx context.Context) {
	log := logs.GetLoggerFromContext(ctx)
	log.Debug().Msg("Очищаем")
	repository.Storage.Mu.Lock()
	defer repository.Storage.Mu.Unlock()
	repository.Storage.Metrics = make(map[string]entity.MetricEntity)
}

func (repository *MetricMemoryRepositoryImpl) CreateDump(ctx context.Context) error {
	log := logs.GetLoggerFromContext(ctx)
	log.Info().Msg("Для данного репозитория метод CreateDump не реализован")
	return nil
}

func (repository *MetricMemoryRepositoryImpl) RestoreMetrics(ctx context.Context) error {
	log := logs.GetLoggerFromContext(ctx)
	log.Info().Msg("Для данного репозитория метод RestoreMetrics пока не реализован")
	return nil
}

func (repository *MetricMemoryRepositoryImpl) UpdateBatch(ctx context.Context, entities []entity.MetricEntity) error {
	if len(entities) == 0 {
		return nil
	}

	repository.Storage.Mu.Lock()
	defer repository.Storage.Mu.Unlock()

	for _, item := range entities {
		repository.update(ctx, item)
	}
	return nil
}
