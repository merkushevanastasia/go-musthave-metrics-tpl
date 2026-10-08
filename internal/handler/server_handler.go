package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/constants"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/dto"
	servererror "github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/error"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/repository"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/service"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/utils"
	"github.com/bytedance/gopkg/util/logger"
	"github.com/gin-gonic/gin"
)

var MetricNamePathKey = "metric_name"
var MetricValuePathKey = "metric_value"
var MetricTypePathKey = "metric_type"

var repo = repository.MetricRepositoryImpl{}
var metricService = service.NewMetricService(repo)

func SetUpRoutes(router *gin.Engine) {
	pathUpdate := fmt.Sprintf("/update/:%s/:%s/:%s", MetricTypePathKey, MetricNamePathKey, MetricValuePathKey)
	pathGet := fmt.Sprintf("/value/:%s/:%s", MetricTypePathKey, MetricNamePathKey)
	router.POST(pathUpdate, HandleMetricUpdate)
	router.GET(pathGet, HandleMetricGet)
	router.GET("/", HandleMetricGetAll)
}

// HandleMetricUpdate хэндлер для обработки запроса на обновление значения метрики
func HandleMetricUpdate(c *gin.Context) {

	// Получаем логгер из контекста Gin (ключ приведен к string, так как Gin использует строки в качестве ключей)
	log := utils.FromContext(c)
	log.Info("Поступил запрос на обработку метрики...")

	// Парсим тип метрики

	gaugeDto, err := createGaugeDto(c)
	if err != nil {
		handleError(c, err)
		return
	}

	metricService.UpdateMetric(c, gaugeDto)

	log.Info("Метрика успешно обработана")

	// Возвращаем успешный статус ответа text/plain
	c.Status(http.StatusOK)
}

// HandleMetricGet хэндлер для обработки запроса на получение значения метрики
func HandleMetricGet(c *gin.Context) {

	log := utils.FromContext(c)
	log.Info("Поступил запрос на получение метрики...")

	// Парсим тип метрики
	metricType, err := parseMetricType(c)
	if err != nil {
		handleError(c, err)
		return
	}

	// Парсим имя метрики
	metricName, err := parseMetricName(c)
	if err != nil {
		handleError(c, err)
		return
	}

	metric, err := metricService.GetMetric(c, metricType, metricName)
	if err != nil {
		handleError(c, err)
		return
	}
	var result string
	if metric.MetricType == constants.GaugeMetricType {
		result = strconv.FormatFloat(metric.Gauge, 'g', -1, 64)
	} else if metric.MetricType == constants.CounterMetricType {
		result = strconv.FormatInt(metric.Counter, 10)
	}

	log.Info("Метрика успешно получена")
	c.String(http.StatusOK, result)

	c.Status(http.StatusOK)
}

// HandleMetricGetAll хэндлер для обработки запроса на получение всех метрик
func HandleMetricGetAll(c *gin.Context) {

	log := utils.FromContext(c)

	log.Info("Поступил запрос на получение всех метрик...")
	metrics, err := metricService.GetAll(c)
	if err != nil {
		handleError(c, err)
		return
	}

	// 2. Формируем HTML в виде обычной строки
	var html strings.Builder
	html.WriteString("<html><head><title>Metrics</title></head><body><h1>Current Metrics</h1><ul>")

	// Замените на ваш цикл по вашим метрикам
	for _, m := range metrics {
		if m.MetricType == constants.GaugeMetricType {
			html.WriteString(fmt.Sprintf("<li>%s: %v</li>", m.MetricName, m.Gauge))
		} else if m.MetricType == constants.CounterMetricType {
			html.WriteString(fmt.Sprintf("<li>%s: %v</li>", m.MetricName, m.Counter))
		}
	}

	html.WriteString("</ul></body></html>")
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html.String()))
	log.Info("Данные по метрикам успешно отправлены...")

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
func handleError(c *gin.Context, err error) {
	if err != nil {
		logger.Error(err.Error())

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
