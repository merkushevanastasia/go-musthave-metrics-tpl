package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/constants"
	servererror "github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/error"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/service"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/utils/logging"
	"github.com/gin-gonic/gin"
)

func SetUpRoutes(router *gin.Engine, metricService service.MetricService) {

	pathUpdate := fmt.Sprintf("/update/:%s/:%s/:%s", constants.MetricTypePathKey, constants.MetricNamePathKey, constants.MetricValuePathKey)
	pathGet := fmt.Sprintf("/value/:%s/:%s", constants.MetricTypePathKey, constants.MetricNamePathKey)

	router.POST(pathUpdate, HandleMetricUpdate(metricService))
	router.GET(pathGet, HandleMetricGetV1(metricService))
	router.GET("/", HandleMetricGetAll(metricService))
	router.POST("/value", HandleMetricGetV2(metricService))
	router.POST("/value/", HandleMetricGetV2(metricService))
	router.POST("/update", HandleMetricUpdateV2(metricService))
}

// handleError маппинг кастомной ошибки в ожидаемый ответ
func handleError(c *gin.Context, err error) {
	logger := logging.FromContext(c.Request.Context())

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
