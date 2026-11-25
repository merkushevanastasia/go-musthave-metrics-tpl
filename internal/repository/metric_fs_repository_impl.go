package repository

import (
	"context"
	"encoding/json"
	"errors"
	"maps"
	"os"
	"path/filepath"

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

func (repository *MetricFSRepositoryImpl) Update(ctx context.Context, metricEntity entity.MetricEntity) error {
	repository.Storage.Mu.Lock()
	defer repository.Storage.Mu.Unlock()

	repository.update(ctx, metricEntity)
	return nil
}

func (repository *MetricFSRepositoryImpl) update(ctx context.Context, metricEntity entity.MetricEntity) {
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

func (repository *MetricFSRepositoryImpl) Get(ctx context.Context, metricName string) (entity.MetricEntity, error) {
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
	if !repository.Config.Restore || repository.Config.FileStoragePath == "" {
		return nil
	}
	log.Info().Msg("Начинаем восстановление из файла: " + repository.Config.FileStoragePath)

	data, err := os.ReadFile(repository.Config.FileStoragePath)
	if errors.Is(err, os.ErrNotExist) {
		log.Info().Msg("Файл хранения метрик отсутствует, загрузка пропущена")
		return nil
	}
	if err != nil {
		return err
	}
	if len(data) == 0 {
		log.Info().Msg("Файл хранения метрик пуст, загрузка пропущена")
		return nil
	}

	var items []entity.MetricEntity
	if err := json.Unmarshal(data, &items); err != nil {
		log.Error().Err(err).Msg("Ошибка десериализации метрик из файла")
		return err
	}

	log.Debug().Int("count", len(items)).Msg("Метрики прочитаны из файла")

	repository.Storage.Mu.Lock()
	defer repository.Storage.Mu.Unlock()

	repository.Storage.Metrics = make(map[string]entity.MetricEntity, len(items))
	for _, item := range items {
		repository.update(ctx, item)
	}
	return nil
}

func (repository *MetricFSRepositoryImpl) CreateDump(ctx context.Context) error {
	log := logs.GetLoggerFromContext(ctx)
	log.Info().Msg("Создаём дамп")
	if repository.Config.FileStoragePath == "" {
		return nil
	}

	metricsMap := repository.GetAll(ctx)

	items := make([]entity.MetricEntity, 0, len(metricsMap))
	for _, metric := range metricsMap {
		items = append(items, metric)
	}

	data, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return err
	}

	dir := filepath.Dir(repository.Config.FileStoragePath)
	tmpFile, err := os.CreateTemp(dir, "metrics-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmpFile.Name()

	if _, err := tmpFile.Write(data); err != nil {
		tmpFile.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpName)
		return err
	}

	return os.Rename(tmpName, repository.Config.FileStoragePath)
}

func (repository *MetricFSRepositoryImpl) UpdateBatch(ctx context.Context, entities []entity.MetricEntity) error {
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
