package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/constants"
	servererror "github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/error"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/model/dto"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/service"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/utils/logging"
	"github.com/gin-gonic/gin"
)

// HandleMetricGetV1 хэндлер для обработки запроса на получение значения метрики
func HandleMetricGetV1(metricService service.MetricService) gin.HandlerFunc {
	return func(c *gin.Context) {

		log := logging.FromContext(c.Request.Context())
		log.Info().Msg("Поступил запрос на получение метрики...")

		_, err := parseMetricTypeV1(c)
		if err != nil {
			handleError(c, err)
			return
		}

		metricName, err := parseMetricNameV1(c)
		if err != nil {
			handleError(c, err)
			return
		}

		metric, err := metricService.GetMetric(c.Request.Context(), metricName)

		if err != nil {
			handleError(c, err)
			return
		}
		var result string
		if metric.MType == constants.GaugeMetricType {
			result = strconv.FormatFloat(*metric.Value, 'g', -1, 64)
		} else if metric.MType == constants.CounterMetricType {
			result = strconv.FormatInt(*metric.Delta, 10)
		}

		log.Info().Msg("Метрика успешно получена")
		c.String(http.StatusOK, result)
	}
}

// HandleMetricGetAll хэндлер для обработки запроса на получение всех метрик
func HandleMetricGetAll(metricService service.MetricService) gin.HandlerFunc {
	return func(c *gin.Context) {

		log := logging.FromContext(c.Request.Context())
		log.Info().Msg("Поступил запрос на получение всех метрик...")

		metrics := metricService.GetAll(c.Request.Context())

		c.HTML(http.StatusOK, "metrics.html", gin.H{
			"Metrics": metrics,
		})

		log.Info().Msg("Данные по метрикам успешно отправлены...")
	}
}

func HandleMetricUpdate(metricService service.MetricService) gin.HandlerFunc {
	return func(c *gin.Context) {

		log := logging.FromContext(c.Request.Context())
		log.Info().Msg("Поступил запрос на обработку метрики...")

		gaugeDto, err := createGaugeDtoV1(c)
		if err != nil {
			handleError(c, err)
			return
		}

		metricService.UpdateMetric(c.Request.Context(), gaugeDto)

		log.Info().Msg("Метрика успешно обработана")

		c.Status(http.StatusOK)
	}
}

func createGaugeDtoV1(c *gin.Context) (*dto.MetricDto, error) {

	metricType, err := parseMetricTypeV1(c)
	if err != nil {
		return nil, err
	}

	metricName, err := parseMetricNameV1(c)
	if err != nil {
		return nil, err
	}

	var result = dto.MetricDto{
		MType: metricType,
		ID:    metricName,
	}

	metricValueStr := c.Param(constants.MetricValuePathKey)

	if metricType == constants.GaugeMetricType {
		gauge, err := parseGauge(metricValueStr)
		if err != nil {
			return nil, err
		}
		result.Value = gauge
	}

	if metricType == constants.CounterMetricType {
		counter, err := parseCounter(metricValueStr)
		if err != nil {
			return nil, err
		}
		result.Delta = counter
	}

	return &result, nil
}

func parseMetricNameV1(c *gin.Context) (string, error) {
	metricName := c.Param(constants.MetricNamePathKey)
	err := validateMetricName(metricName)
	if err != nil {
		return "", err
	}
	return metricName, nil
}

func parseMetricTypeV1(c *gin.Context) (string, error) {
	metricType := c.Param(constants.MetricTypePathKey)
	if metricType != constants.GaugeMetricType && metricType != constants.CounterMetricType {
		return "", servererror.ErrNotAllowedMetricType
	}
	return metricType, nil
}

// Value
func parseGauge(metricValueStr string) (*float64, error) {
	value, err := strconv.ParseFloat(metricValueStr, 64)
	if err != nil {
		return nil, fmt.Errorf("%w %w", servererror.ErrNotValidMetricValue, err)
	}
	return &value, nil
}

// Delta
func parseCounter(metricValueStr string) (*int64, error) {
	value, err := strconv.ParseInt(metricValueStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("%w %w", servererror.ErrNotValidMetricValue, err)
	}
	return &value, nil
}

func validateMetricType(metricType string) error {
	if metricType != constants.GaugeMetricType && metricType != constants.CounterMetricType {
		return servererror.ErrNotAllowedMetricType
	}
	return nil
}

func validateMetricName(metricName string) error {
	if metricName == "" {
		return servererror.ErrNotValidMetricName
	}
	return nil
}
