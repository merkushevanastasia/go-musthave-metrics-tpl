package agent

import (
	"net/http"
	"net/http/httptest"
	"testing"

	config "github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/config/agent"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/constants"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/dto"
	"github.com/stretchr/testify/assert"
)

func TestSendAllSuccess(t *testing.T) {

	actualUrls := make([]string, 0)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		actualUrls = append(actualUrls, r.URL.Path)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()
	InitClient(config.Config{ServerURL: ts.URL})

	type args struct {
		values        *dto.MetricCollection
		baseURL       string
		expectedPaths []string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Успешно отправлены все метрики",
			args: args{
				values: &dto.MetricCollection{
					Metrics: map[string]dto.MetricDto{
						Alloc: {
							MetricType: constants.GaugeMetricType,
							Gauge:      12345.67,
						},
						HeapAlloc: {
							MetricType: constants.GaugeMetricType,
							Gauge:      500000,
						},
						PollCount: {
							MetricType: constants.CounterMetricType,
							Counter:    5,
						},
					},
				},
				expectedPaths: []string{
					"/update/gauge/Alloc/12345.67",
					"/update/gauge/HeapAlloc/500000",
					"/update/counter/PollCount/5",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SendAll(tt.args.values)
			assert.ElementsMatch(t, tt.args.expectedPaths, actualUrls)
		})
	}
}
