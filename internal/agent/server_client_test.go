package agent

import (
	"encoding/json"
	"io"
	"log"
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
	actualBodys := make([]dto.MetricsRequest, 0)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actualUrls = append(actualUrls, r.URL.Path)

		// 2. Считываем тело запроса в []byte
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()
		var req dto.MetricsRequest

		err = json.Unmarshal(bodyBytes, &req)
		if err != nil {
			log.Fatalf("Ошибка десериализации JSON: %v", err)
		}

		actualBodys = append(actualBodys, req)

		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	InitClient(config.Config{ServerURL: ts.URL})

	type args struct {
		values        *dto.MetricCollection
		baseURL       string
		expectedPaths []string
		expectedBodys []dto.MetricsRequest
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
						constants.Alloc: {
							MetricType: constants.GaugeMetricType,
							Gauge:      12345.67,
						},
						constants.HeapAlloc: {
							MetricType: constants.GaugeMetricType,
							Gauge:      500000,
						},
						constants.PollCount: {
							MetricType: constants.CounterMetricType,
							Counter:    5,
						},
					},
				},
				expectedPaths: []string{
					"/update",
					"/update",
					"/update",
				},
				expectedBodys: []dto.MetricsRequest{
					dto.MetricsRequest{
						ID:    constants.Alloc,
						MType: constants.GaugeMetricType,
						Value: floatPtr(12345.67),
					},
					dto.MetricsRequest{
						ID:    constants.HeapAlloc,
						MType: constants.GaugeMetricType,
						Value: floatPtr(500000),
					},
					dto.MetricsRequest{
						ID:    constants.PollCount,
						MType: constants.CounterMetricType,
						Delta: intPtr(5),
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SendAll(tt.args.values)
			assert.ElementsMatch(t, tt.args.expectedPaths, actualUrls)
			assert.ElementsMatch(t, tt.args.expectedBodys, actualBodys)
		})
	}
}
func floatPtr(v float64) *float64 {
	return new(v)
}
func intPtr(v int64) *int64 {
	return new(v)
}
