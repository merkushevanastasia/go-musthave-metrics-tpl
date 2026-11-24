package agent

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"

	config "github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/config/agent"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/constants"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/model/dto"
	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog/log"
)

var metricServerClient *resty.Client

// SendBatch отправляет накопленные метрики одним батчем на POST /updates/
func SendBatch(values *dto.MetricCollection) {
	batch := collectSnapshot(values)
	if len(batch) == 0 {
		return
	}

	log.Info().Int("count", len(batch)).Msg("Отправляем батч метрик на сервер...")
	if err := sendBatch(batch); err != nil {
		log.Error().Err(err).Msg("Произошла ошибка при отправке батча метрик")
		return
	}
	log.Info().Msg("Отправка батча метрик завершена")
}

// SendAll отправляет метрики поштучно на POST /update (сохранено для обратной совместимости)
func SendAll(values *dto.MetricCollection) {
	metrics := collectSnapshot(values)
	if len(metrics) == 0 {
		return
	}

	log.Info().Msg("Отправляем метрики на сервер поштучно...")
	for _, metric := range metrics {
		if err := send(metric); err != nil {
			log.Error().Err(err).Str("metric", metric.ID).Msg("Произошла ошибка при отправке метрики")
		}
	}
	log.Info().Msg("Отправка метрик на сервер завершена")
}

func collectSnapshot(values *dto.MetricCollection) []dto.MetricDto {
	values.Mu.Lock()
	defer values.Mu.Unlock()

	if len(values.Metrics) == 0 {
		return nil
	}

	snapshot := make([]dto.MetricDto, 0, len(values.Metrics))
	for _, metric := range values.Metrics {
		item := metric
		if metric.Delta != nil {
			d := *metric.Delta
			item.Delta = &d
		}
		if metric.Value != nil {
			v := *metric.Value
			item.Value = &v
		}
		snapshot = append(snapshot, item)
	}

	// Сбрасываем счетчик обновлений под тем же мьютексом, чтобы исключить race condition
	values.Add(toEntityCounter(constants.PollCount, 0))

	return snapshot
}

func sendBatch(batch []dto.MetricDto) error {
	resp, err := metricServerClient.R().
		SetHeader("Content-Type", "application/json").
		SetBody(batch).
		Post("/updates/")

	if err != nil {
		return err
	}

	if resp.IsError() {
		log.Error().Str("status", resp.Status()).Msg("Сервер вернул ошибку при приеме батча")
		return errors.New("вернулся ошибочный код: " + resp.Status())
	}

	return nil
}

func send(request dto.MetricDto) error {
	log.Debug().Str("metric", request.ID).Msg("Отправляется метрика")
	resp, err := metricServerClient.R().
		SetHeader("Content-Type", "application/json").
		SetBody(request).
		Post("/update")

	if err != nil {
		return err
	}

	if resp.IsError() {
		log.Error().Str("status", resp.Status()).Msg("Сервер вернул ошибку")
		return errors.New("вернулся ошибочный код: " + resp.Status())
	}

	return nil
}

func InitClient(config config.Config) {
	metricServerClient = resty.New().
		SetBaseURL(config.ServerURL).
		SetHeader("Accept-Encoding", "gzip")

	// Сжимаем тело исходящего JSON-запроса алгоритмом gzip
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
