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
				Alloc:         constants.GaugeMetricType,
				BuckHashSys:   constants.GaugeMetricType,
				Frees:         constants.GaugeMetricType,
				GCCPUFraction: constants.GaugeMetricType,
				GCSys:         constants.GaugeMetricType,
				HeapAlloc:     constants.GaugeMetricType,
				HeapIdle:      constants.GaugeMetricType,
				HeapInuse:     constants.GaugeMetricType,
				HeapObjects:   constants.GaugeMetricType,
				HeapReleased:  constants.GaugeMetricType,
				HeapSys:       constants.GaugeMetricType,
				LastGC:        constants.GaugeMetricType,
				Lookups:       constants.GaugeMetricType,
				MCacheInuse:   constants.GaugeMetricType,
				MCacheSys:     constants.GaugeMetricType,
				MSpanInuse:    constants.GaugeMetricType,
				MSpanSys:      constants.GaugeMetricType,
				Mallocs:       constants.GaugeMetricType,
				NextGC:        constants.GaugeMetricType,
				NumForcedGC:   constants.GaugeMetricType,
				NumGC:         constants.GaugeMetricType,
				OtherSys:      constants.GaugeMetricType,
				PauseTotalNs:  constants.GaugeMetricType,
				StackInuse:    constants.GaugeMetricType,
				StackSys:      constants.GaugeMetricType,
				Sys:           constants.GaugeMetricType,
				TotalAlloc:    constants.GaugeMetricType,
				RandomValue:   constants.CounterMetricType,
				PollCount:     constants.CounterMetricType,
			}

			for metricName, metricType := range metricsToTest {
				val, ok := currentValues.Metrics[metricName]
				if !ok {
					t.Errorf("Метрика %s не была собрана (отсутствует в мапе GaugeMetrics)", metricName)
					continue
				}

				// Некоторые метрики рантайма гарантированно равны 0 на старте программы (так как сборщик мусора еще не успел поработать), делаем для них исключение
				// (это я узнала у ИИ, но спорить не стала)
				canBeZero := metricName == NumForcedGC ||
					metricName == PauseTotalNs ||
					metricName == Lookups ||
					metricName == HeapReleased ||
					metricName == LastGC ||
					metricName == GCCPUFraction ||
					metricName == NumGC ||
					metricName == RandomValue
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

			// Вызываем Update второй раз, чтобы проверить логику инкремента
			Update(currentValues)

			if currentValues.Metrics[PollCount].Counter != 2 {
				t.Errorf("После второго Update ожидался PollCount = 2, получили %d", currentValues.Metrics[PollCount].Counter)
			}
		})
	}
}
