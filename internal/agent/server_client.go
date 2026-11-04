package agent

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"

	config "github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/config/agent"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/model/dto"
	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog/log"
)

var metricServerClient *resty.Client

func SendAll(values *dto.MetricCollection) {
	log.Info().Msg("Отправляем метрики на сервер...")
	for _, metric := range values.Metrics {
		err := send(metric)
		if err != nil {
			log.Error().Any("err", err).Msg("Произошла ошибка")
		}
	}
	log.Info().Msg("Отправка метрик на сервер завершена...")

}

func send(request dto.MetricDto) error {

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
	metricServerClient = resty.New().
		SetBaseURL(config.ServerURL).
		SetHeader("Accept-Encoding", "gzip")

	// Добавляем middleware для автоматического сжатия исходящих POST/PUT запросов с JSON
	metricServerClient.OnBeforeRequest(func(c *resty.Client, r *resty.Request) error {
		if r.Body == nil {
			return nil
		}
		bodyBytes, err := c.JSONMarshal(r.Body)
		if err != nil {
			return fmt.Errorf("failed to marshal JSON before gzip: %w", err)
		}

		var buf bytes.Buffer
		gz := gzip.NewWriter(&buf)
		if _, err := gz.Write(bodyBytes); err != nil {
			return fmt.Errorf("failed to write gzip: %w", err)
		}
		if err := gz.Close(); err != nil {
			return fmt.Errorf("failed to close gzip writer: %w", err)
		}

		r.SetBody(buf.Bytes())
		r.SetHeader("Content-Encoding", "gzip")

		return nil
	})
}
