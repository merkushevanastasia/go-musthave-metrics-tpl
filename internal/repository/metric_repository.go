package repository

import (
	"context"
	"fmt"
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
	log.Debug().Any("memStorage", &repository.Storage).Msg("Хранилище метрик перед обновлением")

	if metricEntity.MType == constants.GaugeMetricType {
		current, isPresent := repository.Storage.Metrics[metricEntity.ID]
		if isPresent {
			newDelta := *metricEntity.Delta + *current.Delta
			metricEntity.Delta = &newDelta
		}
	}
	repository.Storage.Metrics[metricEntity.ID] = metricEntity

	repository.Storage.Metrics[metricEntity.ID] = metricEntity
	log.Debug().Any("memStorage", &repository.Storage).Msg("Хранилище метрик послк обновления")

}

func (repository *MetricRepositoryImpl) Get(ctx context.Context, metricName string) (entity.MetricEntity, error) {
	log := logging.FromContext(ctx)
	log.Debug().Str("имя метрики", metricName).Msg("Запрос значения метрики из хранилища")
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
	return repository.Storage.Metrics
}
