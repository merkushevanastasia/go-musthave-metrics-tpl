package main

import (
	"fmt"
	"log/slog"

	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/handler"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var serverPort = 8080
var level = "debug"

func main() {
	slog.Info("Запуск сервера для сбора рантайм-метрик...")
	run()
}

// LoggingMiddleware Здесь мы логируем начало и конец обработки любого запроса, а так же проставляем requestId в логгер
func LoggingMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		reqID := uuid.New()
		var logger = slog.Default().With("requestId", reqID)
		logger.Debug("Входящий HTTP запрос",
			"method", ctx.Request.Method,
			"path", ctx.Request.URL.Path,
		)
		ctx.Set(utils.LoggerKey, logger)

		ctx.Next()

		logger.Debug("Обработка метода завершена. Отдаем ответ")
	}
}

func run() {

	utils.InitBaseLogger(utils.LoggerConfig{Level: level})
	gin.SetMode(gin.ReleaseMode)

	slog.Info("Инициализация http-server-а...")
	pathUpdate := fmt.Sprintf("/update/:%s/:%s/:%s",
		handler.MetricTypePathKey,
		handler.MetricNamePathKey,
		handler.MetricValuePathKey,
	)
	pathGet := fmt.Sprintf("/value/:%s/:%s",
		handler.MetricTypePathKey,
		handler.MetricNamePathKey,
	)
	router := gin.New()
	router.POST(pathUpdate, handler.HandleMetricUpdate)
	router.GET(pathGet, handler.HandleMetricGet)
	router.GET("/", handler.HandleMetricGetAll)
	// Оборачиваем ВЕСЬ роутер в middleware для логирования каждого запроса и простановки MDC
	router.Use(LoggingMiddleware())
	slog.Info("Инициализация http-server-а выполнена успешно. Запуск...")
	// Запуск сервера на порту
	err := router.Run(fmt.Sprintf(":%d", serverPort))
	if err != nil {
		slog.Error("Ошибка во время запуска http-сервера на порту ", slog.Any("err", err.Error()))
	}
}
