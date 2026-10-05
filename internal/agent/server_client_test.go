package agent

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSendAllSuccess(t *testing.T) {

	actualUrls := make([]string, 0)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actualUrls = append(actualUrls, r.URL.Path)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	type args struct {
		values        *CurrentMetricValues
		baseUrl       string
		expectedPaths []string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Успешно отправлены все метрики",
			args: args{
				values: &CurrentMetricValues{
					GaugeMetrics: map[string]float64{
						"Alloc":     12345.67,
						"HeapAlloc": 500000,
					},
					CounterMetrics: map[string]int64{
						"PollCount": 5,
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
			SendAll(tt.args.values, ts.URL)
			assert.ElementsMatch(t, tt.args.expectedPaths, actualUrls)
		})
	}
}
