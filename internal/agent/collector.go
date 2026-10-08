package agent

import (
	"log/slog"
	"math/rand/v2"
	"runtime"

	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/dto"
)

const Alloc = "Alloc"
const BuckHashSys = "BuckHashSys"
const Frees = "Frees"
const GCCPUFraction = "GCCPUFraction"
const GCSys = "GCSys"
const HeapAlloc = "HeapAlloc"
const HeapIdle = "HeapIdle"
const HeapInuse = "HeapInuse"
const HeapObjects = "HeapObjects"
const HeapReleased = "HeapReleased"
const HeapSys = "HeapSys"
const LastGC = "LastGC"
const Lookups = "Lookups"
const MCacheInuse = "MCacheInuse"
const MCacheSys = "MCacheSys"
const MSpanInuse = "MSpanInuse"
const MSpanSys = "MSpanSys"
const Mallocs = "Mallocs"
const NextGC = "NextGC"
const NumForcedGC = "NumForcedGC"
const NumGC = "NumGC"
const OtherSys = "OtherSys"
const PauseTotalNs = "PauseTotalNs"
const StackInuse = "StackInuse"
const StackSys = "StackSys"
const Sys = "Sys"
const TotalAlloc = "TotalAlloc"

const PollCount = "PollCount"
const RandomValue = "RandomValue"

func Update(metricCollection *dto.MetricCollection) {

	metricCollection.Mu.Lock()
	defer metricCollection.Mu.Unlock()

	slog.Info("Запускаем процесс сбора метрик.", slog.Any("До обновления ", metricCollection.Metrics))
	slog.Info("Собираем метрики...")

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	metricCollection.AddGauge(Alloc, float64(m.Alloc))
	metricCollection.AddGauge(BuckHashSys, float64(m.BuckHashSys))
	metricCollection.AddGauge(Frees, float64(m.Frees))
	metricCollection.AddGauge(GCCPUFraction, float64(m.GCCPUFraction))
	metricCollection.AddGauge(GCSys, float64(m.GCSys))
	metricCollection.AddGauge(HeapAlloc, float64(m.HeapAlloc))
	metricCollection.AddGauge(HeapIdle, float64(m.HeapIdle))
	metricCollection.AddGauge(HeapInuse, float64(m.HeapInuse))
	metricCollection.AddGauge(HeapObjects, float64(m.HeapObjects))
	metricCollection.AddGauge(HeapReleased, float64(m.HeapReleased))
	metricCollection.AddGauge(HeapSys, float64(m.HeapSys))
	metricCollection.AddGauge(LastGC, float64(m.LastGC))
	metricCollection.AddGauge(Lookups, float64(m.Lookups))
	metricCollection.AddGauge(MCacheInuse, float64(m.MCacheInuse))
	metricCollection.AddGauge(MCacheSys, float64(m.MCacheSys))
	metricCollection.AddGauge(MSpanInuse, float64(m.MSpanInuse))
	metricCollection.AddGauge(MSpanSys, float64(m.MSpanSys))
	metricCollection.AddGauge(Mallocs, float64(m.Mallocs))
	metricCollection.AddGauge(NextGC, float64(m.NextGC))
	metricCollection.AddGauge(NumForcedGC, float64(m.NumForcedGC))
	metricCollection.AddGauge(NumGC, float64(m.NumGC))
	metricCollection.AddGauge(OtherSys, float64(m.OtherSys))
	metricCollection.AddGauge(PauseTotalNs, float64(m.PauseTotalNs))
	metricCollection.AddGauge(StackInuse, float64(m.StackInuse))
	metricCollection.AddGauge(StackSys, float64(m.StackSys))
	metricCollection.AddGauge(Sys, float64(m.Sys))
	metricCollection.AddGauge(TotalAlloc, float64(m.TotalAlloc))

	metricCollection.AddCounter(PollCount, 1)
	metricCollection.AddCounter(RandomValue, rand.Int64N(10000))

	slog.Info("Собрали и обновили метрики.", slog.Any("До обновления ", metricCollection.Metrics))
}
