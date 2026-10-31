package repository

import (
	"context"
	"fmt"

	customerrors "github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/error"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/utils/logging"
	"github.com/rs/zerolog"
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

type MetricRepositoryImpl struct {
	Storage MemStorage
}

// MarshalZerologObject Реализация интерфейса zerolog.LogObjectMarshaler
func (s *MemStorage) MarshalZerologObject(e *zerolog.Event) {
	if len(s.Gauge) > 0 {
		gaugeDict := e.CreateDict()
		for k, v := range s.Gauge {
			gaugeDict.Float64(k, v)
		}
		e.Dict("gauge", gaugeDict)
	} else {
		e.Interface("gauge", struct{}{})
	}
	if len(s.Counter) > 0 {
		counterDict := e.CreateDict()
		for k, v := range s.Counter {
			counterDict.Int64(k, v)
		}
		e.Dict("counter", counterDict)
	} else {
		e.Interface("counter", struct{}{})
	}
}

func NewMetricRepositoryImpl() *MetricRepositoryImpl {
	return &MetricRepositoryImpl{
		Storage: MemStorage{
			Gauge:   make(map[string]float64),
			Counter: make(map[string]int64),
		},
	}
}

func (repository MetricRepositoryImpl) UpdateGauge(ctx context.Context, metricName string, value float64) {

	log := logging.FromContext(ctx)
	log.Debug().Object("memStorage", &repository.Storage).Msg("Хранилище метрик перед обновлением")
	repository.Storage.Gauge[metricName] = value
	log.Debug().Object("memStorage", &repository.Storage).Msg("Хранилище метрик послк обновления")

}
func (repository MetricRepositoryImpl) UpdateCounter(ctx context.Context, metricName string, value int64) {
	log := logging.FromContext(ctx)
	log.Debug().Object("memStorage", &repository.Storage).Msg("Хранилище метрик перед обновлением")
	_, isPresent := repository.Storage.Counter[metricName]

	if isPresent {
		repository.Storage.Counter[metricName] += value
	} else {
		repository.Storage.Counter[metricName] = value
	}
	log.Debug().Object("memStorage", &repository.Storage).Msg("Хранилище метрик послк обновления")

}

func (repository MetricRepositoryImpl) GetGauge(ctx context.Context, metricName string) (float64, error) {
	log := logging.FromContext(ctx)
	log.Debug().Str("имя метрики", metricName).Msg("Запрос значения метрики из хранилища")
	currentValue, isPresent := repository.Storage.Gauge[metricName]
	if isPresent {
		return currentValue, nil
	}

	log.Debug().Str("metric_name", metricName).Msg("Такой метрики не существует")
	return currentValue, customerrors.ErrMetricNotFound
}

func (repository MetricRepositoryImpl) GetCounter(ctx context.Context, metricName string) (int64, error) {
	log := logging.FromContext(ctx)
	log.Debug().Str("имя метрики", metricName).Msg("Запрос значения метрики из хранилища")
	currentValue, isPresent := repository.Storage.Counter[metricName]

	if isPresent {
		return currentValue, nil
	}

	log.Debug().Msg(fmt.Sprintf("Такой метрики не существует %s", metricName))
	return currentValue, customerrors.ErrMetricNotFound
}

func (repository MetricRepositoryImpl) GetAll(ctx context.Context) (map[string]float64, map[string]int64) {
	log := logging.FromContext(ctx)
	log.Debug().Msg("Получение всех метрик из хранилища")
	return repository.Storage.Gauge, repository.Storage.Counter
}
