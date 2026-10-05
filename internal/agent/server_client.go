package agent

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
)

func SendAll(values *CurrentMetricValues, baseUrl string) {
	slog.Info("Отправляем метрики на сервер...")
	for metricName, metricValue := range values.GaugeMetrics {
		err := send(baseUrl, "gauge", metricName, strconv.FormatFloat(metricValue, 'f', -1, 64))
		if err != nil {
			slog.Error("Произошла ошибка", slog.Any("err", err))
		}
	}
	for metricName, metricValue := range values.CounterMetrics {
		err := send(baseUrl, "counter", metricName, strconv.FormatInt(metricValue, 10))
		if err != nil {
			slog.Error("Произошла ошибка", slog.Any("err", err))
		}
	}
}
func send(baseUrl string, metricType string, metricName string, metricValue string) error {
	slog.Info("Отправляется ", slog.Any("Метрика", metricName))

	fullPath := fmt.Sprintf("%s/update/%s/%s/%s", baseUrl, metricType, metricName, metricValue)
	response, err := http.Post(fullPath, "text/plain", http.NoBody)
	if err != nil {
		return err
	}

	_, err = io.ReadAll(response.Body)
	err = response.Body.Close()
	if err != nil {
		slog.Error("Произошла ошибка при чтении ответа, но работа продолжается..")
	}

	slog.Error("Отправлено..")
	return nil
}
