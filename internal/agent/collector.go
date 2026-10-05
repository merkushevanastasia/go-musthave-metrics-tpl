package agent

import (
	"log/slog"
	"math/rand/v2"
	"runtime"
)

func Update(currentValues *CurrentMetricValues) {

	currentValues.mu.Lock()
	defer currentValues.mu.Unlock()

	slog.Info("Запускаем процесс сбора метрик.", slog.Any("До обновления gauges", currentValues.GaugeMetrics), slog.Any("До обновления counters", currentValues.CounterMetrics))
	slog.Info("Собираем метрики...")

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	currentValues.GaugeMetrics[Alloc] = float64(m.Alloc)
	currentValues.GaugeMetrics[BuckHashSys] = float64(m.BuckHashSys)
	currentValues.GaugeMetrics[Frees] = float64(m.Frees)
	currentValues.GaugeMetrics[GCCPUFraction] = m.GCCPUFraction
	currentValues.GaugeMetrics[GCSys] = float64(m.GCSys)
	currentValues.GaugeMetrics[HeapAlloc] = float64(m.HeapAlloc)
	currentValues.GaugeMetrics[HeapIdle] = float64(m.HeapIdle)
	currentValues.GaugeMetrics[HeapInuse] = float64(m.HeapInuse)
	currentValues.GaugeMetrics[HeapObjects] = float64(m.HeapObjects)
	currentValues.GaugeMetrics[HeapReleased] = float64(m.HeapReleased)
	currentValues.GaugeMetrics[HeapSys] = float64(m.HeapSys)
	currentValues.GaugeMetrics[LastGC] = float64(m.LastGC)
	currentValues.GaugeMetrics[Lookups] = float64(m.Lookups)
	currentValues.GaugeMetrics[MCacheInuse] = float64(m.MCacheInuse)
	currentValues.GaugeMetrics[MCacheSys] = float64(m.MCacheSys)
	currentValues.GaugeMetrics[MSpanInuse] = float64(m.MSpanInuse)
	currentValues.GaugeMetrics[MSpanSys] = float64(m.MSpanSys)
	currentValues.GaugeMetrics[Mallocs] = float64(m.Mallocs)
	currentValues.GaugeMetrics[NextGC] = float64(m.NextGC)
	currentValues.GaugeMetrics[NumForcedGC] = float64(m.NumForcedGC)
	currentValues.GaugeMetrics[NumGC] = float64(m.NumGC)
	currentValues.GaugeMetrics[OtherSys] = float64(m.OtherSys)
	currentValues.GaugeMetrics[PauseTotalNs] = float64(m.PauseTotalNs)
	currentValues.GaugeMetrics[StackInuse] = float64(m.StackInuse)
	currentValues.GaugeMetrics[StackInuse] = float64(m.StackInuse)
	currentValues.GaugeMetrics[StackSys] = float64(m.StackSys)
	currentValues.GaugeMetrics[Sys] = float64(m.Sys)
	currentValues.GaugeMetrics[TotalAlloc] = float64(m.TotalAlloc)

	currentValues.CounterMetrics[PollCount] += 1
	currentValues.CounterMetrics[RandomValue] = rand.Int64N(10000)

	slog.Info("Собрали и обновили метрики.", slog.Any("До обновления gauges", currentValues.GaugeMetrics), slog.Any("До обновления counters", currentValues.CounterMetrics))

}
