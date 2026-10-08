package main

import (
	"flag"
	"log/slog"

	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/config/server"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/handler"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

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
	conf := server.Config{Level: "debug"}

	// получаем необходимые настройки из аргументов командной строки
	flag.StringVar(&conf.ServerURL, "a", "localhost:8080", "server url")
	flag.Parse()

	// настраиваем дефолтный логгер
	utils.InitBaseLogger(utils.LoggerConfig{Level: conf.Level})
	gin.SetMode(gin.ReleaseMode)

	slog.Info("Инициализация http-server-а")
	// инициализируем http-server
	router := gin.New()
	router.LoadHTMLGlob("templates/*")
	handler.SetUpRoutes(router)
	// Оборачиваем ВЕСЬ роутер в middleware для логирования каждого запроса и простановки MDC
	router.Use(LoggingMiddleware())

	// Запуск сервера на порту
	slog.Info("Инициализация http-server-а выполнена успешно")
	err := router.Run(conf.ServerURL)
	if err != nil {
		slog.Error("Ошибка во время запуска http-сервера на порту ", slog.Any("err", err.Error()))
	}
}
