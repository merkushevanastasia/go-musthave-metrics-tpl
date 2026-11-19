package repository

import (
	"context"
	"fmt"
	"maps"
	"sync"

	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/constants"
	customerrors "github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/error"
	entity "github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/model/entity"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/utils/logging"
)

type MemStorage struct {
	Mu      sync.RWMutex
	Metrics map[string]entity.MetricEntity
}

type MetricRepository interface {
	Update(ctx context.Context, entity entity.MetricEntity)
	Get(ctx context.Context, metricName string) (entity.MetricEntity, error)
	GetAll(ctx context.Context) map[string]entity.MetricEntity
	Clean(ctx context.Context)
}

type MetricRepositoryImpl struct {
	Storage MemStorage
}

func NewMetricRepositoryImpl() *MetricRepositoryImpl {
	return &MetricRepositoryImpl{
		Storage: MemStorage{
			Metrics: make(map[string]entity.MetricEntity),
		},
	}
}

func (repository *MetricRepositoryImpl) Update(ctx context.Context, metricEntity entity.MetricEntity) {
	log := logging.FromContext(ctx)

	repository.Storage.Mu.Lock()
	defer repository.Storage.Mu.Unlock()

	log.Debug().Any("memStorage", &repository.Storage).Msg("Хранилище метрик перед обновлением")

	if metricEntity.MType == constants.CounterMetricType {
		current, isPresent := repository.Storage.Metrics[metricEntity.ID]
		if isPresent {
			var newDelta int64 = 0
			if current.Delta != nil {
				newDelta = newDelta + *current.Delta
			}
			if metricEntity.Delta != nil {
				newDelta = newDelta + *metricEntity.Delta
			}
			metricEntity.Delta = &newDelta
		}
	}
	repository.Storage.Metrics[metricEntity.ID] = metricEntity

	log.Debug().Any("memStorage", &repository.Storage).Msg("Хранилище метрик после обновления")
}

func (repository *MetricRepositoryImpl) Get(ctx context.Context, metricName string) (entity.MetricEntity, error) {
	log := logging.FromContext(ctx)
	log.Debug().Str("имя метрики", metricName).Msg("Запрос значения метрики из хранилища")

	repository.Storage.Mu.RLock()
	defer repository.Storage.Mu.RUnlock()
	currentValue, isPresent := repository.Storage.Metrics[metricName]

	if isPresent {
		return currentValue, nil
	}

	log.Debug().Msg(fmt.Sprintf("Такой метрики не существует %s", metricName))
	return currentValue, customerrors.ErrMetricNotFound
}

func (repository *MetricRepositoryImpl) GetAll(ctx context.Context) map[string]entity.MetricEntity {
	log := logging.FromContext(ctx)
	log.Debug().Msg("Получение всех метрик из хранилища")

	repository.Storage.Mu.RLock()
	defer repository.Storage.Mu.RUnlock()

	copyMap := make(map[string]entity.MetricEntity, len(repository.Storage.Metrics))
	maps.Copy(copyMap, repository.Storage.Metrics)

	return copyMap
}

func (repository *MetricRepositoryImpl) Clean(ctx context.Context) {
	log := logging.FromContext(ctx)
	log.Debug().Msg("Очищаем")
	repository.Storage.Mu.Lock()
	defer repository.Storage.Mu.Unlock()
	repository.Storage.Metrics = make(map[string]entity.MetricEntity)
}
