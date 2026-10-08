package main

import (
	"flag"
	"log/slog"
	"strings"

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
			"host", ctx.Request.Host,
			"method", ctx.Request.Method,
			"path", ctx.Request.URL.Path,
		)
		ctx.Set(utils.LoggerKey, logger)
		ctx.Next()

		logger.Debug("Обработка метода завершена. Отдаем ответ")
	}
}

func run() {

	s := flag.String("a", "localhost:8080", "server url")
	slog.Info("АДРЕС В ТАКОМ ФОРМАТЕ!!!!!!!! ", slog.Any("server url", *s))
	cleaned := strings.TrimPrefix(*s, "http://")
	conf := server.Config{Level: "debug", ServerURL: cleaned}
	flag.Parse()

	// настраиваем дефолтный логгер
	utils.InitBaseLogger(utils.LoggerConfig{Level: conf.Level})
	gin.SetMode(gin.ReleaseMode)
	gin.Recovery()

	slog.Info("Инициализация http-server-а")
	// инициализируем http-server
	router := gin.New()
	router.Use(LoggingMiddleware())
	handler.SetUpRoutes(router)

	// Запуск сервера на порту
	slog.Info("Инициализация http-server-а выполнена успешно")
	err := router.Run(conf.ServerURL)
	if err != nil {
		slog.Error("Ошибка во время запуска http-сервера на порту ", slog.Any("err", err.Error()))
	}
}
