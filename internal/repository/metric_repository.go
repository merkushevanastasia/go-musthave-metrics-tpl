package repository

import (
	"context"
	"fmt"
	"log/slog"

	customerrors "github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/error"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/utils"
)

// MemStorage todo пока что оставлю раздельными, возможно на блоке по БД объединить Gauge и Counter в общую таблицу
type MemStorage struct {
	Gauge   map[string]float64
	Counter map[string]int64
}

type MetricRepository interface {
	UpdateGauge(ctx context.Context, name string, value float64)
	UpdateCounter(ctx context.Context, name string, value int64)
	GetGauge(ctx context.Context, name string) (float64, error)
	GetCounter(ctx context.Context, name string) (int64, error)
	GetAll(ctx context.Context) (map[string]float64, map[string]int64)
}

type MetricRepositoryImpl struct{}

var memStorage = MemStorage{
	Gauge:   make(map[string]float64),
	Counter: make(map[string]int64),
}

func (MetricRepositoryImpl) UpdateGauge(ctx context.Context, metricName string, value float64) {

	log := utils.FromContext(ctx)
	log.Debug("Хранилище метрик перед обновлением.", slog.Any("memStorage", memStorage))
	memStorage.Gauge[metricName] = value
	log.Debug("Хранилище метрик послк обновления.", slog.Any("memStorage", memStorage))

}
func (MetricRepositoryImpl) UpdateCounter(ctx context.Context, metricName string, value int64) {
	log := utils.FromContext(ctx)
	log.Debug("Хранилище метрик перед обновлением.", slog.Any("memStorage", memStorage))
	_, isPresent := memStorage.Counter[metricName]

	if isPresent {
		memStorage.Counter[metricName] += value
	} else {
		memStorage.Counter[metricName] = value
	}
	log.Debug("Хранилище метрик послк обновления.", slog.Any("memStorage", memStorage))

}

func (MetricRepositoryImpl) GetGauge(ctx context.Context, metricName string) (float64, error) {
	log := utils.FromContext(ctx)
	log.Debug("Запрос значения метрики из хранилища", slog.Any("имя метрики", metricName))
	currentValue, isPresent := memStorage.Gauge[metricName]
	if isPresent {
		return currentValue, nil
	}

	log.Debug(fmt.Sprintf("Такой метрики не существует %s", metricName))
	return currentValue, customerrors.ErrMetricNotFound
}

func (MetricRepositoryImpl) GetCounter(ctx context.Context, metricName string) (int64, error) {
	log := utils.FromContext(ctx)
	log.Debug("Запрос значения метрики из хранилища", slog.Any("имя метрики", metricName))
	currentValue, isPresent := memStorage.Counter[metricName]

	if isPresent {
		return currentValue, nil
	}

	log.Debug(fmt.Sprintf("Такой метрики не существует %s", metricName))
	return currentValue, customerrors.ErrMetricNotFound
}

func (MetricRepositoryImpl) GetAll(ctx context.Context) (map[string]float64, map[string]int64) {
	log := utils.FromContext(ctx)
	log.Debug("Получение всех метрик из хранилища")
	return memStorage.Gauge, memStorage.Counter
}
