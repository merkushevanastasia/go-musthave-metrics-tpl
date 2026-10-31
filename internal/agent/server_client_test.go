package agent

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	config "github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/config/agent"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/constants"
	dto2 "github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/model/dto"
	"github.com/stretchr/testify/assert"
)

func TestSendAllSuccess(t *testing.T) {

	actualUrls := make([]string, 0)
	actualBodys := make([]dto2.MetricDto, 0)
	actualEncodings := make([]string, 0)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actualUrls = append(actualUrls, r.URL.Path)
		actualEncodings = append(actualEncodings, r.Header.Get("Content-Encoding"))

		var bodyReader = r.Body
		defer bodyReader.Close()

		if r.Header.Get("Content-Encoding") == "gzip" {
			gzipReader, err := gzip.NewReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			defer gzipReader.Close()
			bodyReader = gzipReader
		}

		// Считываем тело запроса (уже распакованное, если был gzip)
		bodyBytes, err := io.ReadAll(bodyReader)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		var req dto2.MetricDto
		err = json.Unmarshal(bodyBytes, &req)
		if err != nil {
			log.Printf("Ошибка десериализации JSON: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		actualBodys = append(actualBodys, req)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	InitClient(config.Config{ServerURL: ts.URL})

	type args struct {
		values        *dto2.MetricCollection
		baseURL       string
		expectedPaths []string
		expectedBodys []dto2.MetricDto
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Успешно отправлены все метрики",
			args: args{
				values: &dto2.MetricCollection{
					Metrics: map[string]dto2.MetricDto{
						constants.Alloc: {
							MType: constants.GaugeMetricType,
							Value: floatPtr(12345.67),
							ID:    constants.Alloc,
						},
						constants.HeapAlloc: {
							MType: constants.GaugeMetricType,
							Value: floatPtr(500000),
							ID:    constants.HeapAlloc,
						},
						constants.PollCount: {
							MType: constants.CounterMetricType,
							Delta: intPtr(5),
							ID:    constants.PollCount,
						},
					},
				},
				expectedPaths: []string{
					"/update",
					"/update",
					"/update",
				},
				expectedBodys: []dto2.MetricDto{
					{
						ID:    constants.Alloc,
						MType: constants.GaugeMetricType,
						Value: floatPtr(12345.67),
					},
					{
						ID:    constants.HeapAlloc,
						MType: constants.GaugeMetricType,
						Value: floatPtr(500000),
					},
					{
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
