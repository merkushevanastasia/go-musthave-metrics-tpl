package agent

import (
	"testing"

	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/constants"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/dto"
)

func TestUpdate(t *testing.T) {
	type args struct {
		currentValues *dto.MetricCollection
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "TestUpdateOK",
			args: args{
				currentValues: &dto.MetricCollection{Metrics: make(map[string]dto.MetricDto)},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			currentValues := tt.args.currentValues

			// Вызываем тестируемую функцию вперые
			Update(currentValues)

			metricsToTest := map[string]string{
				constants.Alloc:         constants.GaugeMetricType,
				constants.BuckHashSys:   constants.GaugeMetricType,
				constants.Frees:         constants.GaugeMetricType,
				constants.GCCPUFraction: constants.GaugeMetricType,
				constants.GCSys:         constants.GaugeMetricType,
				constants.HeapAlloc:     constants.GaugeMetricType,
				constants.HeapIdle:      constants.GaugeMetricType,
				constants.HeapInuse:     constants.GaugeMetricType,
				constants.HeapObjects:   constants.GaugeMetricType,
				constants.HeapReleased:  constants.GaugeMetricType,
				constants.HeapSys:       constants.GaugeMetricType,
				constants.LastGC:        constants.GaugeMetricType,
				constants.Lookups:       constants.GaugeMetricType,
				constants.MCacheInuse:   constants.GaugeMetricType,
				constants.MCacheSys:     constants.GaugeMetricType,
				constants.MSpanInuse:    constants.GaugeMetricType,
				constants.MSpanSys:      constants.GaugeMetricType,
				constants.Mallocs:       constants.GaugeMetricType,
				constants.NextGC:        constants.GaugeMetricType,
				constants.NumForcedGC:   constants.GaugeMetricType,
				constants.NumGC:         constants.GaugeMetricType,
				constants.OtherSys:      constants.GaugeMetricType,
				constants.PauseTotalNs:  constants.GaugeMetricType,
				constants.StackInuse:    constants.GaugeMetricType,
				constants.StackSys:      constants.GaugeMetricType,
				constants.Sys:           constants.GaugeMetricType,
				constants.TotalAlloc:    constants.GaugeMetricType,
				constants.RandomValue:   constants.CounterMetricType,
				constants.PollCount:     constants.CounterMetricType,
			}

			for metricName, metricType := range metricsToTest {
				val, isPresent := currentValues.Metrics[metricName]
				if !isPresent {
					t.Errorf("Метрика %s не была собрана (отсутствует в Metrics)", metricName)
					continue
				}

				// Некоторые метрики рантайма гарантированно равны 0 на старте программы (так как сборщик мусора еще не успел поработать), делаем для них исключение
				// (это я узнала у ИИ, но спорить не стала)
				canBeZero := metricName == constants.NumForcedGC ||
					metricName == constants.PauseTotalNs ||
					metricName == constants.Lookups ||
					metricName == constants.HeapReleased ||
					metricName == constants.LastGC ||
					metricName == constants.GCCPUFraction ||
					metricName == constants.NumGC ||
					metricName == constants.RandomValue
				if metricType == constants.GaugeMetricType {
					if !canBeZero && val.Gauge <= 0 {
						t.Errorf("Метрика %s должна быть больше 0, но получили %f", metricName, val.Gauge)
					}
				}

				if metricType == constants.CounterMetricType {
					if !canBeZero && val.Counter <= 0 {
						t.Errorf("Метрика %s должна быть больше 0, но получили %d", metricName, val.Counter)
					}
				}
			}

			if currentValues.Metrics[constants.PollCount].Counter != 1 {
				t.Errorf("После первого Update ожидался PollCount = 1, получили %d", currentValues.Metrics[constants.PollCount].Counter)
			}
			// Вызываем Update второй раз, чтобы проверить логику инкремента
			Update(currentValues)

			if currentValues.Metrics[constants.PollCount].Counter != 2 {
				t.Errorf("После второго Update ожидался PollCount = 2, получили %d", currentValues.Metrics[constants.PollCount].Counter)
			}
		})
	}
}
