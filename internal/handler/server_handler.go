package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/constants"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/dto"
	servererror "github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/error"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/service"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/utils/logging"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

var MetricNamePathKey = "metric_name"
var MetricValuePathKey = "metric_value"
var MetricTypePathKey = "metric_type"

func SetUpRoutes(router *gin.Engine, metricService service.MetricService) {
	pathUpdate := fmt.Sprintf("/update/:%s/:%s/:%s", MetricTypePathKey, MetricNamePathKey, MetricValuePathKey)
	pathUpdateV2 := "/update"
	pathGet := fmt.Sprintf("/value/:%s/:%s", MetricTypePathKey, MetricNamePathKey)
	pathGetV2 := "/value"
	router.POST(pathUpdate, HandleMetricUpdate(metricService))
	router.GET(pathGet, HandleMetricGet(metricService))
	router.POST(pathGetV2, HandleMetricGetV2(metricService))
	router.GET("/", HandleMetricGetAll(metricService))
	router.POST(pathUpdateV2, HandleMetricUpdateV2(metricService))
}

// HandleMetricUpdate хэндлер для обработки запроса на обновление значения метрики
func HandleMetricUpdate(metricService service.MetricService) gin.HandlerFunc {
	return func(c *gin.Context) {

		log := logging.FromContext(c.Request.Context())

		log.Info().Msg("Поступил запрос на обработку метрики...")

		gaugeDto, err := createGaugeDto(c)
		if err != nil {
			handleError(c, err, log)
			return
		}

		metricService.UpdateMetric(c.Request.Context(), gaugeDto)

		log.Info().Msg("Метрика успешно обработана")

		c.Status(http.StatusOK)
	}
}

func HandleMetricUpdateV2(metricService service.MetricService) gin.HandlerFunc {
	return func(c *gin.Context) {
		log := logging.FromContext(c.Request.Context())
		var metricRequest dto.MetricsRequest
		if err := json.NewDecoder(c.Request.Body).Decode(&metricRequest); err != nil {
			handleError(c, err, log)
			return
		}

		log.Info().Msg("Поступил запрос на обработку метрики...")

		gaugeDto, err := createGaugeDtoFromMetricRequest(metricRequest)
		if err != nil {
			handleError(c, err, log)
			return
		}

		metricService.UpdateMetric(c.Request.Context(), gaugeDto)

		log.Info().Msg("Метрика успешно обработана")

		c.Status(http.StatusOK)
	}
}

// HandleMetricGet хэндлер для обработки запроса на получение значения метрики
func HandleMetricGet(metricService service.MetricService) gin.HandlerFunc {
	return func(c *gin.Context) {

		log := logging.FromContext(c.Request.Context())
		log.Info().Msg("Поступил запрос на получение метрики...")

		metricType, err := parseMetricType(c)
		if err != nil {
			handleError(c, err, log)
			return
		}

		metricName, err := parseMetricName(c)
		if err != nil {
			handleError(c, err, log)
			return
		}

		metric, err := metricService.GetMetric(c.Request.Context(), metricType, metricName)

		if err != nil {
			handleError(c, err, log)
			return
		}
		var result string
		if metric.MetricType == constants.GaugeMetricType {
			result = strconv.FormatFloat(metric.Gauge, 'g', -1, 64)
		} else if metric.MetricType == constants.CounterMetricType {
			result = strconv.FormatInt(metric.Counter, 10)
		}

		log.Info().Msg("Метрика успешно получена")
		c.String(http.StatusOK, result)

		c.Status(http.StatusOK)
	}
}

func HandleMetricGetV2(metricService service.MetricService) gin.HandlerFunc {
	return func(c *gin.Context) {
		log := logging.FromContext(c.Request.Context())
		var metricRequest dto.MetricsRequest
		if err := json.NewDecoder(c.Request.Body).Decode(&metricRequest); err != nil {
			handleError(c, err, log)
			return
		}

		metric, err := metricService.GetMetric(c.Request.Context(), metricRequest.MType, metricRequest.ID)

		if err != nil {
			handleError(c, err, log)
			return
		}

		var delta *int64
		var value *float64

		if metric.MetricType == constants.GaugeMetricType {
			value = &metric.Gauge
		} else if metric.MetricType == constants.CounterMetricType {
			delta = &metric.Counter
		}

		response := dto.MetricsResponse{
			ID:    metric.MetricName,
			MType: metric.MetricType,
			Delta: delta,
			Value: value,
		}

		c.JSON(http.StatusOK, &response)

		log.Info().Msg("Метрика успешно получена")

	}
}

