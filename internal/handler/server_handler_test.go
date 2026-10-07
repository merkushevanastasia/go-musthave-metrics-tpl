package handler

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
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
		}, {
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
		}, {
			name: "TestHandleMetricUpdateMetricNameError",
			input: InputData{
				URL:    "/update/counter//10",
				Method: http.MethodPost,
			},
			want: ExpectedData{
				Code:        404,
				ContentType: "text/plain; charset=utf-8",
				Response:    "в запросе отсутствует имя метрики в запросе отсутствует имя метрики",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {})
		{
			// 1. Создаем роутер Gin и регистрируем хэндлер с параметрами пути
			router := gin.New()

			// Настройте параметры пути точно так, как они объявлены в вашем main.go
			// Например: /update/:metricType/:metricName/:metricValue
			router.POST("/update/:metric_type/:metric_name/:metric_value", HandleMetricUpdate)

			// 2. Создаем запрос и рекордер
			request := httptest.NewRequest(tt.input.Method, tt.input.URL, nil)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, request)

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
