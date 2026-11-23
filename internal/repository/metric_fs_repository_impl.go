package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"os"

	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/config/server"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/constants"
	customerrors "github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/error"
	entity "github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/model/entity"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/pkg/logs"
)

type MetricFSRepositoryImpl struct {
	Storage MemStorage
	Config  server.Config
}

func NewMetricFSRepositoryImpl(config server.Config) *MetricFSRepositoryImpl {
	return &MetricFSRepositoryImpl{
		Config: config,
		Storage: MemStorage{
			Metrics: make(map[string]entity.MetricEntity),
		},
	}
}

func (repository *MetricFSRepositoryImpl) Update(ctx context.Context, metricEntity entity.MetricEntity) {
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

func (repository *MetricFSRepositoryImpl) Get(ctx context.Context, metricName string) (entity.MetricEntity, error) {
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

func (repository *MetricFSRepositoryImpl) GetAll(ctx context.Context) map[string]entity.MetricEntity {
	log := logs.GetLoggerFromContext(ctx)
	log.Debug().Msg("Получение всех метрик из хранилища")

	repository.Storage.Mu.RLock()
	defer repository.Storage.Mu.RUnlock()

	copyMap := make(map[string]entity.MetricEntity, len(repository.Storage.Metrics))
	maps.Copy(copyMap, repository.Storage.Metrics)

	return copyMap
}

func (repository *MetricFSRepositoryImpl) Clean(ctx context.Context) {
	log := logs.GetLoggerFromContext(ctx)
	log.Debug().Msg("Очищаем")
	repository.Storage.Mu.Lock()
	defer repository.Storage.Mu.Unlock()
	repository.Storage.Metrics = make(map[string]entity.MetricEntity)
}

func (repository *MetricFSRepositoryImpl) RestoreMetrics(ctx context.Context) error {
	log := logs.GetLoggerFromContext(ctx)
	if !repository.Config.Restore {
		return nil
	}
	log.Info().Msg("Начинаем восстановление из файла" + repository.Config.FileStoragePath)

	data, err := os.ReadFile(repository.Config.FileStoragePath)
	if os.IsNotExist(err) {
		log.Info().Msg("Файл хранения метрик отсутствует, загрузка пропущена")
		return err

	}
	if err != nil {
		return err

	}

	var items []entity.MetricEntity
	if err := json.Unmarshal(data, &items); err != nil {
		log.Error().Err(err)
		return err
	}

	log.Debug().Any("Содержимое файла", items).Msg("Содержимое")

	repository.Clean(context.Background())
	for _, item := range items {
		repository.Update(context.Background(), item)
	}
	return nil
}

func (repository *MetricFSRepositoryImpl) CreateDump(ctx context.Context) error {
	log := logs.GetLoggerFromContext(ctx)
	log.Info().Msg("Создаём дамп")
	if repository.Config.FileStoragePath == "" {
		return nil
	}

	metricsMap := repository.GetAll(context.Background())

	var items []entity.MetricEntity
	for _, metric := range metricsMap {
		items = append(items, metric)
	}

	data, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return err
	}
	file, err := os.OpenFile(repository.Config.FileStoragePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(data)
	return err
}
