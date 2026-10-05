package agent

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
)

func SendAll(values *CurrentMetricValues, baseURL string) {
	slog.Info("Отправляем метрики на сервер...")
	for metricName, metricValue := range values.GaugeMetrics {
		err := send(baseURL, "gauge", metricName, strconv.FormatFloat(metricValue, 'f', -1, 64))
		if err != nil {
			slog.Error("Произошла ошибка", slog.Any("err", err))
		}
	}
	for metricName, metricValue := range values.CounterMetrics {
		err := send(baseURL, "counter", metricName, strconv.FormatInt(metricValue, 10))
		if err != nil {
			slog.Error("Произошла ошибка", slog.Any("err", err))
		}
	}
}
func send(baseURL string, metricType string, metricName string, metricValue string) error {
	slog.Info("Отправляется ", slog.Any("Метрика", metricName))

	fullPath := fmt.Sprintf("%s/update/%s/%s/%s", baseURL, metricType, metricName, metricValue)
	response, err := http.Post(fullPath, "text/plain", http.NoBody)
	if err != nil {
		return err
	}

	_, err = io.ReadAll(response.Body)
	if err != nil {
		slog.Error("Произошла ошибка при чтении ответа, но работа продолжается..")
	}
	err = response.Body.Close()
	if err != nil {
		slog.Error("Произошла ошибка при чтении ответа, но работа продолжается..")
	}

	slog.Error("Отправлено..")
	return nil
}
