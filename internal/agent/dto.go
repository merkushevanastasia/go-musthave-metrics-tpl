package agent

import "sync"

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

type CurrentMetricValues struct {
	mu             sync.Mutex
	GaugeMetrics   map[string]float64
	CounterMetrics map[string]int64
}

func NewCurrentValues() *CurrentMetricValues {
	return &CurrentMetricValues{
		GaugeMetrics:   make(map[string]float64),
		CounterMetrics: make(map[string]int64),
	}
}
