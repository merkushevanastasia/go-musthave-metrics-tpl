package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/model/dto"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/service"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/pkg/logs"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func HandleMetricGetV2(metricService service.MetricService) gin.HandlerFunc {
	return func(c *gin.Context) {

		log := logs.GetLoggerFromContext(c.Request.Context())

		metricRequest, err := parseMetricDtoV2(c)
		if err != nil {
			handleError(c, err)
			return
		}

		metric, err := metricService.GetMetric(c.Request.Context(), metricRequest.ID)

		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, &metric)

		log.Info().Msg("Метрика успешно получена")

	}
}

func HandleMetricUpdateV2(metricService service.MetricService) gin.HandlerFunc {
	return func(c *gin.Context) {
		log := logs.GetLoggerFromContext(c.Request.Context())

		log.Info().Msg("Поступил запрос на обработку метрики...")

		metricDto, err := parseMetricDtoV2(c)
		if err != nil {
			handleError(c, err)
			return
		}

		err = validateMetricName(metricDto.ID)
		if err != nil {
			handleError(c, err)
			return
		}
		err = validateMetricType(metricDto.MType)
		if err != nil {
			handleError(c, err)
			return
		}

		metricService.UpdateMetric(c.Request.Context(), &metricDto)

		log.Info().Msg("Метрика успешно обработана")

		c.Status(http.StatusOK)
	}
}
func HandleDBPing(dbPool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		log := logs.GetLoggerFromContext(c.Request.Context())

		log.Info().Msg("Поступил запрос на проверку соединения с БД...")

		if dbPool == nil {
			c.Status(http.StatusInternalServerError)
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
		defer cancel()

		if err := dbPool.Ping(ctx); err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}

		c.Status(http.StatusOK)

		log.Info().Msg("Запрос успешно обработан")

		c.Status(http.StatusOK)
	}
}

func parseMetricDtoV2(c *gin.Context) (dto.MetricDto, error) {
	var metricRequest dto.MetricDto
	if err := json.NewDecoder(c.Request.Body).Decode(&metricRequest); err != nil {
		return dto.MetricDto{}, err
	}
	return metricRequest, nil
}
