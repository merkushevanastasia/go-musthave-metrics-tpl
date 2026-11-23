package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

	log.Debug().Any("Содержимое файла", items).Msg("Содержимое")

	repository.Clean(ctx)
	for _, item := range items {
		repository.Update(ctx, item)
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
