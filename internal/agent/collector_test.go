package agent

import "testing"

func TestUpdate(t *testing.T) {
	type args struct {
		currentValues *CurrentMetricValues
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "TestUpdateOK",
			args: args{
				currentValues: &CurrentMetricValues{
					CounterMetrics: make(map[string]int64),
					GaugeMetrics:   make(map[string]float64),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			currentValues := tt.args.currentValues

			// Вызываем тестируемую функцию вперые
			Update(currentValues)

			metricsToTest := []string{
				Alloc,
				BuckHashSys,
				Frees,
				GCCPUFraction,
				GCSys,
				HeapAlloc,
				HeapIdle,
				HeapInuse,
				HeapObjects,
				HeapReleased,
				HeapSys,
				LastGC,
				Lookups,
				MCacheInuse,
				MCacheSys,
				MSpanInuse,
				MSpanSys,
				Mallocs,
				NextGC,
				NumForcedGC,
				NumGC,
				OtherSys,
				PauseTotalNs,
				StackInuse,
				StackSys,
				Sys,
				TotalAlloc,
			}

			for _, metricName := range metricsToTest {
				val, ok := currentValues.GaugeMetrics[metricName]
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
					metricName == NumGC

				if !canBeZero && val <= 0 {
					t.Errorf("Метрика %s должна быть больше 0, но получили %f", metricName, val)
				}
			}

			// 3. Вызываем Update второй раз, чтобы проверить логику инкремента
			Update(currentValues)

			if currentValues.CounterMetrics[PollCount] != 2 {
				t.Errorf("После второго Update ожидался PollCount = 2, получили %d", currentValues.CounterMetrics[PollCount])
			}
		})
	}
}
