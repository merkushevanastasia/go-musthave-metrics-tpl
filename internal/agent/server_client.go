package agent

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"net/http"
	"time"

	config "github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/config/agent"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/constants"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/model/dto"
	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog/log"
)

var metricServerClient *resty.Client

type httpStatusError struct {
	code   int
	status string
}

func (e *httpStatusError) Error() string {
	return fmt.Sprintf("вернулся ошибочный код: %s", e.status)
}

// SendBatch отправляет накопленные метрики батчем на POST /updates/.
// Если сервер не поддерживает /updates/ (404/405), автоматически откатывается на поштучную отправку POST /update.
func SendBatch(values *dto.MetricCollection) {
	batch := collectSnapshot(values)
	if len(batch) == 0 {
		return
	}

	log.Info().Int("count", len(batch)).Msg("Отправляем батч метрик на сервер...")
	err := sendBatch(batch)
	if err == nil {
		log.Info().Msg("Отправка батча метрик завершена")
		return
	}

	var statusErr *httpStatusError
	if errors.As(err, &statusErr) && (statusErr.code == http.StatusNotFound || statusErr.code == http.StatusMethodNotAllowed) {
		log.Warn().Msg("Эндпоинт /updates/ не поддерживается сервером, переключаемся на поштучную отправку /update")
		sendSlice(batch)
		return
	}

	log.Error().Err(err).Msg("Произошла ошибка при отправке батча метрик")
}

// SendAll отправляет метрики поштучно на POST /update (сохранено для обратной совместимости)
func SendAll(values *dto.MetricCollection) {
	metrics := collectSnapshot(values)
	if len(metrics) == 0 {
		return
	}

	sendSlice(metrics)
}

func sendSlice(metrics []dto.MetricDto) {
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
		return &httpStatusError{code: resp.StatusCode(), status: resp.Status()}
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
		return &httpStatusError{code: resp.StatusCode(), status: resp.Status()}
	}

	return nil
}

func InitClient(config config.Config) {
	delays := []time.Duration{
		1 * time.Second,
		3 * time.Second,
		5 * time.Second,
	}

	metricServerClient = resty.New().
		SetBaseURL(config.ServerURL).
		SetHeader("Accept-Encoding", "gzip").
		SetRetryCount(3).
		SetRetryAfter(func(client *resty.Client, resp *resty.Response) (time.Duration, error) {
			idx := resp.Request.Attempt - 1
			if idx >= 0 && idx < len(delays) {
				return delays[idx], nil
			}
			return delays[len(delays)-1], nil
		}).
		AddRetryCondition(func(r *resty.Response, err error) bool {
			return isRetriableNetworkError(r, err)
		})

	metricServerClient.OnBeforeRequest(func(c *resty.Client, r *resty.Request) error {
		if r.Body == nil {
			return nil
		}
		if _, alreadyCompressed := r.Body.([]byte); alreadyCompressed && r.Header.Get("Content-Encoding") == "gzip" {
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
func isRetriableNetworkError(r *resty.Response, err error) bool {

	if err != nil {
		return true
	}

	if r != nil {
		switch r.StatusCode() {
		case http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
			return true
		}
	}

	return false
}
