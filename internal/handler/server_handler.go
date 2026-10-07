package handler

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	dto "github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/dto/server"
	servererror "github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/error"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/repository"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/service"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/utils"
	"github.com/gin-gonic/gin"
)

var MetricNamePathKey = "metric_name"
var MetricValuePathKey = "metric_value"
var MetricTypePathKey = "metric_type"

// HandleMetricUpdate хэндлер для обработки запроса на обновление значения метрики
func HandleMetricUpdate(c *gin.Context) {
	repo := repository.MetricRepositoryImpl{}
	metricService := service.NewMetricService(repo)

	// Получаем логгер из контекста Gin (ключ приведен к string, так как Gin использует строки в качестве ключей)
	logger := utils.FromContext(c)

	logger.Info("Поступил запрос на обработку метрики...")

	// Парсим тип метрики
	metricType, err := parseMetricType(c)
	if err != nil {
		handleError(c, err, logger)
		return
	}

	// Парсим имя метрики
	metricName, err := parseMetricName(c)
	if err != nil {
		handleError(c, fmt.Errorf("%w %w", servererror.ErrNotValidMetricName, err), logger)
		return
	}

	// Получаем контекст запроса для передачи в сервисный слой
	ctx := c.Request.Context()

	switch metricType {
	case dto.Gauge:
		gaugeDto, err := createGaugeDto(c, metricName)
		if err != nil {
			handleError(c, err, logger)
			return
		}
		metricService.ProcessGauge(ctx, gaugeDto)

	case dto.Counter:
		counterDto, err := createCounterDto(c, metricName)
		if err != nil {
			handleError(c, err, logger)
			return
		}
		metricService.ProcessCounter(ctx, counterDto)
	}

	logger.Info("Метрика успешно обработана")

	// Возвращаем успешный статус ответа text/plain
	c.Status(http.StatusOK)
}

// HandleMetricGet хэндлер для обработки запроса на получение значения метрики
func HandleMetricGet(c *gin.Context) {
	repo := repository.MetricRepositoryImpl{}
	metricService := service.NewMetricService(repo)

	logger := utils.FromContext(c)

	logger.Info("Поступил запрос на обработку метрики...")

	// Парсим тип метрики
	metricType, err := parseMetricType(c)
	if err != nil {
		handleError(c, err, logger)
		return
	}

	// Парсим имя метрики
	metricName, err := parseMetricName(c)
	if err != nil {
		handleError(c, fmt.Errorf("%w %w", servererror.ErrNotValidMetricName, err), logger)
		return
	}

	// Получаем контекст запроса для передачи в сервисный слой
	ctx := c.Request.Context()

	var result string
	switch metricType {
	case dto.Gauge:
		value, err := metricService.GetGaugeValue(ctx, metricName)
		result = strconv.FormatFloat(value, 'g', -1, 64)
		if err != nil {
			handleError(c, err, logger)
			return
		}

	case dto.Counter:
		value, err := metricService.GetCounterValue(ctx, metricName)
		result = strconv.FormatInt(value, 10)

		if err != nil {
			handleError(c, err, logger)
			return
		}
	}

	logger.Info("Метрика успешно получена")
	c.String(http.StatusOK, result)

	c.Status(http.StatusOK)
}

// HandleMetricGetAll хэндлер для обработки запроса на получение всех метрик
func HandleMetricGetAll(c *gin.Context) {

	repo := repository.MetricRepositoryImpl{}
	metricService := service.NewMetricService(repo)

	logger := utils.FromContext(c)

	logger.Info("Поступил запрос на получение всех метрик...")
	counters, gauges := metricService.GetAll(c)

	html := createHtml(counters, gauges)

	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
	logger.Info("Данные по метрикам успешно отправлены...")

}

func createHtml(counters []dto.CounterMetricDto, gauges []dto.GaugeMetricDto) string {
	html := `<!DOCTYPE html>
<html>
<head>
    <title>Все метрики</title>
</head>
<body>
    <table>
        <tr>
            <th>Name</th>
            <th>Value</th>
        </tr>`

	// Добавляем Counter-метрики в таблицу
	for _, c := range counters {
		html += fmt.Sprintf("<tr><td>%s</td><td>%d</td></tr>", c.Name, c.Value)
	}
	// Добавляем Gauge-метрики в таблицу
	for _, g := range gauges {
		// Преобразуем float64 в строку без лишних нулей на конце (-1)
		valStr := strconv.FormatFloat(g.Value, 'f', -1, 64)
		html += fmt.Sprintf("<tr><td>%s</td><td>%s</td></tr>", g.Name, valStr)
	}
	html += `    </table>
</body>
</html>`
	return html
}

// parseMetricType Функция конвертации: строка -> MetricType. Возвращает тип и ошибку, если строка неизвестна
func parseMetricType(c *gin.Context) (dto.MetricType, error) {
	metricType := c.Param(MetricTypePathKey)
	switch metricType {
	case "gauge":
		return dto.Gauge, nil
	case "counter":
		return dto.Counter, nil
	default:
		return -1, servererror.ErrNotAllowedMetricType
	}
}

// createCounterDto создаем CounterMetricDto
func createCounterDto(c *gin.Context, metricName string) (dto.CounterMetricDto, error) {
	metricValueStr := c.Param(MetricValuePathKey)
	metricValue, err := strconv.ParseInt(metricValueStr, 10, 64)
	if err != nil {
		return dto.CounterMetricDto{}, fmt.Errorf("%w %w", servererror.ErrNotValidMetricValue, err)
	}
	return dto.CounterMetricDto{
		Name:  metricName,
		Value: metricValue,
	}, nil
}

// createGaugeDto создаем GaugeMetricDto
func createGaugeDto(c *gin.Context, metricName string) (dto.GaugeMetricDto, error) {
	metricValueStr := c.Param(MetricValuePathKey)
	metricValue, err := strconv.ParseFloat(metricValueStr, 64)
	if err != nil {
		return dto.GaugeMetricDto{}, fmt.Errorf("%w %w", servererror.ErrNotValidMetricValue, err)
	}
	return dto.GaugeMetricDto{
		Name:  metricName,
		Value: metricValue,
	}, nil
}

// parseMetricName парсим имя метрики
func parseMetricName(c *gin.Context) (string, error) {
	metricName := c.Param(MetricNamePathKey)
	if metricName == "" {
		return "", servererror.ErrNotValidMetricName
	}
	return metricName, nil
}

// handleError маппинг кастомной ошибки в ожидаемый ответ
func handleError(c *gin.Context, err error, logger *slog.Logger) {
	if err != nil {
		logger.Error(err.Error())

		if errors.Is(err, servererror.ErrNotValidMetricName) {
			c.String(http.StatusNotFound, err.Error())
			return
		}

		if errors.Is(err, servererror.ErrMetricNotFound) {
			c.String(http.StatusNotFound, err.Error())
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
