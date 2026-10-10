package agent

import (
	"errors"
	"log/slog"
	"strconv"

	config "github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/config/agent"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/constants"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/dto"
	"github.com/go-resty/resty/v2"
)

var metricServerClient *resty.Client

func SendAll(values *dto.MetricCollection) {
	slog.Info("Отправляем метрики на сервер...")
	for metricName, metricValue := range values.Metrics {
		var valueStr string
		if metricValue.MetricType == constants.CounterMetricType {
			valueStr = strconv.FormatInt(metricValue.Counter, 10)
		} else {
			valueStr = strconv.FormatFloat(metricValue.Gauge, 'f', -1, 64)
		}
		err := send(metricValue.MetricType, metricName, valueStr)
		if err != nil {
			slog.Error("Произошла ошибка", slog.Any("err", err))
		}
	}
	slog.Info("Отправка метрик на сервер завершена...")

}

func send(metricType string, metricName string, metricValue string) error {

	slog.Debug("Отправляется ", slog.Any("Метрика", metricName))
	resp, err := metricServerClient.R().
		SetHeader("Content-Type", "text/plain").
		SetPathParams(map[string]string{
			"type":  metricType,
			"name":  metricName,
			"value": metricValue,
		}).
		Post("/update/{type}/{name}/{value}")

	if err != nil {
		return err
	}

	if resp.IsError() {
		slog.Error("Сервер вернул ошибку", "status", resp.Status())
		return errors.New("Вернулсся ошибочный код" + resp.Status())
	}

	slog.Info("Отправлено..")
	return nil
}

func InitClient(config config.Config) {
	metricServerClient = resty.New().SetBaseURL(config.ServerURL)
}
