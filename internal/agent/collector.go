package agent

import (
	"math/rand/v2"
	"runtime"

	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/constants"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/model/dto"
	"github.com/rs/zerolog/log"
)

func Update(metricCollection *dto.MetricCollection) {

	metricCollection.Mu.Lock()
	defer metricCollection.Mu.Unlock()

	log.Debug().Any("До обновления ", metricCollection.Metrics).Msg("Запускаем процесс сбора метрик.")
	log.Info().Msg("Собираем метрики...")

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	metricCollection.Add(toEntityGauge(constants.Alloc, float64(m.Alloc)))
	metricCollection.Add(toEntityGauge(constants.BuckHashSys, float64(m.BuckHashSys)))
	metricCollection.Add(toEntityGauge(constants.Frees, float64(m.Frees)))
	metricCollection.Add(toEntityGauge(constants.GCCPUFraction, m.GCCPUFraction))
	metricCollection.Add(toEntityGauge(constants.GCSys, float64(m.GCSys)))
	metricCollection.Add(toEntityGauge(constants.HeapAlloc, float64(m.HeapAlloc)))
	metricCollection.Add(toEntityGauge(constants.HeapIdle, float64(m.HeapIdle)))
	metricCollection.Add(toEntityGauge(constants.HeapInuse, float64(m.HeapInuse)))
	metricCollection.Add(toEntityGauge(constants.HeapObjects, float64(m.HeapObjects)))
	metricCollection.Add(toEntityGauge(constants.HeapReleased, float64(m.HeapReleased)))
	metricCollection.Add(toEntityGauge(constants.HeapSys, float64(m.HeapSys)))
	metricCollection.Add(toEntityGauge(constants.LastGC, float64(m.LastGC)))
	metricCollection.Add(toEntityGauge(constants.Lookups, float64(m.Lookups)))
	metricCollection.Add(toEntityGauge(constants.MCacheInuse, float64(m.MCacheInuse)))
	metricCollection.Add(toEntityGauge(constants.MCacheSys, float64(m.MCacheSys)))
	metricCollection.Add(toEntityGauge(constants.MSpanInuse, float64(m.MSpanInuse)))
	metricCollection.Add(toEntityGauge(constants.MSpanSys, float64(m.MSpanSys)))
	metricCollection.Add(toEntityGauge(constants.Mallocs, float64(m.Mallocs)))
	metricCollection.Add(toEntityGauge(constants.NextGC, float64(m.NextGC)))
	metricCollection.Add(toEntityGauge(constants.NumForcedGC, float64(m.NumForcedGC)))
	metricCollection.Add(toEntityGauge(constants.NumGC, float64(m.NumGC)))
	metricCollection.Add(toEntityGauge(constants.OtherSys, float64(m.OtherSys)))
	metricCollection.Add(toEntityGauge(constants.PauseTotalNs, float64(m.PauseTotalNs)))
	metricCollection.Add(toEntityGauge(constants.StackInuse, float64(m.StackInuse)))
	metricCollection.Add(toEntityGauge(constants.StackSys, float64(m.StackSys)))
	metricCollection.Add(toEntityGauge(constants.Sys, float64(m.Sys)))
	metricCollection.Add(toEntityGauge(constants.TotalAlloc, float64(m.TotalAlloc)))
	metricCollection.Add(toEntityGauge(constants.RandomValue, rand.Float64()))

	metricCollection.Add(toEntityCounter(constants.PollCount, metricCollection.GetCurrent(constants.PollCount)+1))

	log.Debug().Any("После обновления ", metricCollection.Metrics).Msg("Собрали и обновили метрики.")
	log.Info().Msg("Собрали и обновили метрики.")
}

func toEntityGauge(metricName string, value float64) dto.MetricDto {
	return dto.MetricDto{ID: metricName, Value: &value, MType: constants.GaugeMetricType}
}
func toEntityCounter(metricName string, delta int64) dto.MetricDto {
	return dto.MetricDto{ID: metricName, Delta: &delta, MType: constants.CounterMetricType}
}
