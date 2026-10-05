package handler

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandleMetricUpdate(t *testing.T) {
	type ExpectedData struct {
		Code        int
		Response    string
		ContentType string
	}
	type InputData struct {
		PathVars map[string]string
		URL      string
		Method   string
	}

	tests := []struct {
		name  string
		input InputData
		want  ExpectedData
	}{
		{
			name: "TestHandleMetricUpdateGaugeOk",
			input: InputData{
				PathVars: map[string]string{
					MetricTypePathKey:  "gauge",
					MetricValuePathKey: "10",
					MetricNamePathKey:  "NameMetric",
				},
				URL:    "/update/",
				Method: http.MethodPost,
			},
			want: ExpectedData{
				Code:        200,
				ContentType: "text/plain; charset=utf-8",
				Response:    "",
			},
		}, {
			name: "TestHandleMetricUpdateCounterOk",
			input: InputData{
				PathVars: map[string]string{
					MetricTypePathKey:  "counter",
					MetricValuePathKey: "10",
					MetricNamePathKey:  "NameMetric",
				},
				URL:    "/update/",
				Method: http.MethodPost,
			},
			want: ExpectedData{
				Code:        200,
				ContentType: "text/plain; charset=utf-8",
				Response:    "",
			},
		}, {
			name: "TestHandleMetricUpdateMetricNameError",
			input: InputData{
				PathVars: map[string]string{
					MetricTypePathKey:  "counter",
					MetricValuePathKey: "10",
					MetricNamePathKey:  "",
				},
				URL:    "/update",
				Method: http.MethodPost,
			},
			want: ExpectedData{
				Code:        404,
				ContentType: "text/plain; charset=utf-8",
				Response:    "в запросе отсутствует имя метрики в запросе отсутствует имя метрики\n",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {})
		{

			request := httptest.NewRequest(tt.input.Method, tt.input.URL, nil)
			for k, v := range tt.input.PathVars {
				request.SetPathValue(k, v)
			}
			w := httptest.NewRecorder()
			HandleMetricUpdate(w, request)

			response := w.Result()
			assert.Equal(t, tt.want.Code, response.StatusCode)
			defer func(Body io.ReadCloser) {
				err := Body.Close()
				if err != nil {
					slog.Error("Body close: %v", slog.Any("err", err))
				}
			}(response.Body)
			resBody, err := io.ReadAll(response.Body)
			assert.Equal(t, tt.want.Response, string(resBody))
			assert.NoError(t, err)

		}

	}
}