// HandleMetricGetAll хэндлер для обработки запроса на получение всех метрик
func HandleMetricGetAll(metricService service.MetricService) gin.HandlerFunc {
	return func(c *gin.Context) {

		log := logging.FromContext(c.Request.Context())

		log.Info().Msg("Поступил запрос на получение всех метрик...")
		metrics, err := metricService.GetAll(c.Request.Context())
		if err != nil {
			handleError(c, err, log)
			return
		}

		c.HTML(http.StatusOK, "metrics.html", gin.H{
			"Metrics": metrics,
		})
		log.Info().Msg("Данные по метрикам успешно отправлены...")

	}
}

// createGaugeDto создаем GaugeMetricDto
func createGaugeDto(c *gin.Context) (*dto.MetricDto, error) {
	metricType, err := parseMetricType(c)
	if err != nil {
		return nil, err
	}

	metricName, err := parseMetricName(c)
	if err != nil {
		return nil, err
	}

	metricValueStr := c.Param(MetricValuePathKey)

	var result = dto.MetricDto{
		MetricType: metricType,
		MetricName: metricName,
	}

	if metricType == constants.GaugeMetricType {
		value, err := strconv.ParseFloat(metricValueStr, 64)
		if err != nil {
			return nil, fmt.Errorf("%w %w", servererror.ErrNotValidMetricValue, err)
		}

		result.Gauge = value
	} else {
		metricValue, err := strconv.ParseInt(metricValueStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("%w %w", servererror.ErrNotValidMetricValue, err)
		}
		result.Counter = metricValue
	}

	return &result, nil
}

func createGaugeDtoFromMetricRequest(request dto.MetricsRequest) (*dto.MetricDto, error) {
	metricType := request.MType
	if metricType != constants.GaugeMetricType && metricType != constants.CounterMetricType {
		return nil, servererror.ErrNotAllowedMetricType
	}

	metricName := request.ID
	if metricName == "" {
		return nil, servererror.ErrNotValidMetricName
	}

	var gaugeVal float64
	if request.Value != nil {
		gaugeVal = *request.Value
	}

	var deltaVal int64
	if request.Delta != nil {
		deltaVal = *request.Delta
	}

	var result = dto.MetricDto{
		MetricType: metricType,
		MetricName: metricName,
		Gauge:      gaugeVal,
		Counter:    deltaVal,
	}

	return &result, nil
}

func parseMetricName(c *gin.Context) (string, error) {
	metricName := c.Param(MetricNamePathKey)
	if metricName == "" {
		return "", servererror.ErrNotValidMetricName
	}
	return metricName, nil
}

func parseMetricType(c *gin.Context) (string, error) {
	metricType := c.Param(MetricTypePathKey)
	if metricType != constants.GaugeMetricType && metricType != constants.CounterMetricType {
		return "", servererror.ErrNotAllowedMetricType
	}
	return metricType, nil
}

// handleError маппинг кастомной ошибки в ожидаемый ответ
func handleError(c *gin.Context, err error, logger *zerolog.Logger) {
	if err != nil {
		logger.Error().Msg(err.Error())

		if errors.Is(err, servererror.ErrNotValidMetricName) {
			c.String(http.StatusNotFound, err.Error())
			return
		}

		if errors.Is(err, servererror.ErrMetricNotFound) {
			c.Status(http.StatusNotFound)
			return
		}

		if errors.Is(err, servererror.ErrNotAllowedMetricType) {
			c.String(http.StatusBadRequest, err.Error())
			return
		}

		if errors.Is(err, servererror.ErrNotValidMetricValue) {
			c.String(http.StatusBadRequest, err.Error())
			return
		}

		c.Status(http.StatusInternalServerError)
		return
	}
}
