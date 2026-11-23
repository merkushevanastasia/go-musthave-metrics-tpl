package repository

import (
	"context"
	"fmt"
	"maps"

	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/constants"
	customerrors "github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/error"
	entity "github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/model/entity"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/pkg/logs"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MetricDBRepositoryImpl struct {
	Storage MemStorage
	DB      *pgxpool.Pool
}

func NewMetricDBRepositoryImpl(dbPool *pgxpool.Pool) *MetricDBRepositoryImpl {
	return &MetricDBRepositoryImpl{
		DB: dbPool,
		Storage: MemStorage{
			Metrics: make(map[string]entity.MetricEntity),
		},
	}
}

func (repository *MetricDBRepositoryImpl) Update(ctx context.Context, metricEntity entity.MetricEntity) {
	log := logs.GetLoggerFromContext(ctx)

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

func (repository *MetricDBRepositoryImpl) Get(ctx context.Context, metricName string) (entity.MetricEntity, error) {
	log := logs.GetLoggerFromContext(ctx)
	log.Debug().Str("имя метрики", metricName).Msg("Запрос значения метрики из хранилища")
	log.Debug().Any("memStorage", &repository.Storage).Msg("Хранилище метрик ")

	repository.Storage.Mu.RLock()
	defer repository.Storage.Mu.RUnlock()
	currentValue, isPresent := repository.Storage.Metrics[metricName]
	if isPresent {
		return currentValue, nil
	}

	log.Debug().Msg(fmt.Sprintf("Такой метрики не существует %s", metricName))
	return currentValue, customerrors.ErrMetricNotFound
}

func (repository *MetricDBRepositoryImpl) GetAll(ctx context.Context) map[string]entity.MetricEntity {
	log := logs.GetLoggerFromContext(ctx)
	log.Debug().Msg("Получение всех метрик из хранилища")

	repository.Storage.Mu.RLock()
	defer repository.Storage.Mu.RUnlock()

	copyMap := make(map[string]entity.MetricEntity, len(repository.Storage.Metrics))
	maps.Copy(copyMap, repository.Storage.Metrics)

	return copyMap
}

func (repository *MetricDBRepositoryImpl) Clean(ctx context.Context) {
	log := logs.GetLoggerFromContext(ctx)
	log.Debug().Msg("Очищаем")
	repository.Storage.Mu.Lock()
	defer repository.Storage.Mu.Unlock()
	repository.Storage.Metrics = make(map[string]entity.MetricEntity)
}

func (repository *MetricDBRepositoryImpl) RestoreMetrics(ctx context.Context) error {
	log := logs.GetLoggerFromContext(ctx)
	log.Info().Msg("Для данного репозитория метод RestoreMetrics пока не реализован")
	return nil
}

func (repository *MetricDBRepositoryImpl) CreateDump(ctx context.Context) error {
	log := logs.GetLoggerFromContext(ctx)
	log.Info().Msg("Для данного репозитория метод CreateDump не реализован")
	return nil
}
