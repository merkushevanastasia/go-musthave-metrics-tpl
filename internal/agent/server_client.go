package agent

import (
	"errors"

	config "github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/config/agent"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/constants"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/dto"
	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog/log"
)

var metricServerClient *resty.Client

func SendAll(values *dto.MetricCollection) {
	log.Info().Msg("Отправляем метрики на сервер...")
	for metricName, metricValue := range values.Metrics {
		var request dto.MetricsRequest
		if metricValue.MetricType == constants.CounterMetricType {
			request = dto.MetricsRequest{ID: metricName,
				MType: metricValue.MetricType,
				Delta: &metricValue.Counter}
		} else {
			request = dto.MetricsRequest{ID: metricName,
				MType: metricValue.MetricType,
				Value: &metricValue.Gauge}
		}
		err := send(request)
		if err != nil {
			log.Error().Any("err", err).Msg("Произошла ошибка")
		}
	}
	log.Info().Msg("Отправка метрик на сервер завершена...")

}

func send(request dto.MetricsRequest) error {

	log.Debug().Any("Метрика", request.ID).Msg("Отправляется ")
	resp, err := metricServerClient.R().
		SetHeader("Content-Type", "application/json").
		SetBody(request).
		Post("/update")

	if err != nil {
		return err
	}

	if resp.IsError() {
		log.Error().Any("status", resp.Status()).Msg("Сервер вернул ошибку")
		return errors.New("Вернулсся ошибочный код" + resp.Status())
	}

	log.Info().Msg("Отправлено..")
	return nil
}

func InitClient(config config.Config) {
	metricServerClient = resty.New().SetBaseURL(config.ServerURL)
}
