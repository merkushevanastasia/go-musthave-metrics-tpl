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
)

var MetricNamePathKey = "metric_name"
var MetricValuePathKey = "metric_value"
var MetricTypePathKey = "metric_type"

// HandleMetricUpdate хэндлер для обработки запроса на обновление значения метрики
func HandleMetricUpdate(response http.ResponseWriter, request *http.Request) {

	repo := repository.MetricRepositoryImpl{}
	metricService := service.NewMetricService(repo)

	//Получаем логер с MDC
	logger := utils.FromContext(request.Context())
	logger.Info("Поступил запрос на обработку метрики...")

	// Разрешен только Post-запрос
	if request.Method != http.MethodPost {
		handleError(response, servererror.ErrNotAllowedMethod, logger)
		return
	}

	//Парсим тип метрики
	metricType, err := parseMetricType(request)
	if err != nil {
		handleError(response, err, logger)
		return
	}

	metricName, err := parseMetricName(request)
	if err != nil {
		handleError(response, fmt.Errorf("%w %w", servererror.ErrNotValidMetricName, err), logger)
		return
	}
	switch metricType {
	case dto.Gauge:
		gaugeDto, err := createGaugeDto(request, metricName)

		if err != nil {
			handleError(response, err, logger)
			return
		}
		metricService.ProcessGauge(request.Context(), gaugeDto)

	case dto.Counter:
		counterDto, err := createCounterDto(request, metricName)
		if err != nil {
			handleError(response, err, logger)
			return
		}
		metricService.ProcessCounter(request.Context(), counterDto)

	}

	logger.Info("Метрика успешно обработана")

}

// parseMetricType Функция конвертации: строка -> MetricType. Возвращает тип и ошибку, если строка неизвестна
func parseMetricType(request *http.Request) (dto.MetricType, error) {
	metricType := request.PathValue(MetricTypePathKey)
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
func createCounterDto(request *http.Request, metricName string) (dto.CounterMetricDto, error) {
	metricValueStr := request.PathValue(MetricValuePathKey)
	metricValue, err := strconv.ParseInt(metricValueStr, 10, 64)
	if err != nil {
		return dto.CounterMetricDto{}, fmt.Errorf("%w %w", servererror.ErrNotValidMetricValue, err)
	}
	return dto.CounterMetricDto{
		Name:  metricName,
		Value: metricValue,
	}, err
}

// createCounterDto создаем GaugeMetricDto
func createGaugeDto(request *http.Request, metricName string) (dto.GaugeMetricDto, error) {
	metricValueStr := request.PathValue(MetricValuePathKey)
	metricValue, err := strconv.ParseFloat(metricValueStr, 64)
	if err != nil {
		return dto.GaugeMetricDto{}, fmt.Errorf("%w %w", servererror.ErrNotValidMetricValue, err)
	}
	return dto.GaugeMetricDto{
		Name:  metricName,
		Value: metricValue,
	}, err
}

// parseMetricName парсим имя метрики
func parseMetricName(request *http.Request) (string, error) {
	metricName := request.PathValue(MetricNamePathKey)
	if metricName == "" {
		return "", servererror.ErrNotValidMetricName
	}
	return metricName, nil
}

// handleError маппинг кастомной ошибки в ожидаемые ответ
func handleError(response http.ResponseWriter, err error, logger *slog.Logger) {
	if err != nil {

		logger.Error(err.Error())
		if errors.Is(err, servererror.ErrNotValidMetricName) {
			http.Error(response, err.Error(), http.StatusNotFound)
			return
		}

		if errors.Is(err, servererror.ErrNotAllowedMetricType) {
			http.Error(response, err.Error(), http.StatusBadRequest)
			return
		}

		if errors.Is(err, servererror.ErrNotValidMetricValue) {
			http.Error(response, err.Error(), http.StatusBadRequest)
			return
		}

		http.Error(response, "", http.StatusInternalServerError)
		return
	}
}
