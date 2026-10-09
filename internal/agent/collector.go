package agent

import (
	"log/slog"
	"math/rand/v2"
	"runtime"

	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/constants"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/dto"
)

func Update(metricCollection *dto.MetricCollection) {

	metricCollection.Mu.Lock()
	defer metricCollection.Mu.Unlock()

	slog.Debug("Запускаем процесс сбора метрик.", slog.Any("До обновления ", metricCollection.Metrics))
	slog.Info("Собираем метрики...")

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	metricCollection.AddGauge(constants.Alloc, float64(m.Alloc))
	metricCollection.AddGauge(constants.BuckHashSys, float64(m.BuckHashSys))
	metricCollection.AddGauge(constants.Frees, float64(m.Frees))
	metricCollection.AddGauge(constants.GCCPUFraction, float64(m.GCCPUFraction))
	metricCollection.AddGauge(constants.GCSys, float64(m.GCSys))
	metricCollection.AddGauge(constants.HeapAlloc, float64(m.HeapAlloc))
	metricCollection.AddGauge(constants.HeapIdle, float64(m.HeapIdle))
	metricCollection.AddGauge(constants.HeapInuse, float64(m.HeapInuse))
	metricCollection.AddGauge(constants.HeapObjects, float64(m.HeapObjects))
	metricCollection.AddGauge(constants.HeapReleased, float64(m.HeapReleased))
	metricCollection.AddGauge(constants.HeapSys, float64(m.HeapSys))
	metricCollection.AddGauge(constants.LastGC, float64(m.LastGC))
	metricCollection.AddGauge(constants.Lookups, float64(m.Lookups))
	metricCollection.AddGauge(constants.MCacheInuse, float64(m.MCacheInuse))
	metricCollection.AddGauge(constants.MCacheSys, float64(m.MCacheSys))
	metricCollection.AddGauge(constants.MSpanInuse, float64(m.MSpanInuse))
	metricCollection.AddGauge(constants.MSpanSys, float64(m.MSpanSys))
	metricCollection.AddGauge(constants.Mallocs, float64(m.Mallocs))
	metricCollection.AddGauge(constants.NextGC, float64(m.NextGC))
	metricCollection.AddGauge(constants.NumForcedGC, float64(m.NumForcedGC))
	metricCollection.AddGauge(constants.NumGC, float64(m.NumGC))
	metricCollection.AddGauge(constants.OtherSys, float64(m.OtherSys))
	metricCollection.AddGauge(constants.PauseTotalNs, float64(m.PauseTotalNs))
	metricCollection.AddGauge(constants.StackInuse, float64(m.StackInuse))
	metricCollection.AddGauge(constants.StackSys, float64(m.StackSys))
	metricCollection.AddGauge(constants.Sys, float64(m.Sys))
	metricCollection.AddGauge(constants.TotalAlloc, float64(m.TotalAlloc))

	metricCollection.AddCounter(constants.PollCount, 1)
	metricCollection.AddCounter(constants.RandomValue, rand.Int64N(10000))

	slog.Debug("Собрали и обновили метрики.", slog.Any("До обновления ", metricCollection.Metrics))
}
