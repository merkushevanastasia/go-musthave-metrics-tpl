package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/dto"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

func TestHandleMetricUpdate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	type ExpectedData struct {
		Code        int
		Response    string
		ContentType string
	}
	type InputData struct {
		URL    string
		Method string
	}

	tests := []struct {
		name  string
		input InputData
		want  ExpectedData
	}{
		{
			name: "TestHandleMetricUpdateGaugeOk",
			input: InputData{
				URL:    "/update/gauge/NameMetric/10",
				Method: http.MethodPost,
			},
			want: ExpectedData{
				Code:        200,
				ContentType: "text/plain; charset=utf-8",
				Response:    "",
			},
		},
		{
			name: "TestHandleMetricUpdateCounterOk",
			input: InputData{
				URL:    "/update/counter/NameMetric/10",
				Method: http.MethodPost,
			},
			want: ExpectedData{
				Code:        200,
				ContentType: "text/plain; charset=utf-8",
				Response:    "",
			},
		},
		{
			name: "TestHandleMetricUpdateMetricNameError",
			input: InputData{
				URL:    "/update/counter//10",
				Method: http.MethodPost,
			},
			want: ExpectedData{
				Code:        404,
				ContentType: "text/plain; charset=utf-8",
				Response:    "в запросе отсутствует имя метрики",
			},
		},
	}

	mockServiceImpl := &service.MockMetricService{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()

			router.POST("/update/:metric_type/:metric_name/:metric_value", HandleMetricUpdate(mockServiceImpl))

			request := httptest.NewRequest(tt.input.Method, tt.input.URL, nil)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, request)

			response := w.Result()

			defer func(Body io.ReadCloser) {
				err := Body.Close()
				if err != nil {
					log.Logger.Error().Any("err", err).Msg("Body close error")
				}
			}(response.Body)

			assert.Equal(t, tt.want.Code, response.StatusCode)

			resBody, err := io.ReadAll(response.Body)
			assert.NoError(t, err)
			assert.Equal(t, tt.want.Response, string(resBody))
		})
	}
}
func floatPtr(v float64) *float64 {
	return &v
}
func TestHandleMetricUpdateV2(t *testing.T) {
	gin.SetMode(gin.TestMode)

	type ExpectedData struct {
		Code        int
		Response    string
		ContentType string
	}
	type InputData struct {
		Request dto.MetricsRequest
		Method  string
	}

	tests := []struct {
		name  string
		input InputData
		want  ExpectedData
	}{
		{
			name: "TestHandleMetricUpdateGaugeOk",
			input: InputData{
				Request: dto.MetricsRequest{
					ID:    "NameMetric",
					MType: "gauge",
					Value: floatPtr(10.0),
				},
				Method: http.MethodPost,
			},
			want: ExpectedData{
				Code:        200,
				ContentType: "text/plain; charset=utf-8",
				Response:    "",
			},
		},
		{
			name: "TestHandleMetricUpdateCounterOk",
			input: InputData{
				Request: dto.MetricsRequest{
					ID:    "NameMetric",
					MType: "counter",
					Value: floatPtr(10.0),
				},
				Method: http.MethodPost,
			},
			want: ExpectedData{
				Code:        200,
				ContentType: "text/plain; charset=utf-8",
				Response:    "",
			},
		},
		{
			name: "TestHandleMetricUpdateMetricNameError",
			input: InputData{
				Request: dto.MetricsRequest{
					ID:    "",
					MType: "counter",
					Value: floatPtr(10.0),
				},
				Method: http.MethodPost,
			},
			want: ExpectedData{
				Code:        404,
				ContentType: "text/plain; charset=utf-8",
				Response:    "в запросе отсутствует имя метрики",
			},
		},
	}

	mockServiceImpl := &service.MockMetricService{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			bodyBytes, err := json.Marshal(tt.input.Request)
			if err != nil {
				t.Fatalf("Не удалось замаршалить структуру в JSON: %v", err)
			}

			// 2. Передаем bytes.NewReader(bodyBytes) третьим параметром
			request := httptest.NewRequest(tt.input.Method, "/update", bytes.NewReader(bodyBytes))

			// 3. ОБЯЗАТЕЛЬНО задаем заголовок Content-Type, иначе Gin не поймет, что это JSON
			request.Header.Set("Content-Type", "application/json")
			router.POST("/update", HandleMetricUpdateV2(mockServiceImpl))

			w := httptest.NewRecorder()
			router.ServeHTTP(w, request)

			response := w.Result()

			defer func(Body io.ReadCloser) {
				err := Body.Close()
				if err != nil {
					log.Logger.Error().Any("err", err).Msg("Body close error")
				}
			}(response.Body)

			assert.Equal(t, tt.want.Code, response.StatusCode)

			resBody, err := io.ReadAll(response.Body)
			assert.NoError(t, err)
			assert.Equal(t, tt.want.Response, string(resBody))
		})
	}
}
