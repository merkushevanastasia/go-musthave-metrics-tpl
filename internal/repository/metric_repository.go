package repository

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/utils"
)

type MemStorage struct {
	Gauge   map[string]float64
	Counter map[string]int64
}

type MetricRepository interface {
	UpdateGauge(ctx context.Context, name string, value float64)
	UpdateCounter(ctx context.Context, name string, value int64)
}

type MetricRepositoryImpl struct{}

var memStorage = MemStorage{
	Gauge:   make(map[string]float64),
	Counter: make(map[string]int64),
}

func (MetricRepositoryImpl) UpdateGauge(ctx context.Context, metricName string, value float64) {

	log := utils.FromContext(ctx)
	log.Debug("Хранилище метрик перед обновлением.", slog.Any("memStorage", memStorage))
	// todo потом убрать лишний вызов получения ради лога
	currentValue, isPresent := memStorage.Gauge[metricName]
	if isPresent {
		log.Debug(fmt.Sprintf("Метрика уже есть в системе. Значение метрики %s: %f. Обновляем на %f", metricName, currentValue, value))
	} else {
		log.Debug(fmt.Sprintf("Ранее метрики не было. Создаем новую метрику со значением %s: %f", metricName, value))
	}
	memStorage.Gauge[metricName] = value
	log.Debug("Хранилище метрик послк обновления.", slog.Any("memStorage", memStorage))

}
func (MetricRepositoryImpl) UpdateCounter(ctx context.Context, metricName string, value int64) {
	log := utils.FromContext(ctx)
	log.Debug("Хранилище метрик перед обновлением.", slog.Any("memStorage", memStorage))
	currentValue, isPresent := memStorage.Counter[metricName]

	if isPresent {
		log.Debug(fmt.Sprintf("Метрика уже есть в системе. Значение метрики %s: %d. Обновляем на %d + %d = %d", metricName, currentValue, currentValue, value, currentValue+value))
		memStorage.Counter[metricName] += value
	} else {
		log.Debug(fmt.Sprintf("Ранее метрики не было. Создаем новую метрику со значением %s: %d", metricName, value))
		memStorage.Counter[metricName] = value
	}
	log.Debug("Хранилище метрик послк обновления.", slog.Any("memStorage", memStorage))

}
